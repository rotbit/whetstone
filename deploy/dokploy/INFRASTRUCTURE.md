# Dokploy 基础设施

生产环境的业务服务和基础设施位于同一个 Dokploy Project / Environment，并通过 `dokploy-network` 内网通信。MySQL、Redis、Qdrant 和 MinIO 均不发布宿主机端口，也不创建公网域名。

## 已创建资源

| 用途 | Dokploy 资源 | 镜像 | 业务容器内地址 | 持久化 |
|---|---|---|---|---|
| 关系数据库 | Database `mysql` | `mysql:8.4` | `whetstone-mysql-5rfruq:3306` | Dokploy 数据库卷 |
| 缓存 / asynq | Application `redis` | `redis:7-alpine` | `whetstone-redis-1p83wk:6379` | `whetstone-redis-data:/data`，AOF 已开启 |
| 向量数据库 | Compose `qdrant` | `qdrant/qdrant:v1.18.3` | `http://whetstone-qdrant:6333` | `qdrant_data:/qdrant/storage` |
| 对象存储 | Compose `minio` | `pgsty/minio:RELEASE.2026-06-18T00-00-00Z` | `http://whetstone-minio:9000` | `minio-data:/data` |

`whetstone-qdrant` 和 `whetstone-minio` 是显式配置的共享网络别名，不随 Compose 容器名变化。MySQL 和 Redis 的地址由 Dokploy 生成；如果删除后重新创建资源，需要从 Dokploy 页面复制新地址并更新业务服务环境变量。

Qdrant 和 MinIO 在 Dokploy 中使用的非敏感 Compose 结构分别保存在 [`qdrant.compose.yml`](qdrant.compose.yml) 和 [`minio.compose.yml`](minio.compose.yml)。密钥仍只在 Dokploy Environment 中维护。

## 初始化状态

- MySQL 已创建 `whetstone` 数据库，并导入 [`sql/schema.sql`](../../sql/schema.sql)；当前包含 10 张业务表。
- Redis 已开启密码认证和 AOF；已通过写入、重载、读取验证持久化。
- Qdrant 已启用 API Key，并通过认证接口检查。
- MinIO 已创建默认私有 Bucket：`whetstone`。
- 已从 `app-apis` 容器验证四个地址均可通过 Dokploy 内网访问。

## 密钥放在哪里

真实密码只保存在 Dokploy，不写入 Git、Dockerfile 或 YAML 模板：

| 资源 | Dokploy 中保存的密钥 |
|---|---|
| `mysql` | 数据库用户密码、root 密码 |
| `redis` | Application 的 `REDIS_PASSWORD` |
| `qdrant` | Compose 的 `QDRANT_API_KEY` |
| `minio` | Compose 的 `MINIO_ROOT_USER`、`MINIO_ROOT_PASSWORD` |

业务代码接入数据库后，在真正需要访问该资源的 Application 中注入以下变量。不要把所有密钥无差别配置到每个服务：

```dotenv
MYSQL_HOST=whetstone-mysql-5rfruq
MYSQL_PORT=3306
MYSQL_DATABASE=whetstone
MYSQL_USER=whetstone
MYSQL_PASSWORD=<从 Dokploy mysql 资源复制>

REDIS_ADDR=whetstone-redis-1p83wk:6379
REDIS_PASSWORD=<从 Dokploy redis Environment 复制>

QDRANT_URL=http://whetstone-qdrant:6333
QDRANT_API_KEY=<从 Dokploy qdrant Environment 复制>

MINIO_ENDPOINT=http://whetstone-minio:9000
MINIO_ACCESS_KEY=<MINIO_ROOT_USER 的值>
MINIO_SECRET_KEY=<MINIO_ROOT_PASSWORD 的值>
MINIO_BUCKET=whetstone
MINIO_USE_SSL=false
```

目前仓库中的业务实现尚未读取这些基础设施变量；资源已经准备好，后续实现存储层时再按服务的最小权限接入。

## 日常操作

- 日志：在对应 Database / Application / Compose 的 `Logs` 页面查看。
- 运行状态：在 `Containers` 和 `Monitoring` 页面检查容器、CPU、内存、磁盘与网络。
- MySQL 结构变更：先提交可重复执行的迁移 SQL，再从受控终端执行；不要直接在生产库手工改表后不留记录。
- 备份：在正式存入业务数据前，为 MySQL 配置定时备份；同时为 Qdrant 和 MinIO 的数据卷制定异机备份策略。
- 对外访问：四项资源只允许应用内网访问。如需临时运维，优先使用 Dokploy Terminal，不要长期发布数据库端口。

生产库已经完成初始化。不要在已有数据的环境中直接重放整份 schema；结构升级应使用单独的版本化迁移文件。
