# base-kit

[base 基底](https://github.com/xsxs89757/base) 的后端框架层：配置、JWT 鉴权、菜单权限码 RBAC、
数据层与整套系统管理模块（用户 / 角色 / 菜单 / 部门 / 配置 / 操作日志）。

项目从 base 模板派生，模板负责 `main.go`、业务代码和整个前端；框架层的 bug 修复和新功能
走 `go get -u github.com/xsxs89757/base-kit`，不必再靠 git merge 一个个文件合。

> 代码历史：这些包原先在 [xsxs89757/base](https://github.com/xsxs89757/base) 的 `server/internal/`
> 下，`bf985b6` 之前的提交记录在那个仓库里。

## 用法

```go
package main

import (
    "github.com/xsxs89757/base-kit"

    "base/internal/router"
    "base/internal/store"
)

func main() {
    basekit.Run(basekit.Options{
        Models: store.ProjectModels(), // 下游模型，并入 AutoMigrate
        Seed:   store.ProjectSeed,     // 下游种子数据
        Routes: router.Setup,          // 下游业务路由
    })
}
```

`basekit.Options` 的全部扩展点：

| 字段 | 用途 |
| --- | --- |
| `ConfigPath` / `Config` | 配置文件路径（默认 `config.yaml`）；`Config` 非空时直接注入，测试用 |
| `Models` | 追加到 AutoMigrate 的下游模型 |
| `Seed` | 基底种子数据之后执行，种下游的菜单与业务数据 |
| `PreRoutes` | 在基底 `/admin` 路由**之前**注册；Fiber 先注册先匹配，用来覆盖基底某个接口 |
| `Routes` | 在基底路由之后注册业务路由 |
| `Swagger` | `enable_swagger` 为 true 时调用，由下游挂载 UI（kit 不依赖 swag） |
| `Fiber` | `fiber.New` 之前调整配置 |

只要数据层（定时任务、命令行工具）时用 `basekit.Bootstrap`；要拿到 `*fiber.App` 自己控制监听
（测试、自定义 Listener）时用 `basekit.NewApp`。

## 给基底的表加列

**只声明表名、主键和新列**，登记到 `Options.Models`：

```go
type UserExtension struct {
    ID    uint   `gorm:"primarykey"`
    Dept  string `gorm:"size:64;index"`
    Level int    `gorm:"default:0"`
}

func (UserExtension) TableName() string { return "sys_users" }
```

AutoMigrate 只增不删，kit 先建表、这个结构随后迁移同一张表，只会补上新列。

**不要嵌入 `adminmodel.User`。** 嵌入会把 `Roles` many2many 一起带过来，GORM 按新结构体名
派生外键，往共享的 `user_roles` 表里加一列 `<新结构名>_id`；通过嵌入结构写入时 `user_id` 是 NULL，
而 kit 的 handler 按 `user_id` 查角色——角色关联会静默失效。加 `gorm:"-"` 屏蔽 `Roles` 也挡不住那一列。
约束由 `store/embed_test.go` 锁定。

改字段之外的需求（比如给用户加一段结构化档案），更推荐旁表：`biz_user_profiles{user_id uniqueIndex, ...}`，
与 kit 零耦合。

## 覆盖基底的接口

Fiber 先注册先匹配，在 `PreRoutes` 里注册同样的方法和路径即可：

```go
basekit.Run(basekit.Options{
    PreRoutes: func(app *fiber.App) {
        app.Get("/admin/system/user/list", myUserList) // 盖掉 kit 的实现
    },
})
```

## 下游路由的权限码

```go
middleware.RegisterRoutePermissions(
    middleware.RoutePermission{Method: "GET", Path: "/admin/shop/order/list", Code: "Shop:Order:List"},
)
```

未登记的 `/admin` 路由对非 super 用户一律 403（默认拒绝）。只需登录不校验权限码的用
`middleware.RegisterAuthenticatedRoutes`。注册要在 `basekit.Run` 之前或 `Routes` 回调里完成。

## 从模板迁移

基底 v2.0.0 之前，这些包在模板的 `server/internal/` 下。同步到 v2.0.0 后跑一次导入路径改写：

```bash
go run github.com/xsxs89757/base-kit/cmd/basekit-migrate@latest ./...
cd server && go mod tidy
```

`base/internal/router` 和 `base/internal/store` 不在改写范围内：模板 v2.0.0 之后
仍有这两个同名的本地包（路由挂载点，和转发 `DB` / `IsUniqueViolation` / `syncSeedMenu` 的数据层垫片），
所以 `store.DB`、`store.IsUniqueViolation` 这些写法一个字都不用改。
kit 里 store 的其余能力（`SyncSeedMenus`、`RemoveLegacySeedMenus` 等）需要时直接 import kit。

### 自己往基底目录里加过文件的情况

如果之前在 `internal/handler/admin/`、`internal/model/admin/` 这类目录里放过自己的文件
（真实下游里见过一个目录 15 个），同步之后基底只删了自己那份，你的文件会留在原地。
这时同一个 import 路径下有两个来源：本地包里你的函数，和 kit 包里基底的函数——
改写把引用方指向了 kit，你的函数就会 `undefined`，而错误信息看不出根因。

改写工具会在结束时把这些目录列出来。两种处理方式：

1. **把它们挪到自己的包**（推荐），比如 `internal/handler/biz/`，再改引用方的 import。
   以后基底再动这些目录也不会碰到你。
2. 保留原地，把引用方改成同时 import kit 包和本地包，各起一个别名。

顺带一提，如果你的文件用到了基底同包里的**未导出**标识符（比如测试里的 `newTestApp`），
那些标识符现在在 kit 里且不可见，只能自己补一份。

## 版本

语义化版本。MINOR 只增不改（新配置键带默认值、新种子、新接口）；数据库变更只增列不删列；
破坏性变更走 `/v2` 模块路径。变更记录见 [CHANGELOG.md](CHANGELOG.md)。
