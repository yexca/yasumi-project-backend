# Structure Optimization Guide

## 1. 当前结构的优点

现有仓库整体是健康的，特别是下面几点值得保留：

- 入口、装配、协议层、业务层、持久化层已经有基本分层。
- 主要依赖 Go 标准库，运行和维护成本较低。
- 同步写入规则集中在 `service` 和 `domain`，方向是对的。
- SQL 显式可见，调试和排障成本可控。

所以当前结构更适合做渐进优化，而不是推倒重来。

## 2. 最值得优先优化的点

### 2.1 拆分 `internal/service/sync_upload.go`

当前同步上传主文件已经承载了解码、归一化、校验、语义流转、幂等检查和写入编排，现状已经接近 800 行，后续继续扩展时容易把多个聚合绑得过紧。

建议按职责拆成类似下面的形式：

- `sync_upload_service.go`
  - 只保留入口编排和事务流程。
- `sync_upload_items.go`
  - 只处理 `items` 相关接受逻辑。
- `sync_upload_areas.go`
  - 只处理 `areas`。
- `sync_upload_recurring_templates.go`
  - 只处理 `recurring_task_templates`。
- `sync_upload_operations.go`
  - 只处理 `operation_history`。
- `sync_upload_settings.go`
  - 只处理 `user_settings`。
- `sync_upload_validation.go`
  - 放通用语义校验和错误映射。

这样做的收益：

- 每个聚合的规则集中，内聚更高。
- 新增同步表时不会持续膨胀单文件。
- 测试更容易对准单一责任。

### 2.2 拆分 `internal/repository/repository.go`

当前 repository 文件同时承载了账户、会话、同步数据、设置等多类 SQL，现状已经超过 1100 行。随着需求增长，这会让读写边界越来越模糊。

建议按主题拆分：

- `repository.go`
  - 仅保留 `Repository`、`Tx`、`InTx`、公共错误。
- `account_repository.go`
  - 用户、凭证、会话相关 SQL。
- `sync_items_repository.go`
  - `items` 与 `operation_history`。
- `sync_meta_repository.go`
  - `areas`、`recurring_task_templates`、`user_settings`。

注意：拆文件即可，先不要急着拆成过多包。当前一个 `repository` 包已经够用。

### 2.3 让装配层更专注于依赖拼装

`internal/app/app.go` 目前职责还算克制，但后续继续加服务时，`HTTPServer()` 可能变成很长的依赖创建清单。

建议在需要时补一个轻量 provider 层，例如：

- `internal/app/providers.go`
  - 集中创建 repository adapter、account service、sync service、token issuer。

目标不是引入复杂容器，而是让 `app.go` 继续保持可读。

### 2.4 减少 service 对 repository record 的直接语义依赖

当前 `service` 直接使用 `repository.*Record` 作为主要数据载体，短期很高效，但中长期会让业务语义和表结构一起演化。

建议只在痛点明显时逐步引入：

- 面向用例的输入结构
- 更小的 repository 接口
- 必要的转换函数

这里要克制。不要为了“解耦”而先造一层大 DTO 映射。

### 2.5 稳定契约文档入口

当前历史文档、根 README、外部开发资料之间已经有依赖关系。后续继续迭代时，建议把“当前有效规范”统一收束到 `documents/02-secondary-development/`，避免开发者在历史文档里找现行规则。

## 3. 推荐的优化顺序

建议按低风险顺序推进：

1. 先整理文档入口。
2. 再拆 `sync_upload.go`。
3. 然后拆 `repository.go`。
4. 最后视复杂度决定是否补 provider 或更细的用例结构。

这个顺序可以在不改变对外行为的前提下，持续降低维护成本。

## 4. 暂时不建议做的事

- 暂时不要把 `internal` 按过细的微包继续拆碎。
- 暂时不要引入重量级 DI 框架。
- 暂时不要把所有 SQL 改造成通用 ORM 抽象。
- 暂时不要为了统一风格而一次性重写所有 handler 和 service。

这些动作收益不一定高，但迁移成本和回归风险都不小。

## 5. 一句话结论

当前结构可继续支撑开发，最合理的路线不是重构一切，而是围绕同步写入和 repository 两个增长点做小步拆分，把复杂度拦在局部。
