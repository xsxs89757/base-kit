package admin

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 后台账号口令强度策略，两处使用：
//
//   - 设口令（建用户、管理员重置、本人改密）时强制，最小长度取 max(8, 配置 password_min_length)；
//   - 生产模式登录时按固定基线（8 位）再查一遍，存量弱口令账号哪怕密码对也登不进去，
//     必须由超管重置成合格口令（见 Authenticate）。登录不跟随配置：调高最小长度只约束新口令，
//     不能把已有合格口令的人一并锁在门外。
const (
	PasswordMinLength = 8  // 下限：配置 password_min_length 只能往上调
	PasswordMaxLength = 64 // 与 DTO 的 max 一致
	passwordMaxBytes  = 72 // bcrypt 只认前 72 字节，超出时 GenerateFromPassword 直接报错
)

// weakPasswords 是满足长度与字母数字组合、但出现在常见弱口令字典里的口令（小写比对）。
// 不满足组合规则的（123456、password 之类）不用列，结构检查就拦下了。
var weakPasswords = map[string]struct{}{}

func init() {
	for _, p := range []string{
		"password1", "password12", "password123", "passw0rd", "p@ssw0rd", "p@ssword1", "pass1234", "pass123456",
		"admin123", "admin1234", "admin12345", "admin123456", "admin888", "admin8888", "admin666", "admin6666",
		"admin@123", "admin#123", "admin@1234", "administrator1", "root1234", "root123456", "root@123", "super123", "super1234",
		"qwerty123", "qwerty1234", "qwe123456", "qwe12345", "qweasd123", "qwer1234", "1234qwer", "asdf1234", "zxcv1234",
		"1qaz2wsx", "1qaz@wsx", "1q2w3e4r", "1q2w3e4r5t", "q1w2e3r4", "q1w2e3r4t5", "a1b2c3d4", "zxcvbnm123", "123qwe123",
		"abc12345", "abc123456", "abcd1234", "abcd12345", "abc123abc", "123456abc", "a12345678", "a123456789", "a1234567",
		"12345678a", "123456789a", "aa123456", "aa12345678", "aaa123456", "test1234", "test123456", "welcome1",
		"iloveyou1", "woaini1314", "5201314a",
	} {
		weakPasswords[p] = struct{}{}
	}
}

// PasswordMinLen 返回当前设口令时要求的最小长度：配置 password_min_length 只能往上调，下限 8。
func PasswordMinLen() int {
	n := ConfigInt("password_min_length", PasswordMinLength)
	if n < PasswordMinLength {
		return PasswordMinLength
	}
	if n > PasswordMaxLength {
		return PasswordMaxLength
	}
	return n
}

// PasswordPolicyError 表示口令不满足强度策略，Error() 可直接展示给用户。
type PasswordPolicyError struct{ Reason string }

func (e *PasswordPolicyError) Error() string { return e.Reason }

// CheckPasswordStrength 校验要设置的新口令，nil 表示合格。
func CheckPasswordStrength(password, username string) error {
	return checkPassword(password, username, PasswordMinLen())
}

// passwordMeetsBaseline 按固定基线（不读配置）判断口令是否合格，供生产登录拦截弱口令用。
func passwordMeetsBaseline(password, username string) bool {
	return checkPassword(password, username, PasswordMinLength) == nil
}

func checkPassword(password, username string, minLen int) error {
	n := utf8.RuneCountInString(password)
	if n < minLen {
		return &PasswordPolicyError{fmt.Sprintf("密码长度不能少于%d位", minLen)}
	}
	if n > PasswordMaxLength || len(password) > passwordMaxBytes {
		return &PasswordPolicyError{fmt.Sprintf("密码长度不能超过%d位", PasswordMaxLength)}
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return &PasswordPolicyError{"密码必须同时包含字母和数字"}
	}
	lower := strings.ToLower(password)
	if _, weak := weakPasswords[lower]; weak {
		return &PasswordPolicyError{"密码过于常见，请换一个"}
	}
	if u := strings.ToLower(strings.TrimSpace(username)); utf8.RuneCountInString(u) >= 3 && strings.Contains(lower, u) {
		return &PasswordPolicyError{"密码不能包含用户名"}
	}
	return nil
}
