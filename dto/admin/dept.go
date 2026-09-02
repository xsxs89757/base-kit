package admin

// DeptRequest 创建/更新部门请求
type DeptRequest struct {
	ParentID uint   `json:"pid" example:"0"`
	Name     string `json:"name" validate:"required" example:"技术部"`
	OrderNo  int    `json:"order" example:"1"`
	Status   int    `json:"status" validate:"oneof=0 1" example:"1"`
	Remark   string `json:"remark"`
}
