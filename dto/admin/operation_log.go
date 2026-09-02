package admin

// OperationLogItem 操作日志列表项
type OperationLogItem struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	Duration   int64  `json:"duration"`
	IP         string `json:"ip"`
	UserAgent  string `json:"userAgent"`
	CreateTime string `json:"createTime"`
}
