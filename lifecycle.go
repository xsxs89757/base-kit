package basekit

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// shutdownWait 是 app 关闭时等后台任务返回的上限。超时只打日志、不再等：
// 一个不看 ctx 的任务不该让 Shutdown 永远卡住。
var shutdownWait = 10 * time.Second

// lifecycle 管一个 app 的后台任务：ctx 在 app 关闭时取消，wg 记着还没返回的任务。
type lifecycle struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu      sync.Mutex
	stopped bool
	wg      sync.WaitGroup
}

var (
	lifecyclesMu sync.Mutex
	lifecycles   = map[*fiber.App]*lifecycle{}
)

// Go 起一个随 app 关闭而停止的后台任务（定时扫描、投递 worker 之类），在 Routes 里代替
// go fn(context.Background())。
//
// app.Shutdown()（以及 ShutdownWithTimeout / ShutdownWithContext）时 ctx 取消，
// 等 fn 返回后 Shutdown 才返回，最多等 10 秒，所以 fn 要在 ctx 取消后尽快返回。
//
// 用 context.Background() 起的任务永远不会停。store.DB 是包级变量，每次 NewApp 都换成新库：
// 测试里一个进程先后 NewApp 好几次，前面 app 的任务不会退出，而是转去读写新 app 的库，
// 同一个 worker 就同时跑了好几份。
func Go(app *fiber.App, fn func(ctx context.Context)) {
	lc := lifecycleOf(app)
	lc.mu.Lock()
	if lc.stopped {
		// app 正在关闭：这时再 Add 会和 stop 里的 Wait 冲突，任务也没有意义
		lc.mu.Unlock()
		return
	}
	lc.wg.Add(1)
	lc.mu.Unlock()

	go func() {
		defer lc.wg.Done()
		fn(lc.ctx)
	}()
}

// AppContext 返回随 app 关闭而取消的 ctx，给自己管 goroutine 的组件用。
// 自己起的循环用 Go：关闭时不光取消，还会等它返回。
func AppContext(app *fiber.App) context.Context {
	return lifecycleOf(app).ctx
}

// lifecycleOf 取 app 的 lifecycle，第一次取时创建并挂上关闭钩子。
func lifecycleOf(app *fiber.App) *lifecycle {
	lifecyclesMu.Lock()
	lc, ok := lifecycles[app]
	if !ok {
		ctx, cancel := context.WithCancel(context.Background())
		lc = &lifecycle{ctx: ctx, cancel: cancel}
		lifecycles[app] = lc
	}
	lifecyclesMu.Unlock()

	if !ok {
		// 没 Listen 过的 app（测试里只用 app.Test）调 Shutdown 也会执行关闭钩子
		app.Hooks().OnShutdown(func() error {
			lc.stop()
			lifecyclesMu.Lock()
			if lifecycles[app] == lc {
				delete(lifecycles, app)
			}
			lifecyclesMu.Unlock()
			return nil
		})
	}
	return lc
}

func (lc *lifecycle) stop() {
	lc.mu.Lock()
	lc.stopped = true
	lc.mu.Unlock()
	lc.cancel()

	done := make(chan struct{})
	go func() {
		lc.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(shutdownWait):
		log.Printf("basekit: app 关闭时后台任务 %s 内没有返回，不再等待（任务要在 ctx 取消后尽快返回）", shutdownWait)
	}
}
