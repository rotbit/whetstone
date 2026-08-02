# Whetstone 磨枪

> 临阵磨枪——快，而且光。

**Whetstone**（磨刀石）是一个开源 AI 面试陪练平台：上传简历 + 目标 JD，AI 面试官基于你的项目经历实时提问、追问（文字 / 语音），面完生成逐题评分的复盘报告。

技术栈：Go + [go-zero](https://go-zero.dev) 微服务 · WebSocket 实时流 · MySQL / Redis / Qdrant / asynq

📄 完整设计见 [docs/技术方案.md](docs/技术方案.md) · 架构图见 [docs/架构图.svg](docs/架构图.svg)
🖥 前端仓库：`whetstone-web`

## 目录结构

```
whetstone/
├── app/
│   ├── apis/cmd/app-apis/   # 统一 REST + WebSocket 接入（:8888）
│   │   ├── desc/app.api     #   goctl API 定义
│   │   ├── etc/             #   网关与 RPC 客户端配置
│   │   └── internal/ws/     #   WebSocket 连接管理与流式管道
│   ├── user/rpc/            # 用户 / 次数幂等扣减（:9001）
│   │   ├── pb/              #   proto 与生成代码
│   │   └── client/          #   zRPC 客户端
│   ├── interview/rpc/       # 面试引擎：状态机 / 追问决策（:9002）
│   ├── question/rpc/        # 题库 + RAG 个性化出题（:9003）
│   └── pump/cmd/
│       └── report-worker/   # asynq 消费者：异步评分与报告
├── common/
│   └── llmadapter/          # LLM · ASR · TTS 多厂商统一封装（mini AI 网关）
├── deploy/                  # 本地与 Dokploy Compose 配置
├── Dockerfile               # 按 SERVICE 构建单个服务镜像
├── sql/                     # 数据库 DDL
└── docs/                    # 技术方案 / 架构图
```

约定：`apis/cmd` 是统一 BFF 接入层，`<domain>/rpc` 是业务域，`pump/cmd` 是后台任务。只有 `app-apis` 对外提供 REST 与 WebSocket；goctl 统一使用 `go_zero` 命名风格。

## 快速开始

```bash
# 0. 把 module 占位符替换成你的 GitHub 用户名（macOS）
grep -rl 'github.com/yourname/whetstone' . | xargs sed -i '' 's#yourname#你的用户名#g'

# 1. 安装 goctl 与 protoc 插件
make install-tools

# 2. 通过 Makefile 生成全部 go-zero 桩代码
make generate-all

# 也可以只生成一个服务
make generate type=api name=app-apis
make generate type=rpc name=user

# 3. 启动基础设施
make up

# 4. 建表
mysql -h127.0.0.1 -uroot -proot whetstone < sql/schema.sql

# 5. 本地开发时分别启动服务
go run ./app/user/rpc          -f app/user/rpc/etc/user.yaml
go run ./app/interview/rpc     -f app/interview/rpc/etc/interview.yaml
go run ./app/question/rpc      -f app/question/rpc/etc/question.yaml
go run ./app/apis/cmd/app-apis -f app/apis/cmd/app-apis/etc/app-apis.yaml
go run ./app/pump/cmd/report-worker
```

## 多容器部署

```bash
# 编译五个本地二进制
make build-all

# 构建五个独立镜像
make docker-build-all

# 也可以只构建一个服务
make docker-build SERVICE=user-rpc

# 启动五个应用容器及 MySQL、Redis、Qdrant、MinIO
make up

# 健康检查
curl http://127.0.0.1:8888/healthz
```

每个容器只运行一个 Go 进程：`app-apis`、三个 RPC 和 `report-worker` 可以独立查看日志、重启和配置资源。只有 `app-apis:8888` 对外发布，RPC 通过 Compose 内部 DNS 通信；生产镜像内仅包含 YAML 模板，真实密钥必须在运行时注入，不能提交到仓库。

部署到 Dokploy 时使用一个包含五个容器的 Compose 项目，生产配置默认通过环境变量注入，也支持 File Mount 覆盖完整 YAML。具体设置见 [deploy/dokploy/README.md](deploy/dokploy/README.md)。

## M1 路线（文字版 MVP，4-6 周）

- [ ] 用户注册登录（JWT）+ 免费次数
- [ ] 简历 / JD 上传与 LLM 结构化解析
- [ ] llmadapter：DeepSeek ChatStream（SSE 流式）
- [ ] 面试引擎状态机：提问 → 回答 → 追问（≤2 层）
- [ ] app-apis WebSocket：文字模式实时对话（心跳 / 断线恢复）
- [ ] asynq 异步复盘报告
- [ ] docker compose 一键起全套

## License

[MIT](LICENSE)
