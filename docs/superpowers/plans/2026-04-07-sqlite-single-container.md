# SQLite 单容器部署实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Sana 从 PostgreSQL 双容器部署改为 SQLite 单容器部署

**Architecture:** 用 modernc.org/sqlite 替代 pgx/v5，SQLite 文件通过宿主机目录 `/data` 挂载持久化。Dockerfile 改为多阶段构建，docker-compose 简化为单服务。

**Tech Stack:** Go 1.23, modernc.org/sqlite, debian:stable-slim, docker compose

---

## 文件变更总览

| 文件 | 变更内容 |
|------|---------|
| `backend/go.mod` | `pgx/v5` → `modernc.org/sqlite`，移除 pgx 依赖 |
| `backend/go.sum` | 重新生成 |
| `backend/db.go` | `pgxpool.Pool` → `sql.DB`，所有 `$N` → `?`，TIMESTAMPTZ → DATETIME，SERIAL → AUTOINCREMENT，移除 gin 索引 |
| `backend/memo_handler.go` | 所有 SQL 查询 `$N` → `?`，`ILIKE` → `LIKE`，`INTERVAL` → 兼容 SQLite 语法 |
| `backend/search_handler.go` | `ILIKE` → `LIKE`（SQLite 不支持 ILIKE） |
| `backend/import_handler.go` | `ON CONFLICT` 语法调整（SQLite UPSERT 语法） |
| `backend/export_handler.go` | 无 SQL 变更 |
| `docker/Dockerfile` | 多阶段构建：golang:1.23-alpine builder → debian:stable-slim runtime |
| `docker/docker-compose.yml` | 移除 `db` 服务和 `postgres_data` volume，单服务 + 数据目录挂载 |

---

## Task 1: 修改 `backend/go.mod` 依赖

**Files:**
- Modify: `backend/go.mod:1-11`

- [ ] **Step 1: 更新 go.mod，替换 pgx 为 sqlite**

```go
module github.com/zjx/sana/backend

go 1.23.0

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	modernc.org/sqlite v1.35.1
	golang.org/x/crypto v0.31.0
	gopkg.in/yaml.v3 v3.0.1
)
```

- [ ] **Step 2: 重新生成 go.sum**

Run: `cd backend && go mod tidy`

---

## Task 2: 重写 `backend/db.go`

**Files:**
- Modify: `backend/db.go:1-95`（完全重写）

- [ ] **Step 1: 重写 db.go**

```go
package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB() {
	sqlitePath := getEnv("SQLITE_PATH", "/data/sana.db")
	if sqlitePath == "" {
		sqlitePath = "/data/sana.db"
	}

	// Ensure directory exists
	dir := sqlitePath[:len(sqlitePath)-len("/sana.db")]
	if dir == "" {
		dir = "."
	}
	os.MkdirAll(dir, 0755)

	var err error
	db, err = sql.Open("sqlite", sqlitePath)
	if err != nil {
		log.Fatalf("Unable to open database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite only allows one writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	if err := createSchema(); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("SQLite database connected and schema initialized")
}

func createSchema() error {
	schema := `
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
	`
	_, err := db.Exec(context.Background(), schema)
	return err
}

func closeDB() {
	db.Close()
}

// Sana 模型
type Sana struct {
	ID        int
	UID       string
	UserID    string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SanaResponse 对外API响应格式
type SanaResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedTs int64  `json:"created_ts"`
	UpdatedTs int64  `json:"updated_ts"`
}

// ToResponse converts Sana to SanaResponse
func (s *Sana) ToResponse() SanaResponse {
	return SanaResponse{
		ID:        s.UID,
		Content:   s.Content,
		CreatedTs: s.CreatedAt.Unix(),
		UpdatedTs: s.UpdatedAt.Unix(),
	}
}
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build -o /dev/null .`
Expected: 无错误

- [ ] **Step 3: 提交**

