package admin

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required" example:"vben"`
	Password string `json:"password" validate:"required" example:"123456"`
	// 系统配置 login_captcha 开启时必填：拼图验证通过后 /admin/auth/captcha/verify 返回的一次性凭证
	CaptchaToken string `json:"captchaToken" example:"Q4ZK..."`
}

// CaptchaResponse 登录拼图验证码。坐标、尺寸都按背景图原始像素；未开启时只有 enabled=false
type CaptchaResponse struct {
	Enabled   bool   `json:"enabled" example:"true"`
	CaptchaID string `json:"captchaId,omitempty" example:"K7Q2..."`
	// 带缺口的背景图（JPEG data URI）
	Image string `json:"image,omitempty" example:"data:image/jpeg;base64,/9j/4AAQ..."`
	// 拼图块（PNG data URI，透明底，边长 pieceSize）
	Piece string `json:"piece,omitempty" example:"data:image/png;base64,iVBORw0KGgo..."`
	// 拼图块顶边的纵坐标；横向从 0 开始拖
	PieceY    int `json:"pieceY,omitempty" example:"40"`
	PieceSize int `json:"pieceSize,omitempty" example:"66"`
	Width     int `json:"width,omitempty" example:"320"`
	Height    int `json:"height,omitempty" example:"160"`
}

// CaptchaVerifyRequest 提交拼图位置
type CaptchaVerifyRequest struct {
	CaptchaID string `json:"captchaId" validate:"required" example:"K7Q2..."`
	// 拼图块左边缘拖到的横坐标（背景图原始像素，前端按显示缩放比换算）
	X float64 `json:"x" validate:"gte=0" example:"152.5"`
}

// CaptchaVerifyResponse 拼图验证通过的一次性登录凭证（2 分钟有效）
type CaptchaVerifyResponse struct {
	Token string `json:"token" example:"Q4ZK..."`
}

// ChangePasswordRequest 修改密码请求
// 新密码另按口令强度策略校验：至少 8 位（系统配置 password_min_length 可调高）、同时含字母和数字、不是常见弱口令、不含用户名
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" validate:"required" example:"123456"`
	NewPassword string `json:"newPassword" validate:"required,min=8,max=64" example:"Np7xQ2wLs9"`
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
