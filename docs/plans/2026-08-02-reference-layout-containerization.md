# Reference Layout And Containerization Implementation Plan

> 历史计划：目录重构部分已完成；单容器部署部分已由 `docs/adr/0003-dokploy-multi-container-compose.md` 替代。

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 Whetstone 重构为参考项目风格的领域分层 go-zero Monorepo，并把五个服务可靠地部署到同一个容器。

**Architecture:** API 接入层位于 `app/apis/cmd/app-apis`，业务域分别位于 `app/<domain>/rpc`，报告 Worker 位于 `app/pump/cmd/report-worker`。每个服务继续编译为独立二进制，由同一容器内的 Supervisor 管理；将来拆分为多容器时无需改变业务代码。

**Tech Stack:** Go 1.22、go-zero 1.8.3、goctl、Protocol Buffers、Docker、Alpine Linux、Supervisor、Make。

---

### Task 1：迁移项目目录

**Files:**
- Move: `app/app-apis` → `app/apis/cmd/app-apis`
- Move: `app/user-rpc` → `app/user/rpc`
- Move: `app/interview-rpc` → `app/interview/rpc`
- Move: `app/question-rpc` → `app/question/rpc`
- Move: `app/report-worker` → `app/pump/cmd/report-worker`
- Move: `app/apis/cmd/app-apis/app.api` → `app/apis/cmd/app-apis/desc/app.api`
- Move: each `*.proto` → corresponding `rpc/pb/`

**Step 1:** 创建接入层、领域层和后台任务目录。

**Step 2:** 移动源文件并将所有 Go import 更新到新路径。

**Step 3:** 将 proto 的 `go_package` 统一改为 `./pb`。

**Step 4:** 运行 `find app -maxdepth 5 -type d | sort`，确认不存在旧扁平服务根目录。

### Task 2：重新生成 API 与 RPC 桩代码

**Files:**
- Generate: `app/apis/cmd/app-apis/internal/...`
- Generate: `app/{user,interview,question}/rpc/{client,internal,pb}/...`

**Step 1:** 清理旧路径下仅由 goctl 生成的 RPC 桩代码。

**Step 2:** 使用 goctl 多服务模式从 `rpc/pb/*.proto` 生成代码到 `rpc`。

**Step 3:** 更新 `app-apis/internal/svc/service_context.go` 的 RPC client import。

**Step 4:** 运行 `go test ./...`，预期所有包编译通过。

### Task 3：重写 Makefile

**Files:**
- Modify: `Makefile`

**Step 1:** 增加参考项目风格的 `make generate type=<api|rpc> name=<name>`。

**Step 2:** 增加 `generate-all`、`generate_all`、`build type=...`、`build-all`、`docker-build`。

**Step 3:** 保留 `install-tools`、`test`、`tidy`、`up`、`down`。

**Step 4:** 运行 `make -n generate-all`，预期输出一次 API 和三次 RPC 生成。

### Task 4：增加单容器运行能力

**Files:**
- Create: `Dockerfile`
- Create: `.dockerignore`
- Create: `deploy/supervisor/supervisord.conf`
- Modify: `app/pump/cmd/report-worker/main.go`
- Modify: `app/apis/cmd/app-apis/desc/app.api`

**Step 1:** 为 `app-apis` 增加无需鉴权的 `/healthz`。

**Step 2:** 让报告 Worker 在未接入 asynq 时等待 SIGINT/SIGTERM，而不是立即退出。

**Step 3:** 用多阶段 Dockerfile 编译五个静态二进制，并以非 root 用户运行 Supervisor。

**Step 4:** 配置 Supervisor 的启动顺序、异常重启、日志输出和进程组停止。

**Step 5:** 运行 Dockerfile 静态检查或实际构建，预期镜像构建成功。

### Task 5：同步文档并完成验证

**Files:**
- Modify: `README.md`
- Modify: `docs/技术方案.md`
- Modify: `docs/架构图.svg`

**Step 1:** 更新目录树、生成命令、构建命令和单容器运行说明。

**Step 2:** 记录单容器部署的限制与未来拆分路径。

**Step 3:** 运行 `make generate-all`、`gofmt -w app common`、`go test ./...`。

**Step 4:** 运行 `git diff --check` 并搜索所有旧目录引用，预期均无错误。
