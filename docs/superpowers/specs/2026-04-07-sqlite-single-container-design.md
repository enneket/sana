# SQLite 单容器部署设计

## 背景

当前 Sana 使用 Go 服务 + PostgreSQL 双容器部署，docker-compose 管理两个服务。用户希望简化为单一容器，使用 SQLite 替代 PostgreSQL，实现一个 docker 容器部署。

## 架构变更

| Before | After |
|--------|-------|
| Go 服务 + PostgreSQL（双容器） | Go 服务 + SQLite（单容器） |
| `db:5432` 内网通信 | 本地文件 `/data/sana.db` |
| docker-compose 编排两个服务 | docker-compose 单服务 |

## 改动清单

### 1. 数据库层 (`backend/db.go`)

**驱动变更**
- `github.com/jackc/pgx/v5/pgxpool` → `modernc.org/sqlite`
- `pgxpool.Pool` → `sql.DB`

**环境变量**
- `DATABASE_URL` → `SQLITE_PATH`（默认 `/data/sana.db`）

**Schema SQL 语法调整**
- `SERIAL PRIMARY KEY` → `INTEGER PRIMARY KEY AUTOINCREMENT`
- `$1, $2` 参数占位符 → `?`
- `TIMESTAMPTZ` → `DATETIME`
- `to_tsvector` 全文搜索索引移除（SQLite 不支持）

**Schema**
```sql
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sanas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sanas_user_id ON sanas(user_id);
CREATE INDEX IF NOT EXISTS idx_sanas_updated_at ON sanas(updated_at DESC);
```

### 2. 认证/用户代码 (`backend/main.go`)

- `initDB()`、`closeDB()` 签名不变，调用层无需修改
- 用户表操作保持兼容（bcrypt password hash 存储格式不变）
- 全文搜索 `handleSearchMemos` 需降级为 `LIKE %query%` 搜索

### 3. Docker 多阶段构建 (`docker/Dockerfile`)

```dockerfile
# Stage 1: builder
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o sana .

# Stage 2: runtime
FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /build/sana ./sana
COPY frontend/dist/ ./frontend/dist/
EXPOSE 8080
CMD ["./sana"]
```

### 4. Docker Compose 单服务 (`docker/docker-compose.yml`)

```yaml
services:
  sana:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    ports:
      - "5560:8080"
    environment:
      - JWT_SECRET=${JWT_SECRET:?JWT_SECRET is required}
      - SANA_PASSWORD=${SANA_PASSWORD:?SANA_PASSWORD is required}
      - SQLITE_PATH=/data/sana.db
      - PORT=8080
    volumes:
      - ${SANA_DATA_DIR:-./sana_data}:/data
```

**注意**：`db` 服务和 `postgres_data` volume 移除。

### 5. 数据持久化

- 宿主机目录挂载：`/opt/sana:/data`（或 `${SANA_DATA_DIR}`）
- SQLite 文件：`/data/sana.db`
- 备份方式：`cp /opt/sana/sana.db /backup/`

## 部署变更

```bash
# 创建数据目录
mkdir -p /opt/sana

# 启动
SANA_DATA_DIR=/opt/sana docker compose -f docker/docker-compose.yml up -d
```

## 已知限制

1. **全文搜索降级**：PostgreSQL `to_tsvector` 全文搜索 → SQLite `LIKE` 模糊搜索
2. **并发写入**：SQLite 为单写模型，高并发写入场景不适用（本项目为单用户，影响可忽略）
3. **数据迁移**：无自动迁移工具，schema 变更需手动处理

## 实现顺序

1. 修改 `backend/db.go` 支持 SQLite
2. 更新 `backend/go.mod` 依赖
3. 更新 `docker/Dockerfile` 多阶段构建
4. 简化 `docker/docker-compose.yml`
5. 本地测试验证
6. 更新 README 部署文档（如有）
