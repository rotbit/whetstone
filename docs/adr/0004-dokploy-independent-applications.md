# ADR-0004：Dokploy 使用五个独立 Application

## 状态

Accepted

## 背景

ADR-0003 把五个进程拆为五个容器，但仍由一个 Docker Compose 统一发布。这样无法独立发布和回滚单个服务，也不符合每个服务在 Dokploy 中拥有独立 Application 的运维习惯。

## 决策

- 在同一个 Dokploy Project / Environment 中创建 `app-apis`、三个 zRPC 和 `report-worker` 五个独立 Application。
- 根目录 Dockerfile 提供五个固定命名阶段；每个 Application 使用同名 `Docker Build Stage`，不依赖 build argument。
- 五个服务均使用 Dokploy 共享内部网络。`app-apis` 通过 Dokploy 生成的 Application 服务名访问三个 RPC。
- 仅为 `app-apis:8888` 配置 Traefik 域名和 HTTPS；RPC 不发布主机端口，也不创建公网域名。
- 密钥和跨服务公共配置放在 Dokploy Environment Variables 中，各 Application 通过引用使用。
- 每个 Application 独立输出 stdout/stderr 日志、查看监控、设置资源限制和执行发布。

## 后果

### 正面

- 每个服务可以独立构建、发布、回滚、重启和扩容。
- 日志、部署记录、资源监控与告警边界完全按服务隔离。
- 单个 RPC 或 Worker 的故障与更新不再要求重新发布全部服务。
- 命名 Docker 阶段让 Dokploy 配置显式且可验证，避免构建参数遗漏。

### 负面

- 首次迁移需要创建和配置五个 Application。
- `app-apis` 必须记录 Dokploy 为三个 RPC 生成的内部服务名。
- 跨服务配置和发布顺序需要在 Environment Variables 与发布流程中维护。

### 中性

- 容器数量仍为五个，和 ADR-0003 相同；拆分 Application 主要增加控制面对象，不会额外启动业务容器。
- `deploy/dokploy/docker-compose.yml` 在迁移期保留为回滚方案，验证后不再作为生产入口。

## 参考

- https://docs.dokploy.com/docs/core/applications
- https://docs.dokploy.com/docs/core/applications/build-type
- https://docs.dokploy.com/docs/core/variables
- `deploy/dokploy/README.md`
