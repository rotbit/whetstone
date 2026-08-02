# Go-Zero Layout Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将现有骨架整理为单一 `app-apis` 接入层、三个 zRPC 服务和一个异步报告 Worker，并由 Makefile 统一调用 `goctl` 生成桩代码。

**Architecture:** `app/app-apis` 合并原 user/interview REST API，并保留 WebSocket 长连接接入；业务能力通过 `user-rpc`、`interview-rpc`、`question-rpc` 暴露。报告生成保持异步 Worker，不错误地包装成同步 RPC。所有 goctl 输入文件与生成目标都固定在各自服务根目录。

**Tech Stack:** Go 1.22+、go-zero REST、zRPC、Protocol Buffers、Make。

---

### Task 1: 重组服务目录

**Files:**
- Create: `app/app-apis/app.api`
- Create: `app/app-apis/etc/app-apis.yaml`
- Move: `app/interview/ws/internal/conn/manager.go` → `app/app-apis/internal/ws/conn/manager.go`
- Move: `app/interview/ws/internal/pipeline/pipeline.go` → `app/app-apis/internal/ws/pipeline/pipeline.go`
- Move: `app/user/rpc/user.proto` → `app/user-rpc/user.proto`
- Move: `app/user/rpc/etc/user.yaml` → `app/user-rpc/etc/user.yaml`
- Move: `app/interview/rpc/interview.proto` → `app/interview-rpc/interview.proto`
- Move: `app/interview/rpc/etc/interview.yaml` → `app/interview-rpc/etc/interview.yaml`
- Move: `app/question/rpc/question.proto` → `app/question-rpc/question.proto`
- Move: `app/question/rpc/etc/question.yaml` → `app/question-rpc/etc/question.yaml`
- Move: `app/report/worker/main.go` → `app/report-worker/main.go`
- Delete: 原 `app/user`、`app/interview`、`app/question`、`app/report` 空目录与分散 API 入口

**Step 1: 合并 API 定义**

将认证、用户、面试和 `/ws` 路由统一声明为 `service app-api`，保持已有 HTTP 路径不变。目录名与运行时服务名仍为 `app-apis`；`.api` 内部标识遵循 goctl 的 `*-api` 语法约束。

**Step 2: 合并运行配置**

保留端口 `8888`，在一个配置中声明 JWT、三个 RPC 客户端和 WebSocket 公共地址。

**Step 3: 校验源文件布局**

Run: `find app -maxdepth 3 -type f | sort`

Expected: 仅出现 `app-apis`、`user-rpc`、`interview-rpc`、`question-rpc`、`report-worker` 五个一级服务目录。

### Task 2: 建立统一代码生成 Makefile

**Files:**
- Modify: `Makefile`

**Step 1: 添加工具与目录变量**

使用可覆盖的 `GOCTL ?= goctl` 与 `GO_ZERO_STYLE ?= go_zero`，避免绑定开发机路径。

**Step 2: 添加生成目标**

提供 `generate`/`gen`、`gen-app-apis`、`gen-user-rpc`、`gen-interview-rpc`、`gen-question-rpc`、`check-goctl` 和 `tidy`。

**Step 3: 验证 Makefile 实际命令**

Run: `make -n generate`

Expected: 输出一次 API 生成、三次 RPC 生成和一次 `go mod tidy`。

### Task 3: 生成 go-zero 桩代码

**Files:**
- Generate: `app/app-apis/app.go`
- Generate: `app/app-apis/internal/{config,handler,logic,svc,types}/...`
- Generate: `app/*-rpc/*.go`
- Modify: `go.mod`
- Create: `go.sum`

**Step 1: 检查或安装 goctl**

Run: `make check-goctl`；若缺失则运行 `make install-tools`。

**Step 2: 通过 Makefile 生成**

Run: `make generate`

Expected: goctl 为 API 与三个 RPC 服务生成可编译桩代码。

**Step 3: 保留 WebSocket 扩展点**

生成后的 `/ws` handler 作为协议升级入口；连接管理和流式管道骨架位于 `internal/ws`，供后续注入。

### Task 4: 同步项目文档

**Files:**
- Modify: `README.md`
- Modify: `docs/技术方案.md`
- Modify: `docs/架构图.svg`

**Step 1: 更新服务名称和目录树**

所有对外 API 描述统一为 `app-apis`，内部服务统一使用 `*-rpc` 名称。

**Step 2: 更新启动和生成说明**

说明 `make generate` 是唯一推荐的桩代码生成入口，并列出新路径。

**Step 3: 检查旧路径残留**

Run: `rg 'app/(user|interview|question|report)/(api|rpc|ws|worker)|api-gateway' README.md docs app Makefile`

Expected: 无旧目录或旧 API 网关名称残留。

### Task 5: 验证重构结果

**Files:**
- Test: 全部 Go 包与生成文件

**Step 1: 格式化**

Run: `gofmt -w app common`

**Step 2: 编译测试**

Run: `go test ./...`

Expected: 所有包编译并通过测试。

**Step 3: 验证可重复生成**

Run: `make generate && go test ./...`

Expected: 重复生成成功，且测试继续通过。
