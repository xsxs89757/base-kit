package middleware

import (
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/xsxs89757/base-kit/config"
	adminmodel "github.com/xsxs89757/base-kit/model/admin"
	"github.com/xsxs89757/base-kit/store"

	"github.com/gofiber/fiber/v2"
)

// sensitiveBodyRe 匹配请求体 JSON 里的敏感字段值（密码/密钥等），用于写操作日志前脱敏。
// 之前只按 path 含 "login" 整体 [REDACTED]，漏掉了创建/编辑用户（password）与修改密码
// （oldPassword/newPassword）等路径，会把明文口令写进 operation_logs.body（该表可被持
// System:OperationLog:List 的管理员读取）。改为按字段脱敏：命中键名的值统一替换为 "***"，
// 其余审计字段保留可读。
var sensitiveBodyRe = regexp.MustCompile(`("(?:password|oldPassword|newPassword|confirmPassword|appSecret|app_secret|apiV3Key|api_v3_key|privateKey|private_key|secret)"\s*:\s*)"(?:[^"\\]|\\.)*"`)

func redactSensitiveBody(body string) string {
	if body == "" {
		return body
	}
	return sensitiveBodyRe.ReplaceAllString(body, `${1}"***"`)
}

var (
	opLogOnce sync.Once
	opLogCh   chan adminmodel.OperationLog
)

// startOpLogWriter 用单个后台 writer 串行写日志。
// 之前是每请求 `go DB.Create` 且丢弃错误：SQLite 上并发插库互相争锁
// （实测 100 并发写 INSERT 峰值 200ms+，还拖慢同期读请求），goroutine 数量也无上限。
// channel 满时丢弃该条并打日志，绝不阻塞请求。
func startOpLogWriter() {
	opLogCh = make(chan adminmodel.OperationLog, 1024)
	go func() {
		for entry := range opLogCh {
			if err := store.DB.Create(&entry).Error; err != nil {
				log.Printf("[operation-log] write failed: %v", err)
			}
		}
	}()
	go opLogCleanupLoop()
}

// opLogCleanupLoop 每天清理一次超过保留天数的日志；
// server.op_log_retention_days <= 0 时永久保留（默认行为，与历史一致）。
func opLogCleanupLoop() {
	days := config.C.Server.OpLogRetentionDays
	if days <= 0 {
		return
	}
	for {
		cutoff := time.Now().AddDate(0, 0, -days)
		res := store.DB.Where("created_at < ?", cutoff).Delete(&adminmodel.OperationLog{})
		if res.Error != nil {
			log.Printf("[operation-log] cleanup failed: %v", res.Error)
		} else if res.RowsAffected > 0 {
			log.Printf("[operation-log] cleaned %d entries older than %d days", res.RowsAffected, days)
		}
		time.Sleep(24 * time.Hour)
	}
}

func OperationLog() fiber.Handler {
	opLogOnce.Do(startOpLogWriter)
	return func(c *fiber.Ctx) error {
		if c.Method() == "GET" || c.Method() == "OPTIONS" {
			return c.Next()
		}

		path := c.Path()
		if strings.HasPrefix(path, "/swagger") {
			return c.Next()
		}

		start := time.Now()
		body := redactSensitiveBody(string(c.Body()))
		if len(body) > 2048 {
			body = body[:2048] + "...[truncated]"
		}

		err := c.Next()

		duration := time.Since(start).Milliseconds()

		// handler 返回 error 时 Fiber 的 ErrorHandler 尚未运行，
		// c.Response() 里还是默认 200，必须从 err 推导真实状态码，
		// 否则所有失败操作都会被记成"成功"（实测 400 全记成 200）
		status := c.Response().StatusCode()
		if err != nil {
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		userID, _ := c.Locals("userId").(uint)
		username, _ := c.Locals("username").(string)
		if username == "" {
			username = "-"
		}

		entry := adminmodel.OperationLog{
			UserID:    userID,
			Username:  username,
			Method:    c.Method(),
			Path:      path,
			Status:    status,
			Duration:  duration,
			IP:        c.IP(),
			UserAgent: c.Get("User-Agent"),
			Body:      body,
		}

		select {
		case opLogCh <- entry:
		default:
			log.Printf("[operation-log] buffer full, entry dropped: %s %s", entry.Method, entry.Path)
		}

		return err
	}
}
