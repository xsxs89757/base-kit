package admin

import (
	"testing"
	"time"
)

func TestLoginLimiterLocksUserIPAfterRepeatedFailures(t *testing.T) {
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	l := newLoginLimiter(func() time.Time { return now })

	for i := 0; i < maxFailsUserIP-1; i++ {
		l.fail("admin", "1.1.1.1")
	}
	if l.locked("admin", "1.1.1.1") != 0 {
		t.Fatal("locked before reaching the threshold")
	}
	l.fail("admin", "1.1.1.1")
	if l.locked("admin", "1.1.1.1") == 0 {
		t.Fatal("expected lock after reaching the user+ip threshold")
	}
	if l.locked("admin", "2.2.2.2") != 0 {
		t.Fatal("another source should not be locked by user+ip failures")
	}

	now = now.Add(loginLockFor + time.Second)
	if l.locked("admin", "1.1.1.1") != 0 {
		t.Fatal("lock should expire")
	}
}

func TestLoginLimiterLocksUsernameAcrossRotatingIPs(t *testing.T) {
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	l := newLoginLimiter(func() time.Time { return now })

	// 攻击者轮换来源：每个 IP 只失败一次，按用户名维度仍应被锁
	for i := 0; i < maxFailsUsername; i++ {
		l.fail("Admin", "10.0.0."+string(rune('a'+i%26))+string(rune('a'+i/26)))
	}
	if l.locked("admin", "9.9.9.9") == 0 {
		t.Fatal("expected username lock across rotating IPs (case-insensitive)")
	}
}

func TestLoginLimiterSuccessOnlyClearsUserIP(t *testing.T) {
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	l := newLoginLimiter(func() time.Time { return now })

	for i := 0; i < maxFailsIP; i++ {
		l.fail("user"+string(rune('a'+i)), "3.3.3.3")
	}
	l.succeed("super", "3.3.3.3")
	if l.locked("someone", "3.3.3.3") == 0 {
		t.Fatal("a successful login must not clear the per-IP lock")
	}
}

func TestLoginLimiterWindowResets(t *testing.T) {
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	l := newLoginLimiter(func() time.Time { return now })

	for i := 0; i < maxFailsUserIP-1; i++ {
		l.fail("admin", "4.4.4.4")
	}
	now = now.Add(loginFailWindow + time.Second)
	l.fail("admin", "4.4.4.4")
	if l.locked("admin", "4.4.4.4") != 0 {
		t.Fatal("failures outside the window should not accumulate")
	}
}
