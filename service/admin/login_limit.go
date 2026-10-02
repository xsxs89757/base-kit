package admin

import (
	"strings"
	"sync"
	"time"
)

// 后台登录失败闸：在窗口期内按三个维度计失败次数，任一维度超限就锁定一段时间。
//
//   - 用户名+IP：挡单点猜密码；
//   - IP：挡同一来源轮换用户名；
//   - 用户名：挡换着 IP 打同一个账号（实际爆破常见几十个云主机 IP 轮流猜同一个账号）。
//     阈值放宽，避免别人随手几次失败就把真管理员锁在门外。
//
// 进程内存实现：后台单实例部署，重启清零可以接受，目的只是把在线爆破拖慢到不可行。
const (
	loginFailWindow  = 15 * time.Minute
	loginLockFor     = 15 * time.Minute
	maxFailsUserIP   = 5
	maxFailsIP       = 20
	maxFailsUsername = 30
	loginLimitMaxKey = 100000 // 键数量上限，防被海量随机用户名撑爆内存
)

type loginFailEntry struct {
	count       int
	windowStart time.Time
	lockedUntil time.Time
}

type loginLimiter struct {
	mu      sync.Mutex
	entries map[string]*loginFailEntry
	now     func() time.Time
}

var adminLoginLimiter = newLoginLimiter(time.Now)

func newLoginLimiter(now func() time.Time) *loginLimiter {
	return &loginLimiter{entries: map[string]*loginFailEntry{}, now: now}
}

func loginLimitKeys(username, ip string) []struct {
	key string
	max int
} {
	u := strings.ToLower(strings.TrimSpace(username))
	return []struct {
		key string
		max int
	}{
		{"ui:" + u + "|" + ip, maxFailsUserIP},
		{"ip:" + ip, maxFailsIP},
		{"u:" + u, maxFailsUsername},
	}
}

// locked 返回仍需等待的时长；0 表示放行。
func (l *loginLimiter) locked(username, ip string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	var wait time.Duration
	for _, k := range loginLimitKeys(username, ip) {
		if e := l.entries[k.key]; e != nil && now.Before(e.lockedUntil) {
			if d := e.lockedUntil.Sub(now); d > wait {
				wait = d
			}
		}
	}
	return wait
}

func (l *loginLimiter) fail(username, ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if len(l.entries) >= loginLimitMaxKey {
		l.sweepLocked(now)
	}
	for _, k := range loginLimitKeys(username, ip) {
		e := l.entries[k.key]
		if e == nil || now.Sub(e.windowStart) > loginFailWindow {
			e = &loginFailEntry{windowStart: now, lockedUntil: e.lockedUntilOrZero()}
			l.entries[k.key] = e
		}
		e.count++
		if e.count >= k.max {
			e.lockedUntil = now.Add(loginLockFor)
			e.count = 0
			e.windowStart = now
		}
	}
}

// succeed 只清用户名+IP 维度：IP 与用户名维度可能正被别人爆破，不能因一次成功登录归零。
func (l *loginLimiter) succeed(username, ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, loginLimitKeys(username, ip)[0].key)
}

func (l *loginLimiter) sweepLocked(now time.Time) {
	for k, e := range l.entries {
		if now.Sub(e.windowStart) > loginFailWindow && !now.Before(e.lockedUntil) {
			delete(l.entries, k)
		}
	}
}

func (e *loginFailEntry) lockedUntilOrZero() time.Time {
	if e == nil {
		return time.Time{}
	}
	return e.lockedUntil
}

// LoginLockedFor 登录前调用：返回该用户名/来源还需等待多久（0 = 放行）。
func LoginLockedFor(username, ip string) time.Duration {
	return adminLoginLimiter.locked(username, ip)
}

// RecordLoginFailure 登录失败后调用。
func RecordLoginFailure(username, ip string) { adminLoginLimiter.fail(username, ip) }

// RecordLoginSuccess 登录成功后调用。
func RecordLoginSuccess(username, ip string) { adminLoginLimiter.succeed(username, ip) }
