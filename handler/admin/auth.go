package admin

import (
	"errors"
	"time"

	"github.com/xsxs89757/base-kit/config"
	"github.com/xsxs89757/base-kit/dto"
	admindto "github.com/xsxs89757/base-kit/dto/admin"
	"github.com/xsxs89757/base-kit/middleware"
	adminmodel "github.com/xsxs89757/base-kit/model/admin"
	adminsvc "github.com/xsxs89757/base-kit/service/admin"
	"github.com/xsxs89757/base-kit/store"
	"github.com/xsxs89757/base-kit/validator"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名和密码登录，返回 accessToken
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body admindto.LoginRequest true "登录参数"
// @Success 200 {object} dto.Response{data=admindto.LoginResponse}
// @Failure 400 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Router /admin/auth/login [post]
func Login(c *fiber.Ctx) error {
	var req admindto.LoginRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := adminsvc.Authenticate(req.Username, req.Password)
	if err != nil {
		return dto.Fail(c, fiber.StatusForbidden, "Username or password is incorrect.")
	}

	roles := adminsvc.GetRoleNames(user)
	accessToken, err := middleware.GenerateAccessToken(user.ID, user.Username, roles)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to generate token")
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to generate refresh token")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    refreshToken,
		HTTPOnly: true,
		SameSite: "None",
		Secure:   true,
		MaxAge:   int(config.C.JWT.RefreshExpire / time.Second),
	})

	return dto.Success(c, fiber.Map{
		"id":          user.ID,
		"username":    user.Username,
		"realName":    user.RealName,
		"avatar":      user.Avatar,
		"roles":       roles,
		"homePath":    resolveAccessibleHomePath(user),
		"accessToken": accessToken,
	})
}

// Logout 退出登录
// @Summary 退出登录
// @Description 清除 refresh token cookie
// @Tags 认证
// @Produce json
// @Success 200 {object} dto.Response
// @Router /admin/auth/logout [post]
func Logout(c *fiber.Ctx) error {
	clearRefreshCookie(c)
	return dto.Success(c, "")
}

// clearRefreshCookie 用与 Login 完全相同的属性覆盖删除。跨站部署时浏览器只接受
// SameSite=None; Secure 的 Set-Cookie，缺了属性的删除指令会被直接丢弃，旧 cookie 一直留着。
func clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		HTTPOnly: true,
		SameSite: "None",
		Secure:   true,
		MaxAge:   -1,
	})
}

// RefreshToken 刷新令牌
// @Summary 刷新 Access Token
// @Description 使用 cookie 中的 refresh token 获取新的 access token
// @Tags 认证
// @Produce plain
// @Success 200 {string} string "新的 access token"
// @Failure 403 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /admin/auth/refresh [post]
func RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("jwt")
	if refreshToken == "" {
		return dto.Fail(c, fiber.StatusForbidden, "Forbidden Exception")
	}

	// 只认 refresh 类型：access token 塞进 cookie 不能用来续签
	claims, err := middleware.ParseToken(refreshToken, middleware.TokenTypeRefresh)
	if err != nil {
		clearRefreshCookie(c)
		return dto.Fail(c, fiber.StatusForbidden, "Forbidden Exception")
	}

	user, err := adminsvc.GetUserByID(claims.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		clearRefreshCookie(c)
		return dto.Fail(c, fiber.StatusForbidden, "Forbidden Exception")
	}
	if err != nil {
		// 数据库抖动不是鉴权失败：不动 cookie，让前端稍后重试
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to load user")
	}

	// 禁用（status=0）或已删除的管理员不得再用 refresh cookie 续签 access token，
	// 否则后台"禁用账号"在 access token 过期前（最长可达数天）形同虚设。
	// 改密之前签发的 refresh token 同样作废，否则"改密即下线"会被续签绕过。
	if user.Status != 1 || middleware.TokenRevokedByPasswordChange(claims.IssuedAt, user.PasswordChangedAt) {
		clearRefreshCookie(c)
		return dto.Fail(c, fiber.StatusForbidden, "Forbidden Exception")
	}

	roles := adminsvc.GetRoleNames(user)
	newToken, err := middleware.GenerateAccessToken(user.ID, user.Username, roles)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "Failed to generate token")
	}

	return c.SendString(newToken)
}

