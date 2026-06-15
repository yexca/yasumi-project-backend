# Yasumi Backend

Yasumi 的 Go 后端服务，包含一方账号体系和 MVP 阶段的同步写入校验边界。

## 当前内容

- Go 可执行入口位于 `cmd/yasumi-api` 和 `cmd/yasumi-migrate`。
- 内部包包含应用组装、类型化配置、HTTP 路由、结构化日志、数据库迁移和 PostgreSQL 仓储访问。
- 基础设施端点：`GET /healthz`、`GET /readyz`、`GET /metrics`。
- 账号相关接口：`POST /v1/auth/register`、`POST /v1/auth/login`、`POST /v1/auth/logout`、`POST /v1/auth/refresh`、`GET /v1/session`、`POST /v1/sync/token`。
- 同步上传适配接口：`POST /v1/sync/upload`。
- 内置 PostgreSQL 迁移，覆盖账号表、`items`、`recurring_task_templates`、`areas`、`operation_history` 和 `user_settings`。
- 本地环境示例：`.env.example` 和 `env/local.env.example`。
- 根目录 `Dockerfile` 可构建 API 运行镜像和迁移命令。
- 根目录 `docker-compose.example.yml` 可从项目根目录构建并运行本地服务栈。

目前没有实现 MVP 业务 CRUD 路由。同步业务写入通过 sync upload 边界进入。

## 本地命令

如果本机已安装 Go：

```powershell
go test ./...
go fmt ./...
go vet ./...
go run ./cmd/yasumi-api
go run ./cmd/yasumi-migrate
```

如果本机没有 Go，可以使用 Docker 工具链：

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain gofmt -w cmd internal
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go vet ./...
```

## Docker 本地环境

推荐使用根目录的示例配置：

```powershell
Copy-Item .env.example .env
docker compose -f .\docker-compose.example.yml up --build
```

默认本地 Docker 镜像名称为 `yexca/yasumi-backend:0.1.0`。

从项目根目录执行 Docker Compose 时，Compose 会自动读取根目录 `.env` 用于变量替换。`.env` 已被 Git 和 Docker build context 忽略，本地密钥应放在 `.env`，只提交 `.env.example`。

公开发布镜像时，`YASUMI_SYNC_TOKEN_SECRET` 这类运行时密钥应通过环境变量或密钥注入提供，不应固化进镜像。

默认服务栈会启动 PostgreSQL、PowerSync 所需的 MongoDB、PowerSync，自带执行数据库迁移，然后启动 API。

可访问：

```text
http://localhost:7659/healthz
http://localhost:7659/readyz
http://localhost:7659/metrics
```

本地前端默认请求 `http://127.0.0.1:7659`。如果从其他源访问前端，请把对应源加入 `YASUMI_HTTP_ALLOWED_ORIGINS`。

单独执行迁移：

```powershell
docker compose -f .\docker-compose.example.yml run --rm migrate
```

如果需要显式构建版本化运行镜像：

```powershell
docker build -t yexca/yasumi-backend:0.1.0 .
```

当推送 `0.1.0` 或 `v0.1.0` 这类版本 tag，或手动触发 `Docker Publish` 工作流时，GitHub Actions 会把运行镜像发布到 Docker Hub。该工作流使用 `yexca` 账号，并要求仓库已配置 `DOCKERHUB_TOKEN` secret。

默认命令会一并启动本地同步验证所需的 PowerSync 基础设施：

```powershell
docker compose -f .\docker-compose.example.yml up --build
```

当前端仓库位于相邻目录 `../yasumi-project-frontend` 时，同一个 Compose 文件也可以启动前端容器：

```powershell
docker compose -f .\docker-compose.example.yml --profile frontend up --build
```

`/readyz` 会在 PowerSync 可达后才返回就绪；如果只想看进程是否存活，使用 `/healthz`。

旧的 `env/docker-compose.yml` 开发环境仍然可用，且现在同样默认包含 PowerSync 基础设施：

```powershell
docker compose --env-file .\env\local.env.example -f .\env\docker-compose.yml up --build
```

## 配置

程序启动时只读取环境变量。根目录 Docker Compose 的默认值见 `.env.example`；旧的 `env/docker-compose.yml` 流程见 `env/local.env.example`。

Go 程序本身不会自动加载 `.env` 文件。Docker Compose 会读取根目录 `.env` 做变量替换，并把配置后的环境变量传入容器。

配置无效时，程序会在启动阶段快速失败。

## 本地账号测试

创建本地账号并使用返回的 `access_token` 调用认证接口：

```powershell
$auth = Invoke-RestMethod -Uri http://localhost:7659/v1/auth/register -Method Post -ContentType "application/json" -Body "{\"username\":\"local_user\",\"email\":\"local@example.com\",\"password\":\"password123\"}"
$token = $auth.session.access_token
curl -H "Authorization: Bearer $token" http://localhost:7659/v1/session
curl -X POST -H "Authorization: Bearer $token" -H "Content-Type: application/json" -d "{\"device_id\":\"device-01\",\"client_version\":\"0.1.0\"}" http://localhost:7659/v1/sync/token
curl -X POST -H "Authorization: Bearer $token" -H "Content-Type: application/json" -d "{\"client_batch_id\":\"batch-01\",\"device_id\":\"device-01\",\"mutations\":[]}" http://localhost:7659/v1/sync/upload
```

## 发布准备

文档索引见 `documents/README.md`。

部署和运维检查见 `documents/01-original-development/deployment-operations.md`。

后端 MVP 发布清单见 `documents/01-original-development/mvp-release-checklist.md`。

## 历史代码

旧的 Gin/Gorm 原型保留在 `legacy/old-gin-gorm`，仅供参考。当前有效代码使用 `cmd/` 和 `internal/` 结构。
