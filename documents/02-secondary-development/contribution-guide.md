# Contribution Guide

## 1. 提交规范

本项目统一使用以下提交格式：

```text
action(module): summary
```

其中 `action` 只允许使用：

- `feat`
- `fix`
- `refactor`
- `chore`
- `docs`
- `perf`

## 2. action 含义

- `feat`: 新功能或新增能力。
- `fix`: 缺陷修复或行为纠正。
- `refactor`: 不改变外部行为的结构调整。
- `chore`: 配置、脚本、环境、维护性杂项。
- `docs`: 文档新增、整理、修订。
- `perf`: 以性能优化为主的改动。

## 3. module 命名建议

`module` 应尽量贴近实际改动边界，保持简短、可定位、不过度泛化。建议优先使用：

- `documents`
- `auth`
- `httpapi`
- `sync`
- `repository`
- `domain`
- `config`
- `migrations`
- `telemetry`
- `docker`

如果改动横跨多个目录，优先选择对外行为变化最明显的那个模块，而不是写成过于宽泛的 `core` 或 `misc`。

## 4. 提交示例

```text
feat(sync): support recurring template upload validation
fix(auth): reject expired refresh sessions
refactor(repository): split account query helpers
docs(documents): reorganize development and extension guides
chore(docker): align local compose defaults
perf(httpapi): reduce metrics label allocations
```

## 5. 文档同步要求

出现以下变化时，应同步更新文档：

- 新增或移除接口
- 新增配置项或环境变量
- 新增迁移或数据结构约束
- 调整同步写入契约
- 调整项目推荐开发流程

优先更新：

1. 根目录 `README.md` 或 `README.zh-cn.md`
2. `documents/02-secondary-development/`
3. 必要时补充 `documents/03-optimization-guidance/`

## 6. 二开提交前最小检查

- 改动是否只落在清晰的模块边界内。
- handler、service、repository 的职责是否仍然分离。
- 新规则是否优先进入 `domain` 或 `service`，而不是散在多处。
- 新配置和新接口是否补了文档。
- 提交信息是否符合 `action(module): summary`。
