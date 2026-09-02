package validator

import (
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var (
	V        *validator.Validate
	messages sync.Map // key: "StructName.FieldName.tag" → Chinese message
)

func Init() {
	V = validator.New(validator.WithRequiredStructEnabled())
}

// RegisterMessage 注册字段级中文错误信息 (类似 FluentValidation 的 WithMessage)。
// key 格式: "StructName.FieldName.tag"，例如 "LoginRequest.Username.required"
func RegisterMessage(key, msg string) {
	messages.Store(key, msg)
}

// BindAndValidate 解析请求体并执行校验，失败时返回 *fiber.Error 中断请求。
func BindAndValidate(c *fiber.Ctx, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "请求参数格式错误")
	}
	if err := V.Struct(dst); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			return fiber.NewError(fiber.StatusBadRequest, formatErrors(ve))
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return nil
}

func formatErrors(ve validator.ValidationErrors) string {
	msgs := make([]string, 0, len(ve))
	for _, fe := range ve {
		msgs = append(msgs, translateFieldError(fe))
	}
	return strings.Join(msgs, "; ")
}

func translateFieldError(fe validator.FieldError) string {
	key := fe.StructNamespace() + "." + fe.Tag()
	if msg, ok := messages.Load(key); ok {
		return msg.(string)
	}
	return defaultMessage(fe)
}

// defaultMessage 通用 tag → 中文兜底翻译
func defaultMessage(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + "不能为空"
	case "min":
		return field + "长度不能少于" + fe.Param()
	case "max":
		return field + "长度不能超过" + fe.Param()
	case "email":
		return field + "格式不正确"
	case "oneof":
		return field + "取值必须是: " + fe.Param()
	default:
		return field + "验证失败(" + fe.Tag() + ")"
	}
}
