# Dokploy 五 Application 部署说明

生产环境在同一个 Dokploy Project / Environment 中创建五个独立 **Application**。每个 Application 只运行一个 Go 进程，可以独立部署、回滚、扩容、查看日志和限制资源。

MySQL、Redis、Qdrant、MinIO 的生产资源、内部地址、密钥位置和备份要求见 [INFRASTRUCTURE.md](INFRASTRUCTURE.md)。

| Application | Docker Build Stage | 内部端口 | 公网域名 |
|---|---|---:|---|
| `app-apis` | `app-apis` | 8888 | 需要 |
| `user-rpc` | `user-rpc` | 9001 | 不需要 |
| `interview-rpc` | `interview-rpc` | 9002 | 不需要 |
| `question-rpc` | `question-rpc` | 9003 | 不需要 |
| `report-worker` | `report-worker` | - | 不需要 |

根目录 `Dockerfile` 为五个服务提供同名的命名阶段。Dokploy 不需要 Docker build argument，只需要为每个 Application 选择对应的 `Docker Build Stage`。

## 1. 创建 Application

五个 Application 使用相同的代码源：

```text
Repository URL: https://github.com/rotbit/whetstone.git
Branch: master
Build Path: /
Build Type: Dockerfile
Dockerfile Path: Dockerfile
Docker Context Path: .
```

每个 Application 的 `Docker Build Stage` 按上表填写。先部署三个 RPC 和 `report-worker`，最后部署 `app-apis`。

Dokploy 会为每个 Application 生成唯一的内部服务名，例如：

```text
whetstone-user-rpc-xxxxxx
```

五个 Application 默认连接到 Dokploy 的共享网络。`app-apis` 使用这三个实际服务名访问 RPC：

```text
USER_RPC_ENDPOINT=whetstone-user-rpc-xxxxxx:9001
INTERVIEW_RPC_ENDPOINT=whetstone-interview-rpc-xxxxxx:9002
QUESTION_RPC_ENDPOINT=whetstone-question-rpc-xxxxxx:9003
```

不要在 Advanced / Ports 中发布 RPC 端口，也不要给 RPC 创建公网域名。

## 2. 集中管理配置

推荐在 Dokploy Environment 级别维护可复用配置和密钥：

```dotenv
WHETSTONE_AUTH_ACCESS_SECRET=请替换为至少32位随机字符串
WHETSTONE_AUTH_ACCESS_EXPIRE=604800
WHETSTONE_WEBSOCKET_PUBLIC_URL=wss://api.example.com/ws
WHETSTONE_USER_RPC_ENDPOINT=whetstone-user-rpc-xxxxxx:9001
WHETSTONE_INTERVIEW_RPC_ENDPOINT=whetstone-interview-rpc-xxxxxx:9002
WHETSTONE_QUESTION_RPC_ENDPOINT=whetstone-question-rpc-xxxxxx:9003
WHETSTONE_LOG_LEVEL=info
WHETSTONE_LOG_STAT=false
WHETSTONE_TZ=Asia/Shanghai
```

`app-apis` 的 Environment：

```dotenv
CONFIG_SOURCE=env
AUTH_ACCESS_SECRET=${{environment.WHETSTONE_AUTH_ACCESS_SECRET}}
AUTH_ACCESS_EXPIRE=${{environment.WHETSTONE_AUTH_ACCESS_EXPIRE}}
WEBSOCKET_PUBLIC_URL=${{environment.WHETSTONE_WEBSOCKET_PUBLIC_URL}}
USER_RPC_ENDPOINT=${{environment.WHETSTONE_USER_RPC_ENDPOINT}}
INTERVIEW_RPC_ENDPOINT=${{environment.WHETSTONE_INTERVIEW_RPC_ENDPOINT}}
QUESTION_RPC_ENDPOINT=${{environment.WHETSTONE_QUESTION_RPC_ENDPOINT}}
LOG_LEVEL=${{environment.WHETSTONE_LOG_LEVEL}}
LOG_STAT=${{environment.WHETSTONE_LOG_STAT}}
TZ=${{environment.WHETSTONE_TZ}}
```

三个 RPC 和 `report-worker` 的 Environment：

```dotenv
CONFIG_SOURCE=env
LOG_LEVEL=${{environment.WHETSTONE_LOG_LEVEL}}
LOG_STAT=${{environment.WHETSTONE_LOG_STAT}}
TZ=${{environment.WHETSTONE_TZ}}
```

数据库和对象存储凭据按最小权限配置到真正需要它们的 Application，不要把全套基础设施密钥复制给所有服务。变量名称和当前生产内网地址见 [INFRASTRUCTURE.md](INFRASTRUCTURE.md)。

镜像中的 YAML 是结构模板，go-zero 启动时用 Environment Variables 展开 `${VARIABLE}`。真实密钥不提交到 Git，也不通过镜像构建参数传递。

如果以后需要远程维护整份 YAML，可以在对应 Application 的 Advanced / Volumes 中创建 File Mount，把文件只读挂载到 `/etc/whetstone/<service>.yaml`，并把 `CONFIG_SOURCE` 改为 `file`。日常配置仍建议使用 Environment Variables，结构复杂且非敏感的配置才使用 File Mount。

## 3. 域名和 HTTPS

只在 `app-apis` 的 Domains 页面添加域名：

```text
Container Port: 8888
Path: /
HTTPS: 开启
Certificate: Let's Encrypt
```

Domains 中的 Container Port 只供 Traefik 内部转发，不会把 `8888` 直接发布到公网，因此无需在 Advanced / Ports 中再添加端口。

发布后检查：

```bash
curl https://api.example.com/healthz
```

预期响应：

```json
{"status":"ok"}
```

## 4. 日志、监控和发布

每个 Application 都有独立的页面：

- `Logs`：查看当前服务 stdout/stderr 的 JSON 日志。
- `Deployments`：查看构建和发布过程，定位构建失败。
- `Monitoring`：分别查看 CPU、内存、磁盘和网络。
- `Advanced / Resources`：为单个服务设置 CPU 和内存限制。
- `General / Deploy`：只重新发布当前服务。

日志由 go-zero 和 Worker 直接写 stdout/stderr，不在容器内部保存日志文件，因此 Dokploy 可以实时查看并负责容器日志轮转。

## 5. 旧 Compose

`deploy/dokploy/docker-compose.yml` 暂时保留为迁移回滚方案，不再作为生产环境的主部署入口。五个 Application 全部验证通过后，应在 Dokploy 中停用旧 Compose，避免重复运行和资源浪费。

## 官方资料

- [Dokploy Applications](https://docs.dokploy.com/docs/core/applications)
- [Dockerfile Build Type](https://docs.dokploy.com/docs/core/applications/build-type)
- [Environment Variables](https://docs.dokploy.com/docs/core/variables)
- [Domains](https://docs.dokploy.com/docs/core/domains)
- [Application Advanced Settings](https://docs.dokploy.com/docs/core/applications/advanced)
