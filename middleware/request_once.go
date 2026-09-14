package middleware

// router.SetupAdmin 把 JWTAuth、PermissionAuth、OperationLog 挂在 /admin 前缀上，之后注册的 /admin/* 路由
// 都会经过它们。下游在自己的分组上再挂一遍（base 模板 router/project.go 的注释就是这么建议的），
// 同一请求就会经过两次：每个写操作记两条日志，鉴权和权限码各判两遍。
// 所以这三个中间件在同一请求里只生效一次，再次经过时直接放行。

// onceKey 是记录「本请求已经生效过」的 Locals 键。类型不导出，其他包（包括下游）写不进同一个键，
// 没法伪造「已经鉴权过」。
type onceKey int

const (
	// JWTAuth 已为本请求装配好身份
	jwtAuthDone onceKey = iota
	// PermissionAuth 已放行，值是放行时的 "METHOD path"
	permissionGranted
	// 外层的 OperationLog 已接手本请求
	operationLogStarted
)
