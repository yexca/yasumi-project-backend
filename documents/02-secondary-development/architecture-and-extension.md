# Secondary Development Architecture and Extension Notes

## 1. 当前结构概览

当前仓库是一个以 Go 标准库为主的轻量后端，核心结构如下：

- `cmd/yasumi-api`
  - API 服务入口。
- `cmd/yasumi-migrate`
  - 数据库迁移入口。
- `internal/app`
  - 应用装配层，负责把配置、日志、数据库连接、路由和服务拼装起来。
- `internal/httpapi`
  - HTTP 路由、处理中间件、请求解析、响应输出。
- `internal/auth`
  - 账号体系、会话、访问令牌、密码校验。
- `internal/service`
  - 业务用例层，目前重点是同步上传校验与落库流程。
- `internal/repository`
  - PostgreSQL 访问层，封装事务和 SQL。
- `internal/domain`
  - 领域常量、校验规则、状态流转规则、错误定义。
- `internal/synctoken`
  - PowerSync 兼容的同步令牌签发。
- `internal/telemetry`
  - 日志与指标相关能力。
- `internal/migrations`
  - 内嵌 SQL 迁移。

## 2. 当前依赖方向

推荐继续保持下面这条主链路：

1. `cmd/*` 只做启动。
2. `internal/app` 只做装配，不承载业务规则。
3. `internal/httpapi` 负责协议转换，不直接写 SQL。
4. `internal/service` 负责业务编排和用例规则。
5. `internal/domain` 负责纯规则和语义约束。
6. `internal/repository` 负责持久化细节。

这个方向的好处是：

- 业务规则集中在 service/domain，内聚度更高。
- HTTP 和数据库可以分别演进，耦合面更小。
- 替换存储、补测试、加接口时，影响范围更容易控制。

## 3. 二次开发原则

### 高内聚

- 同一个功能的协议、用例、规则、落库应围绕一个明确主题组织。
- 新增需求时，优先补在已有边界内，而不是创建横跨多层的“万能工具”。

### 低耦合

- `httpapi` 不直接依赖具体 SQL 结构。
- `service` 不直接感知 HTTP 细节。
- `repository` 不承担业务规则判断。
- 新功能尽量通过小接口接入，而不是把整个大对象一路透传。

### 轻量化

- 优先复用现有目录与模式。
- 没有明显复用收益时，不急着抽象。
- 先做面向当前需求的最小闭环，再考虑下一层提炼。

## 4. 常见扩展方式

### 新增一个认证后的接口

建议按下面顺序落地：

1. 在 `internal/domain` 增加校验规则或错误语义。
2. 在 `internal/service` 增加用例方法。
3. 在 `internal/repository` 补充读写方法。
4. 在 `internal/httpapi/handlers.go` 增加处理函数。
5. 在 `internal/httpapi/router.go` 注册路由。
6. 视影响面补充 service、repository、router 测试。

### 扩展同步上传支持的新表

建议保持与现有 `areas`、`items`、`user_settings` 一致的节奏：

1. 先定义该表的领域校验规则。
2. 在 `repository` 增加对应 Record 和事务方法。
3. 在 `service` 中新增该聚合的 decode、normalize、validate、accept 流程。
4. 明确 revision、device_id、时间戳、幂等等服务端规则。
5. 为错误映射和冲突场景补测试。

### 新增配置项

建议只改三处：

1. `internal/config/config.go`
2. `.env.example`
3. 相关 README 或二开文档

避免把配置解析逻辑散落到业务代码里。

### 新增数据库结构

建议按下面控制影响范围：

1. 在 `internal/migrations/sql` 新增增量迁移。
2. 只为实际使用场景增加字段或索引。
3. 同步更新对应的 repository record 与 SQL。
4. 如涉及同步契约，再补充原始文档或二开文档说明。

## 5. 开发时应避免的情况

- 在 handler 中直接堆业务判断。
- 在 repository 中偷偷做状态流转判断。
- 为了少写几行代码，让 service 直接依赖 HTTP request/response。
- 大量创建跨模块共享但语义模糊的 util。
- 一次性做大范围目录重写，导致历史验证资料失效。

## 6. 推荐的增量工作方式

1. 先明确改动属于哪个模块。
2. 只打开该模块上下游一层的代码。
3. 先补规则，再补存储，再补入口。
4. 最后回写文档和提交信息。

这样更适合当前仓库，也更符合轻量维护目标。
