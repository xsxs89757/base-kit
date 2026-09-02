# 更新日志

base-kit 的版本记录。格式参考 Keep a Changelog，版本号遵循语义化版本。

版本规则：MINOR 只增不改（新配置键带默认值、新种子、新接口）；数据库变更只增列不删列；
破坏性变更走 `/v2` 模块路径。

## [Unreleased]

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
