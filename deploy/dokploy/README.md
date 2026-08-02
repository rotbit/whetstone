# Dokploy 多容器部署说明

本项目在 Dokploy 中按一个 **Docker Compose** 项目部署，包含五个容器，每个容器只运行一个 Go 进程：

| Compose 服务 | 内部端口 | 对外发布 | 职责 |
|---|---:|---|---|
| `app-apis` | 8888 | 是 | REST / WebSocket 网关 |
| `user-rpc` | 9001 | 否 | 用户、次数和订单 |
| `interview-rpc` | 9002 | 否 | 面试引擎 |
| `question-rpc` | 9003 | 否 | 题库和 RAG |
| `report-worker` | - | 否 | 异步报告任务 |

`app-apis` 通过 Compose DNS 访问 `user-rpc:9001`、`interview-rpc:9002` 和 `question-rpc:9003`。不需要为 RPC 创建公网域名或发布主机端口。

## 1. 创建 Dokploy Compose

1. 在 Dokploy 中创建 Project 和 Environment。
2. 创建 `Docker Compose` 服务，Compose Type 选择 **Docker Compose**，不要选择 Stack。
3. Source 选择 GitHub 或 Git，配置仓库和部署分支。
4. Compose Path 填 `./deploy/dokploy/docker-compose.yml`。
5. 保存下面的环境变量，然后点击 Deploy。
6. 在 Domains 中为 `app-apis` 添加域名，Container Port 填 `8888`。

选择 Docker Compose 是因为当前配置使用 `build` 从同一仓库生成五个服务镜像；Dokploy 的 Stack 模式不支持 `build`。

## 2. 环境变量

在 Compose 的 Environment 页面配置：

```dotenv
AUTH_ACCESS_SECRET=请替换为至少32位随机字符串
AUTH_ACCESS_EXPIRE=604800
WEBSOCKET_PUBLIC_URL=wss://api.example.com/ws
LOG_LEVEL=info
LOG_STAT=false
TZ=Asia/Shanghai
```

可以用下面的命令生成 JWT 密钥：

```bash
openssl rand -hex 32
```

Dokploy 会把 Compose Environment 保存为 `.env`；`docker-compose.yml` 已显式使用 `${VARIABLE}` 将需要的变量注入相应容器。

如果希望跨服务或跨环境复用密钥，可以先定义 Environment-level Variable：

```dotenv
WHETSTONE_AUTH_ACCESS_SECRET=真实密钥
```

再在 Compose Environment 中引用：

```dotenv
AUTH_ACCESS_SECRET=${{environment.WHETSTONE_AUTH_ACCESS_SECRET}}
```

不要把真实密钥提交到 Git，也不要通过 Docker build argument 传递密钥。修改运行配置后重新 Deploy。

## 3. YAML 配置管理

镜像中的生产 YAML 是模板，go-zero 启动时会展开运行环境中的 `${VARIABLE}`。默认情况下只需要维护 Dokploy Environment，不需要远程挂载 YAML。

配置文件在各容器中的路径为：

```text
/etc/whetstone/app-apis.yaml
/etc/whetstone/user.yaml
/etc/whetstone/interview.yaml
/etc/whetstone/question.yaml
```

如需整份 YAML 远程覆盖，可以使用 Dokploy File Mount。对于 Compose，文件会保存在 Dokploy 为该项目维护的 `files` 目录中，需要在 Compose 中为对应服务增加只读挂载，例如：

```yaml
services:
  app-apis:
    environment:
      CONFIG_SOURCE: file
    volumes:
      - ../../../files/app-apis.yaml:/etc/whetstone/app-apis.yaml:ro
```

这里使用 `../../../files`，是因为 Compose 文件位于仓库的 `deploy/dokploy` 子目录，而 Dokploy 的持久化 `files` 目录位于仓库 `code` 目录的同级。

密钥仍建议放 Environment Variables；File Mount 更适合结构复杂的非敏感配置。

## 4. 日志和资源

进入 Compose 的 Logs 页面，可以按服务选择：

```text
app-apis
user-rpc
interview-rpc
question-rpc
report-worker
```

每个容器只包含一个进程，日志不再混合。go-zero 和 Worker 都输出 JSON，并包含 `service` 字段。构建失败或发布失败则进入 Deployments 查看对应部署记录。

Monitoring 也可以分别查看五个容器的 CPU、内存、磁盘和网络；后续可对单个服务设置资源限制或独立扩容。

## 5. 发布顺序和健康检查

Compose 会先创建三个 RPC，再启动 `app-apis`。如果 RPC 尚未就绪导致网关首次启动失败，`restart: unless-stopped` 会自动重启网关。

Compose 为 `app-apis` 配置了健康检查：

```text
GET http://127.0.0.1:8888/healthz
```

预期响应：

```json
{"status":"ok"}
```

## 官方资料

- [Dokploy Docker Compose](https://docs.dokploy.com/docs/core/docker-compose)
- [Docker Compose Domains](https://docs.dokploy.com/docs/core/docker-compose/domains)
- [Environment Variables](https://docs.dokploy.com/docs/core/variables)
- [Dokploy Troubleshooting / File Mounts](https://docs.dokploy.com/docs/core/troubleshooting)
