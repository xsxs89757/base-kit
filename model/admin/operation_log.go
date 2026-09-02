package admin

import "time"

type OperationLog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    uint      `json:"userId" gorm:"index"`
	Username  string    `json:"username" gorm:"size:64;index"`
	Method    string    `json:"method" gorm:"size:10"`
	Path      string    `json:"path" gorm:"size:256"`
	Status    int       `json:"status"`
	Duration  int64     `json:"duration"`
	IP        string    `json:"ip" gorm:"size:64"`
	UserAgent string    `json:"userAgent" gorm:"size:512"`
	Body      string    `json:"body" gorm:"type:text"`
}

func (OperationLog) TableName() string {
	return "sys_operation_logs"
}