// ChangePassword 修改当前用户密码
// @Summary 修改密码
// @Description 用户修改自己的密码，需验证旧密码；成功后此前签发的 token 全部失效，需重新登录
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body admindto.ChangePasswordRequest true "密码信息"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Router /admin/auth/change-password [post]
func ChangePassword(c *fiber.Ctx) error {
	var req admindto.ChangePasswordRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	userID := c.Locals("userId").(uint)
	user, err := adminsvc.GetUserByID(userID)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "用户不存在")
	}

	if err := adminsvc.VerifyPassword(user, req.OldPassword); err != nil {
		return dto.Fail(c, fiber.StatusForbidden, "旧密码不正确")
	}

	if err := adminsvc.ChangePassword(userID, req.NewPassword); err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "修改密码失败")
	}
	middleware.InvalidateUserAuthCache(userID)

	return dto.Success(c, nil)
}

// GetAccessCodes 获取权限码
// @Summary 获取当前用户权限码
// @Description 返回当前登录用户拥有的所有权限码列表
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]string}
// @Failure 401 {object} dto.Response
// @Router /admin/auth/codes [get]
func GetAccessCodes(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	user, err := adminsvc.GetUserByUsername(username)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "User not found")
	}
	codes := adminsvc.GetAccessCodes(user)
	if codes == nil {
		codes = []string{}
	}
	return dto.Success(c, codes)
}

// GetUserInfo 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 返回当前登录用户的基本信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=admindto.UserInfoResponse}
// @Failure 401 {object} dto.Response
// @Router /admin/user/info [get]
func GetUserInfo(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uint)
	user, err := adminsvc.GetUserByID(userID)
	if err != nil {
		return dto.Fail(c, fiber.StatusInternalServerError, "User not found")
	}
	roles := adminsvc.GetRoleNames(user)
	return dto.Success(c, fiber.Map{
		"id":       user.ID,
		"username": user.Username,
		"realName": user.RealName,
		"avatar":   user.Avatar,
		"roles":    roles,
		"homePath": resolveAccessibleHomePath(user),
	})
}

func resolveAccessibleHomePath(user *adminmodel.User) string {
	if adminsvc.IsSuperUser(user) {
		return user.HomePath
	}

	roleIDs := adminsvc.ActiveRoleIDs(user)
	if len(roleIDs) == 0 {
		return user.HomePath
	}

	if user.HomePath != "" {
		var count int64
		store.DB.Model(&adminmodel.Menu{}).
			Joins("JOIN role_menus ON role_menus.menu_id = sys_menus.id").
			Joins("JOIN sys_roles ON sys_roles.id = role_menus.role_id").
			Where("role_menus.role_id IN ? AND sys_roles.status = ?", roleIDs, 1).
			Where("sys_menus.path = ? AND sys_menus.type IN ? AND sys_menus.status = ?", user.HomePath, []string{"menu", "embedded", "link"}, 1).
			Count(&count)
		if count > 0 {
			return user.HomePath
		}
	}

	var menu adminmodel.Menu
	if err := store.DB.
		Joins("JOIN role_menus ON role_menus.menu_id = sys_menus.id").
		Joins("JOIN sys_roles ON sys_roles.id = role_menus.role_id").
		Where("role_menus.role_id IN ? AND sys_roles.status = ?", roleIDs, 1).
		Where("sys_menus.path <> '' AND sys_menus.type IN ? AND sys_menus.status = ?", []string{"menu", "embedded", "link"}, 1).
		Order("sys_menus.order_no ASC, sys_menus.id ASC").
		First(&menu).Error; err == nil {
		return menu.Path
	}

	return user.HomePath
}
