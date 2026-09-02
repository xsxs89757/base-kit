# 更新日志

base-kit 的版本记录。格式参考 Keep a Changelog，版本号遵循语义化版本。

版本规则：MINOR 只增不改（新配置键带默认值、新种子、新接口）；数据库变更只增列不删列；
破坏性变更走 `/v2` 模块路径。

## [Unreleased]

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
框架层的 bug 修复和新功能不必等模板发版，下游 `cd server && go get -u github.com/xsxs89757/base-kit` 即可。

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
