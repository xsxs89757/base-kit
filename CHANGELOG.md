# 更新日志

base-kit 的版本记录。格式参考 Keep a Changelog，版本号遵循语义化版本。

版本规则：MINOR 只增不改（新配置键带默认值、新种子、新接口）；数据库变更只增列不删列；
破坏性变更走 `/v2` 模块路径。

## [Unreleased]

### 修复

- 下游在 `/admin` 下重复挂 `JWTAuth` / `PermissionAuth` / `OperationLog` 时，每个写操作记两条操作日志，
  鉴权和权限码也各跑两遍。`SetupAdmin` 把这三个中间件挂在 `/admin` 前缀上，`Routes` 里注册的 `/admin` 路由
  本来就经过它们，而模板 `router/project.go` 的注释让下游「参考 admin.go 中 protected 分组的中间件挂法」再挂一遍；
  `PreRoutes` 的分组上挂了、请求又落到后面 kit 或 `Routes` 的路由，也是两遍。现在三个中间件同一请求只生效一次，
  已经重复挂载的下游升级后不改代码也只记一条。`PermissionAuth` 按 method + path 记放行，
  内部重定向（`c.Path(...)` 后 `RestartRouting`）到别的路由照样重新判定。
- 操作日志在 MySQL 上记不下文件上传。multipart 请求体是二进制，utf8mb4 列在严格模式下拒绝整条 INSERT
  （`Error 1366 Incorrect string value: '\x89PNG...'`），writer 只打一行日志，这条记录就没了；SQLite 上一切正常，
  所以本地看不出来。现在上传请求改记表单摘要：普通字段照录（敏感字段照样脱敏，JSON 字符串字段里的也脱敏），
  文件只记文件名、大小和类型，例如 `{"category":"头像","file":{"filename":"a.png","size":60016,"contentType":"image/png"}}`。
- 同一根因的另外几种丢日志：请求体按字节截到 2KB，会切开中文等多字节字符（超过 2KB 的中文表单大概率整条丢失）；
  path / User-Agent 超出列宽触发 `Error 1406`。现在按字符边界截断，所有列写库前统一替换非法 UTF-8、去掉 NUL、
  按列宽截断。PostgreSQL 上是同样的问题（`SQLSTATE 22021` / `22001`），一并修好。
- 操作日志条目直接持有 Fiber 返回的字符串，它们指向会被后续请求复用的缓冲区，而日志是后台 writer 异步写库的，
  写库时可能已经变成别的请求的内容：并发时 method / path / User-Agent 会记错（实测 `DELETE /system/user/1`
  被记成 `POSTTE /system/role/2`）。现在入队前拷贝。

### 文档

- README「覆盖基底的接口」的示例漏挂中间件：`PreRoutes` 排在 kit 的 `/admin` 中间件前面，照抄出来的接口
  不用登录就能调。示例补上三个中间件；新增「下游路由的中间件」一节，说明哪些路由已经挂好、哪些要自己挂。
  `SetupAdmin`、`Options.PreRoutes` / `Routes` 和三个中间件的注释同步写明。

## [1.0.3] - 2026-09-02

### 变更

- 生产模式（`server.mode: production`）下 5xx 响应不再回显内部错误原文，对外统一返回
  `Internal Server Error`，原文进日志。recover 中间件把 panic 也转成这类错误，之前 DB 错误、文件路径
  会原样出现在响应体里。开发模式行为不变，4xx（404/405 等）不受影响。

### 修复

- `basekit-migrate` 的推荐用法改为在 `server/` 下不带 `@latest` 运行，版本跟随 go.mod 里钉的 kit。
  `@latest` 在 GOPROXY 有延迟的几分钟里会拿到旧版——而 v1.0.1 正好带着「改写 `base/internal/store`」那个 bug。
- `basekit-migrate` 结束时提示 import 分组未重排（kit 路径留在原来 `base/...` 那一组），用 `goimports -w .` 归位。

### 文档