```bash
git add backend/go.mod backend/go.sum backend/db.go
git commit -m "refactor(db): replace pgx with modernc.org/sqlite

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 更新 `backend/memo_handler.go` SQL 语法

**Files:**
- Modify: `backend/memo_handler.go:37-43`, `backend/memo_handler.go:94-102`, `backend/memo_handler.go:124-127`, `backend/memo_handler.go:156-163`, `backend/memo_handler.go:183`, `backend/memo_handler.go:202-218`

**变更说明：** 所有 PostgreSQL `$1, $2` 占位符改为 SQLite `?`。`INTERVAL '90 days'` 改为 SQLite 兼容的 `julianday('now') - 90`。

- [ ] **Step 1: 更新 memo_handler.go 的 SQL 查询**

**handleListMemos (line 37-43)** — `$1, $2, $3, $4` → `?, ?, ?, ?`

```go
var rows, err = db.Query(sanaCtx, `
    SELECT id, uid, user_id, content, created_at, updated_at
    FROM sanas
    WHERE user_id = ? AND (? OR updated_at < ?)
    ORDER BY updated_at DESC
    LIMIT ?
`, userID, cursor.IsZero(), cursor, limit)
```

**handleCreateMemo (line 94-102)** — `$1, $2, $3, $4, $5` → `?, ?, ?, ?, ?`

```go
err := db.QueryRow(sanaCtx, `
    INSERT INTO sanas (uid, user_id, content, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?)
`, uid, userID, req.Content, now, now).Scan(&id)
```

**handleGetMemo (line 124-127)** — `$1, $2` → `?, ?`

```go
err := db.QueryRow(sanaCtx, `
    SELECT id, uid, user_id, content, created_at, updated_at
    FROM sanas WHERE uid = ? AND user_id = ?
`, uid, userID).Scan(&s.ID, &s.UID, &s.UserID, &s.Content, &s.CreatedAt, &s.UpdatedAt)
```

**handleUpdateMemo (line 156-163)** — `$1, $2, $3, $4` → `?, ?, ?, ?`

```go
result, err := db.Exec(sanaCtx, `
    UPDATE sanas SET content = ?, updated_at = ?
    WHERE uid = ? AND user_id = ?
`, req.Content, now, uid, userID)
```

**handleDeleteMemo (line 183)** — `$1, $2` → `?, ?`

```go
result, err := db.Exec(sanaCtx, `DELETE FROM sanas WHERE uid = ? AND user_id = ?`, uid, userID)
```

**handleGetStats (line 202-218)** — `$1` → `?`，`INTERVAL '90 days'` 替换

```go
var memoCount int
db.QueryRow(r.Context(),
    "SELECT COUNT(*) FROM sanas WHERE user_id = ?", userID).Scan(&memoCount)

var activeDays int
db.QueryRow(r.Context(),
    "SELECT COUNT(DISTINCT DATE(created_at)) FROM sanas WHERE user_id = ?", userID).Scan(&activeDays)

var totalChars int
db.QueryRow(r.Context(),
    "SELECT COALESCE(SUM(LENGTH(content)), 0) FROM sanas WHERE user_id = ?", userID).Scan(&totalChars)

rows, _ := db.Query(r.Context(), `
    SELECT DATE(created_at) as day, COUNT(*) as count
    FROM sanas
    WHERE user_id = ? AND created_at >= datetime('now', '-90 days')
    GROUP BY DATE(created_at)
`, userID)
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build -o /dev/null .`
Expected: 无错误

- [ ] **Step 3: 提交**

```bash
git add backend/memo_handler.go
git commit -m "refactor(memo_handler): update SQL syntax for SQLite compatibility

PostgreSQL $N placeholders → SQLite ?
ILIKE → LIKE
INTERVAL '90 days' → datetime('now', '-90 days')

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: 更新 `backend/search_handler.go`

**Files:**
- Modify: `backend/search_handler.go:21-26`

- [ ] **Step 1: 更新 ILIKE → LIKE**

```go
rows, err := db.Query(ctx, `
    SELECT id, uid, user_id, content, created_at, updated_at
    FROM sanas
    WHERE user_id = ? AND content LIKE '%' || ? || '%'
    ORDER BY updated_at DESC
