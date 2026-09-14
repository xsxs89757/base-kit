package middleware

import (
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// kit 把 JWTAuth / PermissionAuth 挂在 /admin 前缀上，下游分组再挂一遍时同一请求会经过两次。
// 下面的用例在两次之间换掉判定依据：第二次如果真的重新判定，结果就会变。

func TestJWTAuthRunsOncePerRequest(t *testing.T) {
	setupJWTTest(t)
	user := createTestUser(t, "erin", 1)
	token, _ := GenerateAccessToken(user.ID, user.Username, nil)

	app := fiber.New()
	app.Use(JWTAuth())
	// 身份第一次已经装配好；换成无效 token，第二次若重新解析就会 401
	app.Use(func(c *fiber.Ctx) error {
		c.Request().Header.Set(fiber.HeaderAuthorization, "Bearer invalid")
		return c.Next()
	})
	app.Use(JWTAuth())
	app.Get("/whoami", func(c *fiber.Ctx) error {
		return c.SendString(c.Locals("username").(string))
	})

	resp := bearerRequest(t, app, http.MethodGet, "/whoami", token)
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(body) != "erin" {
		t.Fatalf("second JWTAuth pass: expected 200 erin, got %d %s", resp.StatusCode, body)
	}
}

func TestPermissionAuthRunsOncePerRoute(t *testing.T) {
	setupPermissionTestDB(t)
	snapshotRouteTables(t)
	grantMenuCode(t, "clerk", "Shop:Order:List")
	RegisterRoutePermissions(RoutePermission{Method: "GET", Path: "/admin/shop/order/list", Code: "Shop:Order:List"})

	appWithRoles := func(roles []string) *fiber.App {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("roles", roles)
			return c.Next()
		})
		app.Use(PermissionAuth())
		// 清空角色，第二次若重新判定就会 403
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("roles", []string{})
			return c.Next()
		})
		app.Use(PermissionAuth())
		app.Get("/admin/shop/order/list", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
		return app
	}

	if got := doRequest(t, appWithRoles([]string{"clerk"}), http.MethodGet, "/admin/shop/order/list"); got != http.StatusOK {
		t.Fatalf("granted on first pass: expected 200, got %d", got)
	}
	if got := doRequest(t, appWithRoles([]string{"viewer"}), http.MethodGet, "/admin/shop/order/list"); got != http.StatusForbidden {
		t.Fatalf("denied on first pass: expected 403, got %d", got)
	}
}

// 放行记录要带上 method+path：内部重定向（c.Path 改写后 RestartRouting）到别的路由必须重新判定。
// 两个路径等长是故意的：c.Path() 返回的字符串指向 Fiber 会原地改写的缓冲区，直接存它，
// 改写后读出来就是新路径，第二次判定会被跳过。
func TestPermissionAuthRechecksAfterInternalRedirect(t *testing.T) {
	setupPermissionTestDB(t)
	snapshotRouteTables(t)
	grantMenuCode(t, "clerk", "Shop:Order:List")
	RegisterRoutePermissions(
		RoutePermission{Method: "GET", Path: "/admin/shop/order/list", Code: "Shop:Order:List"},
		RoutePermission{Method: "GET", Path: "/admin/shop/order/stat", Code: "Shop:Order:Stat"},
	)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("roles", []string{"clerk"})
		return c.Next()
	})
	app.Use(PermissionAuth())
	app.Get("/admin/shop/order/list", func(c *fiber.Ctx) error {
		c.Path("/admin/shop/order/stat")
		return c.RestartRouting()
	})
	app.Get("/admin/shop/order/stat", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	if got := doRequest(t, app, http.MethodGet, "/admin/shop/order/list"); got != http.StatusForbidden {
		t.Fatalf("internal redirect to a route the role lacks: expected 403, got %d", got)
	}
}
