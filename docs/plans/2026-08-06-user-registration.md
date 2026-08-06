# User Registration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现手机号和密码注册接口，通过 user-rpc 将用户写入 MySQL，并由 app-apis 返回 JWT。

**Architecture:** 保持现有 `app-apis -> user-rpc -> MySQL` 分层。`users` Model 必须由固定版本 goctl
从当前 MySQL `users` 表生成；user-rpc 负责校验、密码哈希和唯一键错误转换，app-apis 只负责调用 RPC 与签发 JWT。

**Tech Stack:** Go、go-zero REST/zRPC、goctl、MySQL 8.4、bcrypt、JWT。

---

### Task 1: Generate the MySQL model

**Files:**
- Create: `app/user/model/users_model.go`
- Create: `app/user/model/users_model_gen.go`
- Create: `app/user/model/vars.go`

**Step 1:** 安装仓库固定的 goctl 1.10.2，并确认版本。

**Step 2:** 使用 `goctl model mysql datasource` 从实际 MySQL `users` 表生成 Model，只把密码放入进程内存。

**Step 3:** 检查生成代码包含 `Insert` 和 `FindOneByPhone`，并执行 `gofmt`。

### Task 2: Add the registration RPC

**Files:**
- Modify: `app/user/rpc/pb/user.proto`
- Generate: `app/user/rpc/pb/user.pb.go`
- Generate: `app/user/rpc/pb/user_grpc.pb.go`
- Generate: `app/user/rpc/client/user/user.go`
- Generate: `app/user/rpc/internal/server/user/user_server.go`
- Create: `app/user/rpc/internal/logic/user/register_logic.go`

**Step 1:** 在 proto 中声明 `Register`、`RegisterReq` 和 `RegisterResp`。

**Step 2:** 运行 `make generate type=rpc name=user`，确认 goctl 生成的客户端与服务端桩可编译。

**Step 3:** 编写注册逻辑测试，覆盖成功、非法手机号、短密码和重复手机号。

**Step 4:** 实现手机号规范化、8 至 72 字节密码校验、bcrypt 哈希和 `users` 插入。

### Task 3: Wire MySQL into user-rpc

**Files:**
- Modify: `app/user/rpc/internal/config/config.go`
- Modify: `app/user/rpc/internal/svc/service_context.go`
- Modify: `app/user/rpc/etc/user.yaml`
- Modify: `deploy/dokploy/etc/user.yaml`

**Step 1:** 增加 `Mysql.DataSource` 配置，不在仓库记录真实密码。

**Step 2:** 在 ServiceContext 中使用生成的 `model.NewUsersModel` 初始化 Model。

**Step 3:** 运行 user-rpc 单元测试，确认不依赖真实数据库即可覆盖业务分支。

### Task 4: Implement the REST endpoint

**Files:**
- Modify: `app/apis/cmd/app-apis/desc/app.api`
- Generate: `app/apis/cmd/app-apis/internal/types/types.go`
- Modify: `app/apis/cmd/app-apis/internal/logic/auth/register_logic.go`
- Test: `app/apis/cmd/app-apis/internal/logic/auth/register_logic_test.go`

**Step 1:** 将注册请求固定为 `phone` 和 `password`，继续复用现有 `TokenResp`。

**Step 2:** 运行 `make generate type=api name=app-apis` 更新 API 桩代码。

**Step 3:** 调用 `UserRpc.Register`，使用配置中的密钥和有效期签发包含 `uid` 的 HS256 JWT。

**Step 4:** 测试 RPC 请求转发、JWT claims 和 RPC 错误透传。

### Task 5: Verify with curl

**Files:**
- Verify: all changed Go and YAML files

**Step 1:** 运行 `gofmt`、`go test ./...`、`go vet ./...` 和 `git diff --check`。

**Step 2:** 使用环境变量临时注入 MySQL DSN，启动 user-rpc 和 app-apis。

**Step 3:** 用 `curl` 调用 `POST /api/v1/auth/register`，确认返回 JWT；重复调用确认返回 409。

**Step 4:** 交付可复用的 curl 命令，并说明测试账号对数据库的影响。
