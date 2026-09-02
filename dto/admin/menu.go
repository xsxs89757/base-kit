package admin

// MenuRequest 创建/更新菜单请求
type MenuRequest struct {
	ParentID  uint   `json:"pid" example:"0"`
	Name      string `json:"name" validate:"required" example:"Dashboard"`
	Path      string `json:"path" example:"/dashboard"`
	Component string `json:"component" example:"/dashboard/index"`
	Redirect  string `json:"redirect"`
	Type      string `json:"type" validate:"required,oneof=catalog menu button embedded link" example:"menu"`
	Icon      string `json:"icon" example:"lucide:layout-dashboard"`
	Title     string `json:"title" validate:"required" example:"仪表盘"`
	AuthCode  string `json:"authCode" example:"System:Menu:List"`
	// OrderNo 用指针：区分"未提交"和"显式提交 0"。
	// 未提交时 Update 不会重置已有排序值，避免前端少填表单字段就把顺序打乱。
	OrderNo   *int   `json:"order" example:"1"`
	Status    int    `json:"status" validate:"oneof=0 1" example:"1"`
	KeepAlive bool   `json:"keepAlive"`
	AffixTab  bool   `json:"affixTab"`
	IframeSrc string `json:"iframeSrc"`
	Link      string `json:"link"`

	// Vben Admin 路由 meta 装饰字段
	ActiveIcon         string `json:"activeIcon" example:""`
	ActivePath         string `json:"activePath" example:""`
	HideInMenu         bool   `json:"hideInMenu"`
	HideInBreadcrumb   bool   `json:"hideInBreadcrumb"`
	HideInTab          bool   `json:"hideInTab"`
	HideChildrenInMenu bool   `json:"hideChildrenInMenu"`
	BadgeType          string `json:"badgeType" validate:"omitempty,oneof=dot normal" example:""`
	Badge              string `json:"badge" example:""`
	BadgeVariants      string `json:"badgeVariants" example:""`
}
