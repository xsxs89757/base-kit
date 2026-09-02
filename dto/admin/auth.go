package admin

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required" example:"vben"`
	Password string `json:"password" validate:"required" example:"123456"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" validate:"required" example:"123456"`
	NewPassword string `json:"newPassword" validate:"required,min=6" example:"654321"`
}

// LoginResponse 登录响应数据
type LoginResponse struct {
	ID          uint     `json:"id" example:"1"`
	Username    string   `json:"username" example:"vben"`
	RealName    string   `json:"realName" example:"Vben"`
	Avatar      string   `json:"avatar"`
	Roles       []string `json:"roles" example:"super"`
	HomePath    string   `json:"homePath"`
	AccessToken string   `json:"accessToken" example:"eyJhbGciOi..."`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	ID       uint     `json:"id" example:"1"`
	Username string   `json:"username" example:"vben"`
	RealName string   `json:"realName" example:"Vben"`
	Avatar   string   `json:"avatar"`
	Roles    []string `json:"roles" example:"super"`
	HomePath string   `json:"homePath"`
}
