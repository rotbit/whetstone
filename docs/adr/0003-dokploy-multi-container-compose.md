# ADR-0003：Dokploy 使用五容器 Compose 部署

## 状态

Superseded by ADR-0004

## 背景

原方案把 `app-apis`、三个 zRPC 和 `report-worker` 作为五个进程放入一个容器，由 Supervisor 管理。该方式构建简单，但运行日志、资源统计、故障边界和重启操作都只能以整个容器为单位。

项目决定接受少量容器管理开销，换取按服务观察和控制的能力。同时需要避免五个独立 Dokploy Application 之间依赖自动生成的 Swarm 服务名。

## 决策

- 使用一个 Dokploy Docker Compose 项目部署五个容器，每个容器只运行一个 Go 进程。
- 复用根目录 Dockerfile，通过构建参数 `SERVICE` 选择需要编译的入口；每个运行镜像只包含对应的一个二进制。
- `app-apis` 通过 Compose DNS 连接 `user-rpc:9001`、`interview-rpc:9002` 和 `question-rpc:9003`。
- 只有 `app-apis` 加入 `dokploy-network` 并绑定公网域名；RPC 只加入项目内部网络，不发布主机端口。
- 继续使用 Dokploy Environment Variables 展开 go-zero YAML，真实密钥不进入 Git、构建参数或镜像层。
- 每个服务向 stdout/stderr 输出带 `service` 字段的 JSON 日志，由 Dokploy 分服务展示。

## 后果

### 正面

- 日志、资源统计、重启和故障状态可以按服务查看。
- RPC 使用稳定的 Compose 服务名，不依赖 Dokploy 生成的容器或 Swarm 名称。
- 不再需要 Supervisor 和 Python 运行时，单个服务镜像更小。
- 后续可以对网关、RPC 或 Worker 分别设置资源限制和副本数。

### 负面

- 一次部署需要构建五个服务镜像，首次构建时间会增加。
- 容器数量从一个增加到五个，带来少量 namespace、cgroup 和日志管理开销。
- Compose 的 `depends_on` 只保证启动顺序，不保证 RPC 已完成就绪；网关需要依靠重启策略处理首次连接失败。

### 中性

- 仍由一个 Compose 项目统一部署和回滚，暂不追求完全独立的发布流水线。
- MySQL、Redis、Qdrant 和 MinIO 继续作为独立基础设施服务运行。

## 备选方案

**一个容器运行五个进程**

- 资源开销略低，但日志、监控和故障隔离较差，因此被本决策替代。

**五个独立 Dokploy Application**

- 发布自治能力更强，但内部 DNS 名称和部署顺序管理更复杂；当前阶段选择一个 Compose 项目。

**把 RPC 暴露成公网域名**

- 服务发现直观，但扩大攻击面并增加 TLS 与鉴权成本，因此不采用。

## 参考

- https://docs.dokploy.com/docs/core/docker-compose
- https://docs.dokploy.com/docs/core/docker-compose/domains
- `deploy/dokploy/docker-compose.yml`
