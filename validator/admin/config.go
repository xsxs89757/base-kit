package admin

import v "github.com/xsxs89757/base-kit/validator"

func init() {
	v.RegisterMessage("ConfigRequest.ConfigKey.required", "配置键不能为空")
	v.RegisterMessage("ConfigRequest.Status.oneof", "状态取值不合法")
}
