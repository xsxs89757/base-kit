package admin

// CreateRoleRequest 创建/更新角色请求
//
// 注意：MenuIDs / Permissions 使用指针切片，目的是区分三种情况：
//  1. 字段未提交（指针 == nil）：保持原有菜单关联不变，常见于"仅修改状态/备注"等场景。
//  2. 显式提交空数组 []：清空所有菜单关联。
//  3. 提交非空数组：替换为指定的菜单集合。
//
// 这样可以避免前端在状态切换或表单初始化异常时把整个 permissions 误传成 [] 而清空角色权限。
type CreateRoleRequest struct {
	Name        string  `json:"name" validate:"required" example:"新角色"`
	Code        string  `json:"code" validate:"required" example:"new_role"`
	Status      int     `json:"status" validate:"oneof=0 1" example:"1"`
	Remark      string  `json:"remark"`
	MenuIDs     *[]uint `json:"menuIds"`
	Permissions *[]uint `json:"permissions"`
}

// GrantedMenuIDs 返回请求中显式指定的菜单 ID 列表。
// 仅在 HasGrantedMenuIDs() 为 true 时调用才有意义。
func (r CreateRoleRequest) GrantedMenuIDs() []uint {
	if r.MenuIDs != nil {
		return *r.MenuIDs
	}
	if r.Permissions != nil {
		return *r.Permissions
	}
	return nil
}

// HasGrantedMenuIDs 报告调用方是否显式提交了菜单/权限字段。
// 仅当字段存在（即使是空数组）时返回 true。
func (r CreateRoleRequest) HasGrantedMenuIDs() bool {
	return r.MenuIDs != nil || r.Permissions != nil
}

// RoleItem 角色列表项
type RoleItem struct {
	ID          uint   `json:"id" example:"1"`
	Name        string `json:"name" example:"超级管理员"`
	Code        string `json:"code" example:"super"`
	Status      int    `json:"status" example:"1"`
	Remark      string `json:"remark" example:"拥有所有权限"`
	Permissions []uint `json:"permissions"`
	CreateTime  string `json:"createTime" example:"2026/01/01 00:00:00"`
}
