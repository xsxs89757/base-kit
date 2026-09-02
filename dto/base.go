package dto

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Response 统一响应结构
type Response struct {
	Code    int    `json:"code" example:"0"`
	Data    any    `json:"data"`
	Error   any    `json:"error"`
	Message string `json:"message" example:"ok"`
}

// PageData 分页数据
type PageData struct {
	Items any   `json:"items"`
	Total int64 `json:"total" example:"100"`
}

// PageResponse 分页响应 (Swagger 用)
type PageResponse struct {
	Code    int      `json:"code" example:"0"`
	Data    PageData `json:"data"`
	Error   any      `json:"error"`
	Message string   `json:"message" example:"ok"`
}

// IDResponse 创建成功返回 ID
type IDResponse struct {
	ID uint `json:"id" example:"1"`
}

// 分页参数边界：pageSize 无上限时 ?pageSize=100000 会把整表拖出来，
// 这里统一夹到 [1, MaxPageSize]，page 至少为 1（否则 offset 为负）。
const (
	DefaultPageSize = 20
	MaxPageSize     = 200
)

// ParsePage 解析并规范化列表接口的 page / pageSize 查询参数。
// 非数字、0、负数一律回落到默认值；pageSize 超过 MaxPageSize 时截断。
func ParsePage(c *fiber.Ctx) (page, pageSize int) {
	page, _ = strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Query("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}

func Success(c *fiber.Ctx, data any) error {
	return c.JSON(Response{
		Code:    0,
		Data:    data,
		Error:   nil,
		Message: "ok",
	})
}

func PageSuccess(c *fiber.Ctx, items any, total int64) error {
	return c.JSON(Response{
		Code: 0,
		Data: PageData{
			Items: items,
			Total: total,
		},
		Error:   nil,
		Message: "ok",
	})
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{
		Code:    -1,
		Data:    nil,
		Error:   message,
		Message: message,
	})
}