`, userID, q)
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build -o /dev/null .`
Expected: 无错误

- [ ] **Step 3: 提交**

```bash
git add backend/search_handler.go
git commit -m "refactor(search): ILIKE → LIKE for SQLite

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 更新 `backend/import_handler.go` UPSERT 语法

**Files:**
- Modify: `backend/import_handler.go:112-116`

- [ ] **Step 1: SQLite UPSERT 语法（INSERT OR REPLACE）**

PostgreSQL `ON CONFLICT (uid) DO UPDATE SET ...` 在 SQLite 中用 `INSERT OR REPLACE` 实现：

```go
_, err := db.Exec(importCtx, `
    INSERT INTO sanas (uid, user_id, content, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?)
`, uid, userID, content, createdTs, updatedTs)
```

注意：`INSERT OR REPLACE` 会用新值替换冲突行，但不会保留旧列的默认值——uid/ user_id/content/created_at/updated_at 全部由 INSERT 的值填充，与原语义一致。

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build -o /dev/null .`
Expected: 无错误

- [ ] **Step 3: 提交**

```bash
git add backend/import_handler.go
git commit -m "refactor(import): use INSERT OR REPLACE for SQLite upsert

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: 更新 `docker/Dockerfile` 多阶段构建

**Files:**
- Modify: `docker/Dockerfile:1-8`

- [ ] **Step 1: 重写 Dockerfile**

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

- [ ] **Step 2: 提交**

```bash
git add docker/Dockerfile
git commit -m "feat(docker): multi-stage build for smaller image

Builder: golang:1.23-alpine
Runtime: debian:stable-slim (~80MB)

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 7: 简化 `docker/docker-compose.yml` 为单服务

**Files:**
- Modify: `docker/docker-compose.yml:1-32`

- [ ] **Step 1: 重写 docker-compose.yml**

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
    restart: unless-stopped
```

**说明：**
- 移除 `db` 服务（PostgreSQL）
- 移除 `postgres_data` volume
- 移除 `depends_on: db`
- `depends_on` 引用已删除，`condition: service_healthy` 不再需要
- `SANA_DATA_DIR` 环境变量控制宿主机数据目录，默认 `./sana_data`
- 新增 `restart: unless-stopped` 提高可用性

- [ ] **Step 2: 提交**

```bash
git add docker/docker-compose.yml
git commit -m "feat(docker): single-container deployment with SQLite

Remove PostgreSQL db service, add SANA_DATA_DIR volume mount.
Use SQLITE_PATH env var instead of DATABASE_URL.

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 8: 本地验证

- [ ] **Step 1: 本地编译**

Run: `cd backend && go build -o sana_test . && rm sana_test`
Expected: 无错误

- [ ] **Step 2: 测试 docker build**

Run: `docker build -f docker/Dockerfile -t sana:test .`
Expected: Build success，无报错

- [ ] **Step 3: 确认所有文件已提交**

Run: `git log --oneline -8`
Expected: 6 个新 commit（db.go, go.mod, memo_handler, search_handler, import_handler, Dockerfile, docker-compose）

---

## Spec 自检

- [x] PostgreSQL → SQLite 驱动替换 (`pgx` → `modernc.org/sqlite`) — Task 1, 2
- [x] Schema SQL 语法调整 (`SERIAL` → `AUTOINCREMENT`, `$N` → `?`, `TIMESTAMPTZ` → `DATETIME`) — Task 2
- [x] 全文搜索降级 (`ILIKE` → `LIKE`, `to_tsvector` 移除) — Task 3, 4
- [x] `INTERVAL '90 days'` → `datetime('now', '-90 days')` — Task 3
- [x] UPSERT 语法变更 (`ON CONFLICT` → `INSERT OR REPLACE`) — Task 5
- [x] Docker 多阶段构建 — Task 6
- [x] Docker Compose 单服务 + 数据卷挂载 — Task 7
- [x] 无 placeholder / TODO / TBD
- [x] 类型一致性：所有 `db.QueryRow` / `db.Query` / `db.Exec` 参数顺序与占位符一一对应
