package router

import (
	admin "github.com/xsxs89757/base-kit/handler/admin"
	"github.com/xsxs89757/base-kit/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupAdmin registers all /admin/* routes (backend management).
func SetupAdmin(app *fiber.App) {
	g := app.Group("/admin")

	// Public auth routes
	auth := g.Group("/auth")
	auth.Post("/login", admin.Login)
	auth.Post("/logout", admin.Logout)
	auth.Post("/refresh", admin.RefreshToken)

	// Protected routes: JWTAuth 装配用户/角色，PermissionAuth 按 middleware/permission.go 的路由表校验权限码
	protected := g.Group("", middleware.JWTAuth(), middleware.PermissionAuth())

	// Operation log middleware (records POST/PUT/DELETE)
	protected.Use(middleware.OperationLog())

	// Auth info
	protected.Get("/auth/codes", admin.GetAccessCodes)
	protected.Post("/auth/change-password", admin.ChangePassword)

	// User info
	protected.Get("/user/info", admin.GetUserInfo)

	// Menu routes (for frontend routing)
	protected.Get("/menu/all", admin.GetAllMenus)

	// System management
	sys := protected.Group("/system")

	// Roles
	sys.Get("/role/all", admin.GetAllRoles)
	sys.Get("/role/list", admin.GetRoleList)
	// 角色管理专用菜单树：不依赖"菜单管理"权限，便于在没有菜单管理权限时仍能完成角色授权
	sys.Get("/role/menu-tree", admin.GetRoleMenuTree)
	sys.Post("/role", admin.CreateRole)
	sys.Put("/role/:id", admin.UpdateRole)
	sys.Delete("/role/:id", admin.DeleteRole)

	// Menus management
	sys.Get("/menu/list", admin.GetMenuList)
	sys.Get("/menu/name-exists", admin.CheckMenuNameExists)
	sys.Get("/menu/path-exists", admin.CheckMenuPathExists)
	sys.Post("/menu", admin.CreateMenu)
	sys.Put("/menu/:id", admin.UpdateMenu)
	sys.Delete("/menu/:id", admin.DeleteMenu)

	// Departments
	sys.Get("/dept/list", admin.GetDeptList)
	sys.Post("/dept", admin.CreateDept)
	sys.Put("/dept/:id", admin.UpdateDept)
	sys.Delete("/dept/:id", admin.DeleteDept)

	// Users management
	sys.Get("/user/list", admin.GetUserList)
	sys.Post("/user", admin.CreateUser)
	sys.Put("/user/:id", admin.UpdateUser)
	sys.Delete("/user/:id", admin.DeleteUser)

	// Configs management
	sys.Get("/config/list", admin.GetConfigList)
	sys.Get("/config/groups", admin.GetConfigGroups)
	sys.Post("/config", admin.CreateConfig)
	sys.Put("/config/:id", admin.UpdateConfig)
	sys.Delete("/config/:id", admin.DeleteConfig)

	// Operation logs
	sys.Get("/operation-log/list", admin.GetOperationLogList)
	sys.Delete("/operation-log/clear", admin.ClearOperationLog)
	sys.Delete("/operation-log/:id", admin.DeleteOperationLog)
}
