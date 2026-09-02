package middleware

import "testing"

// TestRedactSensitiveBody 锁定操作日志脱敏修复：命中敏感键名的值统一替换为 ***，
// 其余审计字段保留可读；覆盖登录、创建用户、修改密码三类明文口令路径。
func TestRedactSensitiveBody(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"login password", `{"username":"admin","password":"secret123"}`, `{"username":"admin","password":"***"}`},
		{"change password", `{"oldPassword":"old-pw","newPassword":"new-pw-6"}`, `{"oldPassword":"***","newPassword":"***"}`},
		{"create user password", `{"username":"eve","password":"p@ss","status":1}`, `{"username":"eve","password":"***","status":1}`},
		{"no secret field", `{"name":"foo","status":1}`, `{"name":"foo","status":1}`},
		{"empty body", ``, ``},
	}
	for _, tc := range cases {
		if got := redactSensitiveBody(tc.in); got != tc.want {
			t.Errorf("%s: redactSensitiveBody(%q) = %q, want %q", tc.name, tc.in, got, tc.want)
		}
	}
}
