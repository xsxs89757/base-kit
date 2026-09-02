package admin

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32" example:"newuser"`
	Password string `json:"password" validate:"required,min=6" example:"123456"`
	RealName string `json:"realName" validate:"required" example:"新用户"`
	Email    string `json:"email" validate:"omitempty,email" example:"user@example.com"`
	Phone    string `json:"phone" validate:"omitempty" example:"13800138000"`
	Status   int    `json:"status" validate:"oneof=0 1" example:"1"`
	Remark   string `json:"remark"`
	RoleIDs  []uint `json:"roleIds" example:"1,2"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	RealName string `json:"realName" example:"更新名称"`
	Password string `json:"password" validate:"omitempty,min=6"`
	Email    string `json:"email" validate:"omitempty,email" example:"user@example.com"`
	Phone    string `json:"phone" validate:"omitempty" example:"13800138000"`
	Status   int    `json:"status" validate:"oneof=0 1" example:"1"`
	Remark   string `json:"remark"`
	RoleIDs  []uint `json:"roleIds" example:"1,2"`
}

// UserItem 用户列表项
type UserItem struct {
	ID       uint     `json:"id" example:"1"`
	Username string   `json:"username" example:"vben"`
	RealName string   `json:"realName" example:"Vben"`
	Email    string   `json:"email" example:"vben@example.com"`
	Phone    string   `json:"phone" example:"13800138000"`
	Status   int      `json:"status" example:"1"`
	Roles    []string `json:"roles" example:"super"`
	RoleIDs  []uint   `json:"roleIds" example:"1"`
	Remark   string   `json:"remark"`
}