- `go get -u` 改为 `go get ...@latest`：`-u` 会把 kit 的依赖一并升到最新 minor，与模板钉死版本的初衷相反。
- `Bootstrap` 不传 `Seed` 时模板垫片里的 `store.DB` 为 nil，README 点明。

## [1.0.2] - 2026-09-02

### 修复

- `basekit-migrate` 不再改写 `base/internal/store`。模板 v2.0.0 之后仍有一个同名的本地包
  （数据层挂载点的垫片，转发 `DB` / `IsUniqueViolation` 并给 `project.go` 提供 `syncSeedMenu`），
  改写会让 `main.go` 里的 `store.ProjectModels` / `store.ProjectSeed` 变成 undefined。
  顺带的好处：下游写 `store.DB`、`store.IsUniqueViolation` 的地方一个字都不用改。
  拿脚手架建的下游走完整升级流程时发现的。

## [1.0.1] - 2026-09-02

### 新增

- `basekit-migrate` 结束时检查 kit 接管的目录里是否还留着下游自己的 `.go` 文件，
  列出来并说明怎么处理。拿真实下游试跑时发现的：一个项目在 6 个基底目录里放了 48 个自己的文件，
  改写后引用方指向 kit，那些函数全部 `undefined`，报错信息看不出根因。
- `cmd/basekit-migrate` 补测试：映射表命中与不命中（`base/internal/router` 和下游自建包必须原样保留）、
  残留目录提示（空目录和未改写的包不报）。

## [1.0.0] - 2026-09-02

首个稳定版。API 与 v0.1.0 一致，配套 [base](https://github.com/xsxs89757/base) 模板 v2.0.0
（模板的 `server/go.mod` 从此钉这个版本）。

从这个版本起遵守上面的版本规则：MINOR 只增不改，数据库变更只增列，破坏性变更走 `/v2`。
框架层的 bug 修复和新功能不必等模板发版，下游 `cd server && go get github.com/xsxs89757/base-kit@latest` 即可。

### 验证

- 模板 `server/main_test.go` 冒烟测试（启动 → 登录 → 用户信息 → 菜单 → Swagger）通过。
- 用 `--parseDependencyLevel 3 --packagePrefix base,github.com/xsxs89757/base-kit` 重新生成的
  `swagger.json` 与抽库前逐字节一致。

## [0.1.0] - 2026-09-02

从 [xsxs89757/base](https://github.com/xsxs89757/base) 的 `server/internal/` 抽出框架层与系统管理模块，
首个可用版本（模板尚未切换过来，等 v1.0.0 一起发布）。

### 新增

- 根包 `basekit`：`Options` / `Bootstrap` / `NewApp` / `Run`，取代模板里手写的 `main.go` 装配逻辑。
  下游扩展点：`Models`、`Seed`、`PreRoutes`（覆盖基底接口）、`Routes`、`Swagger`、`Fiber`。
- `store.Init(Options) error`：不再 `log.Fatal`，由调用方决定进程存亡；模型与种子数据通过参数传入，
  取代模板时代靠同包文件 `project.go` 的编译期挂载。
- 导出种子助手 `SyncSeedMenus` / `SyncSeedMenu` / `RefreshRoleMenus` / `RemoveLegacySeedMenus` 和 `MenuDef`。
- `config.Path()` / `config.LoadExtra(dst)`：下游把自己的配置段读进自己的结构体。
- `cmd/basekit-migrate`：go/ast 导入路径改写工具，下游从模板迁移时跑一次。
- `store/embed_test.go` 锁定「给基底表加列」的扩展方式，`store/mysql_test.go` 在真实 MySQL 上验证迁移与种子。

### 已知约束

- 给 `sys_users` 等基底表加列时，扩展结构**只能声明表名、主键和新列**，不能嵌入 `adminmodel.User`：
  嵌入会把 `Roles` many2many 带过来，往共享的 `user_roles` 表加一列 `<结构名>_id`，
  通过嵌入结构写入时 `user_id` 为 NULL，kit 的 handler 按 `user_id` 查角色会静默查不到。详见 README。
