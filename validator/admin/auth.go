package admin

import v "github.com/xsxs89757/base-kit/validator"

func init() {
	v.RegisterMessage("LoginRequest.Username.required", "用户名不能为空")
	v.RegisterMessage("LoginRequest.Password.required", "密码不能为空")
	v.RegisterMessage("ChangePasswordRequest.OldPassword.required", "旧密码不能为空")
	v.RegisterMessage("ChangePasswordRequest.NewPassword.required", "新密码不能为空")
	v.RegisterMessage("ChangePasswordRequest.NewPassword.min", "新密码长度不能少于6位")
}
