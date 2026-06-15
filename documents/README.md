# Yasumi Documents

该目录已经按用途重新整理，方便区分历史开发记录、二次开发说明和结构优化建议。

## 目录结构

- `01-original-development/`
  - 原始开发过程文档归档。
  - 保留阶段验收、开发记录、验证记录、部署检查和 MVP 发布清单。
- `02-secondary-development/`
  - 面向后续二次开发的说明文档。
  - 重点覆盖当前架构、扩展方式、低耦合实践和提交规范。
- `03-optimization-guidance/`
  - 面向现有结构的轻量化优化建议。
  - 只给出渐进式方案，不要求大规模重写。

## 建议阅读顺序

1. 先看 `02-secondary-development/README.md`
2. 再看 `02-secondary-development/architecture-and-extension.md`
3. 提交代码前看 `02-secondary-development/contribution-guide.md`
4. 评估重构时看 `03-optimization-guidance/structure-optimization-guide.md`
5. 需要追溯历史背景时再回到 `01-original-development/`

## 约束

- 文档中不得包含真实密钥、令牌、个人信息或生产环境连接信息。
- `01-original-development/` 作为归档区，原则上只做链接修复、脱敏和必要勘误。
- 提交规范统一为 `action(module): summary`，详见 `02-secondary-development/contribution-guide.md`。
