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
│   ├── user/                # 用户域
│   │   ├── api/             #   REST：注册登录 / 简历 / 次数 / 订单（:8888）
│   │   └── rpc/             #   内部：用户 / 次数幂等扣减（:9001）
│   ├── interview/           # 面试域
│   │   ├── api/             #   REST：会话管理 / 复盘报告（:8889）
│   │   ├── ws/              #   WebSocket 长连接网关，实时对话（:8890，手写）
│   │   └── rpc/             #   面试引擎：状态机 / 追问决策（:9002）
│   ├── question/
│   │   └── rpc/             # 题库 + RAG 个性化出题（:9003）
│   └── report/
│       └── worker/          # asynq 消费者：异步评分与报告（手写）
├── common/
│   └── llmadapter/          # LLM · ASR · TTS 多厂商统一封装（mini AI 网关）
├── deploy/                  # docker-compose：MySQL / Redis / Qdrant / MinIO
├── sql/                     # 数据库 DDL
└── docs/                    # 技术方案 / 架构图
```

约定：每个服务的 `api`（对外 REST）与 `rpc`（对内 gRPC）放在同一域目录下，goctl 统一 `-style go_zero` 命名。

## 快速开始

```bash
# 0. 把 module 占位符替换成你的 GitHub 用户名（macOS）
grep -rl 'github.com/yourname/whetstone' . | xargs sed -i '' 's#yourname#你的用户名#g'

# 1. 安装 goctl 与 protoc 插件
make install-tools

# 2. 生成 go-zero 代码（.api / .proto 已写好）
make gen

# 3. 启动基础设施
make up

# 4. 建表
mysql -h127.0.0.1 -uroot -proot whetstone < sql/schema.sql

# 5. 分别启动服务（生成代码后）
go run app/user/rpc/user.go        -f app/user/rpc/etc/user.yaml
go run app/user/api/user.go        -f app/user/api/etc/user-api.yaml
go run app/interview/rpc/interview.go -f app/interview/rpc/etc/interview.yaml
go run app/interview/api/interview.go -f app/interview/api/etc/interview-api.yaml
go run app/interview/ws/main.go    # WebSocket 网关（骨架）
```

## M1 路线（文字版 MVP，4-6 周）

- [ ] 用户注册登录（JWT）+ 免费次数
- [ ] 简历 / JD 上传与 LLM 结构化解析
- [ ] llmadapter：DeepSeek ChatStream（SSE 流式）
- [ ] 面试引擎状态机：提问 → 回答 → 追问（≤2 层）
- [ ] ws-gateway：文字模式实时对话（心跳 / 断线恢复）
- [ ] asynq 异步复盘报告
- [ ] docker compose 一键起全套

## License

[MIT](LICENSE)
