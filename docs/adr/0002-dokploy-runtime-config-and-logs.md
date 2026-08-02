# ADR-0002：Dokploy 运行时配置与容器日志策略

## 状态

Superseded by ADR-0003

## 背景

Whetstone 以一个 Dokploy Application 交付，但容器内包含五个进程。项目配置采用 go-zero YAML，生产密钥不能固化在 Git 或镜像中；同时 Dokploy 需要从容器标准输出收集所有进程的日志。

## 决策

- 默认使用 Dokploy Environment Variables 管理环境差异和敏感值。
- 镜像只携带不含真实密钥的 YAML 模板，go-zero 通过 `conf.UseEnv()` 在启动时展开环境变量。
- 对需要整份远程维护的 YAML，允许使用 Dokploy File Mount 覆盖 `/etc/whetstone/*.yaml`。
- Supervisor 自身及全部子进程只写 stdout/stderr，不在容器文件系统中保存业务日志。
- go-zero 日志使用 JSON 编码并追加 `service` 字段，便于在单一 Dokploy Logs 视图中识别来源。
- Dokploy 只暴露 `app-apis` 的 `8888`；RPC 继续使用容器回环地址。

## 后果

### 正面

- 密钥不进入 Git、Docker build argument 或镜像层。
- 环境变量可在项目级、环境级和服务级复用，更新路径明确。
- File Mount 保留了整份 YAML 远程配置的能力。
- Dokploy Logs 能直接看到五个进程及 Supervisor 的生命周期日志。

### 负面

- 五个进程的日志仍共享一个容器日志流，只能依靠 `service` 字段区分。
- 配置修改需要重新部署或重启容器才能生效。
- File Mount 中保存完整 YAML 时，配置结构和应用版本需要人工保持兼容。

### 中性

- 单容器 Monitoring 展示的是整个容器资源，而不是每个 Go 进程的独立资源。
- 将来拆成多个 Dokploy Application 时，可以继续使用相同的环境变量和 JSON 日志约定。

## 备选方案

**把生产 YAML 直接提交到仓库**

- 操作最简单，但容易泄漏密钥并造成环境配置混用，因此不采用。

**只使用 Dokploy File Mount**

- 适合复杂配置，但密钥复用和分环境管理不如 Environment Variables，因此作为可选覆盖方式。

**让服务写容器内日志文件**

- 会绕过 Dokploy Logs，并需要额外处理持久化和轮转，因此不采用。

## 参考

- https://docs.dokploy.com/docs/core/applications
- https://docs.dokploy.com/docs/core/variables
- https://docs.dokploy.com/docs/core/troubleshooting
