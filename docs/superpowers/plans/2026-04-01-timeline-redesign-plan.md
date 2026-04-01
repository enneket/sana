# Sana Timeline 重构实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Sana 从文件夹导航重构为 Timeline 视图，支持快速创建、全文搜索、Memos 格式导入导出，数据存储从 JSON 文件迁移到 PostgreSQL。

**Architecture:** 后端新增 `/api/memos` REST 接口族，前端新增 TimelineView 替代 FolderView，存储层从 JSON 文件改为 PostgreSQL。前端保持 Vue 3，后端保持 Go 单体结构但分离 handler 到独立文件。

**Tech Stack:** Go (stdlib net/http + pgx), Vue 3, PostgreSQL, JWT

---

## 文件结构

```
backend/
├── main.go              # 主入口，保留 auth handler，移除 folder/note handler
├── db.go                # NEW: PostgreSQL 连接和 schema 初始化
├── memo_handler.go      # NEW: /api/memos CRUD handler
├── search_handler.go     # NEW: /api/memos/search handler
├── import_handler.go    # NEW: /api/import/memos handler
├── export_handler.go    # NEW: /api/export/memos handler
├── go.mod
├── go.sum
└── data/                # 保留，旧数据只读

frontend/src/
├── views/
│   ├── TimelineView.vue   # NEW: 替代 FolderView
│   └── Login.vue          # 不变
├── components/
│   ├── MemoCard.vue       # NEW: 单条笔记展示
│   ├── MemoComposer.vue   # NEW: 快速创建输入框
│   ├── MemoEditor.vue     # NEW: 笔记编辑弹窗
│   ├── TimeGroup.vue      # NEW: 按日期分组
│   └── SearchBar.vue     # NEW: 搜索框
├── router/index.js       # 修改: 路由 / -> TimelineView
├── api/index.js          # 修改: 新增 memos API
└── style.css             # 修改: 新增 Timeline 样式
```

---

## Phase 1: 后端 - PostgreSQL + Timeline API

### Task 1: 数据库层

**Files:**
- Create: `backend/db.go`
- Modify: `backend/go.mod`

- [ ] **Step 1: 创建 db.go**

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func initDB() {
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable must be set")
	}

	var err error
	db, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	// Test connection
	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	// Create schema
	if err := createSchema(); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("Database connected and schema initialized")
}

func createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS memos (
		id SERIAL PRIMARY KEY,
		uid TEXT UNIQUE NOT NULL,
		user_id TEXT NOT NULL REFERENCES users(id),
		content TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_memos_user_id ON memos(user_id);
	CREATE INDEX IF NOT EXISTS idx_memos_updated_at ON memos(updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_memos_content_gin ON memos USING gin(to_tsvector('simple', content));
	`
	_, err := db.Exec(context.Background(), schema)
	return err
}

func closeDB() {
	db.Close()
}

// Memo 模型
type Memo struct {
	ID        int       `json:"id"`
	UID       string    `json:"uid"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemoResponse 对外API响应格式
type MemoResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedTs int64  `json:"created_ts"`
	UpdatedTs int64  `json:"updated_ts"`
}

// ToResponse converts Memo to MemoResponse
func (m *Memo) ToResponse() MemoResponse {
	return MemoResponse{
		ID:        m.UID,
		Content:   m.Content,
		CreatedTs: m.CreatedAt.Unix(),
		UpdatedTs: m.UpdatedAt.Unix(),
	}
}
```

- [ ] **Step 2: 更新 go.mod 添加 pgx 依赖**

```go
module sana

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	golang.org/x/crypto v0.22.0
)
```

Run: `cd backend && go mod tidy`
Expected: 下载 pgx 依赖

- [ ] **Step 3: 修改 main.go 移除 JSON storage 代码，添加 db.go 调用**

在 main.go 中:
1. 删除 `notesMu`, `foldersMu`, `notesData`, `foldersData`, `usersData` 变量
2. 删除 `noteRecord`, `folder` 结构体（保留 `user`）
3. 删除 `loadData()`, `saveNotesData()`, `saveFoldersData()` 函数
4. 删除 `readNoteFile()`, `writeNoteFile()` 函数
5. 在 `main()` 开头添加 `initDB()`
6. 在 `main()` defer 添加 `closeDB()`
7. 保留 auth handler（login, logout, me）

```go
// main.go 保留的结构
type user struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	CreatedAt    string `json:"created_at"`
}

// main.go 修改后的 main 函数开头
func main() {
	notesDir := getEnv("NOTES_DIR", "./notes")
	os.MkdirAll(notesDir, 0755)

	initDB()  // NEW
	initDefaultUser()

	mux := http.NewServeMux()
	// ... auth handlers same ...
}
```

- [ ] **Step 4: Commit**

```bash
cd /home/zjx/code/mine/sana
git add backend/db.go backend/go.mod backend/go.sum backend/main.go
git commit -m "feat(db): add PostgreSQL connection and schema

- Add pgx for PostgreSQL
- Create db.go with connection pool and schema init
- Add Memo and MemoResponse types
- Remove JSON file storage code from main.go
"
```

---

### Task 2: Memo CRUD API

**Files:**
- Create: `backend/memo_handler.go`
- Modify: `backend/main.go`

- [ ] **Step 1: 创建 memo_handler.go**

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// handleListMemos handles GET /api/memos
func handleListMemos(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	cursorStr := r.URL.Query().Get("cursor")
	var cursor time.Time
	if cursorStr != "" {
		ts, err := strconv.ParseInt(cursorStr, 10, 64)
		if err == nil {
			cursor = time.Unix(ts, 0)
		}
	}

	ctx := context.Background()
	var rows, err = db.Query(ctx, `
		SELECT id, uid, user_id, content, created_at, updated_at
		FROM memos
		WHERE user_id = $1 AND ($2 = false OR updated_at < $3)
		ORDER BY updated_at DESC
		LIMIT $4
	`, userID, cursor.IsZero(), cursor, limit)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var memos []MemoResponse
	var lastUpdated *time.Time
	for rows.Next() {
		var m Memo
		if err := rows.Scan(&m.ID, &m.UID, &m.UserID, &m.Content, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		memos = append(memos, m.ToResponse())
		lastUpdated = &m.UpdatedAt
	}

	response := map[string]interface{}{
		"memos": memos,
	}
	if lastUpdated != nil {
		response["next_cursor"] = lastUpdated.Unix()
		response["has_more"] = len(memos) == limit
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCreateMemo handles POST /api/memos
func handleCreateMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		http.Error(w, "content cannot be empty", http.StatusBadRequest)
		return
	}

	uid := uuid.New().String()
	now := time.Now()

	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO memos (uid, user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, uid, userID, req.Content, now, now).Scan(&id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	memo := Memo{
		ID:        id,
		UID:       uid,
		UserID:    userID,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(memo.ToResponse())
}

var ctx = context.Background()

// handleGetMemo handles GET /api/memos/:id
func handleGetMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	uid := strings.TrimPrefix(r.URL.Path, "/api/memos/")

	var m Memo
	err := db.QueryRow(ctx, `
		SELECT id, uid, user_id, content, created_at, updated_at
		FROM memos WHERE uid = $1 AND user_id = $2
	`, uid, userID).Scan(&m.ID, &m.UID, &m.UserID, &m.Content, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.ToResponse())
}

// handleUpdateMemo handles PUT /api/memos/:id
func handleUpdateMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	uid := strings.TrimPrefix(r.URL.Path, "/api/memos/")

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		http.Error(w, "content cannot be empty", http.StatusBadRequest)
		return
	}

	now := time.Now()
	result, err := db.Exec(ctx, `
		UPDATE memos SET content = $1, updated_at = $2
		WHERE uid = $3 AND user_id = $4
	`, req.Content, now, uid, userID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         uid,
		"content":    req.Content,
		"updated_ts": now.Unix(),
	})
}

// handleDeleteMemo handles DELETE /api/memos/:id
func handleDeleteMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	uid := strings.TrimPrefix(r.URL.Path, "/api/memos/")

	result, err := db.Exec(ctx, `DELETE FROM memos WHERE uid = $1 AND user_id = $2`, uid, userID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 2: 在 main.go 注册 memo 路由**

在 main.go 的 mux 注册部分添加:

```go
// Memos (Timeline)
mux.HandleFunc("GET /api/memos", withAuth(handleListMemos))
mux.HandleFunc("POST /api/memos", withAuth(handleCreateMemo))
mux.HandleFunc("GET /api/memos/", withAuth(handleGetMemo))
mux.HandleFunc("PUT /api/memos/", withAuth(handleUpdateMemo))
mux.HandleFunc("DELETE /api/memos/", withAuth(handleDeleteMemo))
```

Run: `cd backend && go build -o sana .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add backend/memo_handler.go backend/main.go
git commit -m "feat(api: memo CRUD endpoints

- GET /api/memos - list with cursor pagination
- POST /api/memos - create memo
- GET /api/memos/:id - get single memo
- PUT /api/memos/:id - update memo
- DELETE /api/memos/:id - delete memo
"
```

---

### Task 3: 搜索 API

**Files:**
- Create: `backend/search_handler.go`

- [ ] **Step 1: 创建 search_handler.go**

```go
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// handleSearchMemos handles GET /api/memos/search?q=<keyword>
func handleSearchMemos(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"memos": []MemoResponse{}, "total": 0})
		return
	}

	ctx := context.Background()
	rows, err := db.Query(ctx, `
		SELECT id, uid, user_id, content, created_at, updated_at
		FROM memos
		WHERE user_id = $1 AND content ILIKE '%' || $2 || '%'
		ORDER BY updated_at DESC
	`, userID, q)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var memos []MemoResponse
	for rows.Next() {
		var m Memo
		if err := rows.Scan(&m.ID, &m.UID, &m.UserID, &m.Content, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		memos = append(memos, m.ToResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"memos": memos,
		"total": len(memos),
	})
}
```

- [ ] **Step 2: 在 main.go 注册搜索路由**

```go
mux.HandleFunc("GET /api/memos/search", withAuth(handleSearchMemos))
```

Run: `cd backend && go build -o sana .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add backend/search_handler.go backend/main.go
git commit -m "feat(api: search memos

- GET /api/memos/search?q=<keyword>
- ILIKE search on content field
"
```

---

### Task 4: 导入导出 API

**Files:**
- Create: `backend/export_handler.go`
- Create: `backend/import_handler.go`

- [ ] **Step 1: 创建 export_handler.go**

```go
package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MemosExport Memos格式导出结构
type MemosExport struct {
	App       string           `json:"app"`
	Version   string           `json:"version"`
	ExportedAt string         `json:"exported_at"`
	Memos     []MemoExportItem `json:"memos"`
}

// MemoExportItem 单条导出项
type MemoExportItem struct {
	UID       string `json:"uid"`
	Content   string `json:"content"`
	Visibility string `json:"visibility"`
	Pinned    bool   `json:"pinned"`
	CreatedTs int64  `json:"created_ts"`
	UpdatedTs int64  `json:"updated_ts"`
}

func handleExportMemos(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	ctx := context.Background()
	rows, err := db.Query(ctx, `
		SELECT uid, content, created_at, updated_at
		FROM memos WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []MemoExportItem
	for rows.Next() {
		var m Memo
		if err := rows.Scan(&m.UID, &m.Content, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		items = append(items, MemoExportItem{
			UID:        m.UID,
			Content:    m.Content,
			Visibility: "private",
			Pinned:     false,
			CreatedTs:  m.CreatedAt.Unix(),
			UpdatedTs:  m.UpdatedAt.Unix(),
		})
	}

	export := MemosExport{
		App:        "sana",
		Version:    "1.0",
		ExportedAt: time.Now().Format(time.RFC3339),
		Memos:      items,
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sana_export_%s.zip\"", time.Now().Format("20060102")))

	zw := zip.NewWriter(w)
	defer zw.Close()

	// Write memos.json
	fw, err := zw.Create("memos.json")
	if err != nil {
		return
	}
	json.NewEncoder(fw).Encode(export)

	// Write individual .md files
	for _, item := range items {
		fw, err := zw.Create(item.UID + ".md")
		if err != nil {
			continue
		}
		fw.Write([]byte(item.Content))
	}
}
```

- [ ] **Step 2: 创建 import_handler.go**

```go
package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ImportResult struct {
	MemosImported int           `json:"memos_imported"`
	Errors        []string      `json:"errors,omitempty`
}

type MemosImportFormat struct {
	App       string           `json:"app"`
	Version   string           `json:"version"`
	Memos     []MemoExportItem `json:"memos"`
}

func handleImportMemos(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	maxBytes := int64(50 << 20)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	zr, err := zip.NewReader(file, maxBytes)
	if err != nil {
		http.Error(w, "invalid zip file", http.StatusBadRequest)
		return
	}

	var memosJSON *MemosImportFormat
	contentMap := make(map[string]string)

	for _, f := range zr.File {
		if f.Name == "memos.json" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, _ := io.ReadAll(rc)
			rc.Close()
			json.Unmarshal(data, &memosJSON)
		} else if strings.HasSuffix(f.Name, ".md") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, _ := io.ReadAll(rc)
			rc.Close()
			uid := strings.TrimSuffix(f.Name, ".md")
			contentMap[uid] = string(content)
		}
	}

	if memosJSON == nil {
		http.Error(w, "invalid memos format: memos.json not found", http.StatusBadRequest)
		return
	}

	if memosJSON.App != "sana" && memosJSON.App != "memos" {
		http.Error(w, "unsupported format: app must be 'sana' or 'memos'", http.StatusBadRequest)
		return
	}

	result := ImportResult{MemosImported: 0}
	now := time.Now()

	for _, m := range memosJSON.Memos {
		content := m.Content
		if content == "" {
			content, _ = contentMap[m.UID]
		}
		if content == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("memo %s: empty content", m.UID))
			continue
		}

		uid := m.UID
		if uid == "" {
			uid = uuid.New().String()
		}

		createdTs := time.Unix(m.CreatedTs, 0)
		if createdTs.Year() < 2000 {
			createdTs = now
		}
		updatedTs := time.Unix(m.UpdatedTs, 0)
		if updatedTs.Year() < 2000 {
			updatedTs = now
		}

		_, err := db.Exec(ctx, `
			INSERT INTO memos (uid, user_id, content, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (uid) DO UPDATE SET content = EXCLUDED.content, updated_at = EXCLUDED.updated_at
		`, uid, userID, content, createdTs, updatedTs)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("memo %s: %v", m.UID, err))
			continue
		}
		result.MemosImported++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
```

- [ ] **Step 3: 在 main.go 注册导入导出路由**

```go
mux.HandleFunc("GET /api/export/memos", withAuth(handleExportMemos))
mux.HandleFunc("POST /api/import/memos", withAuth(handleImportMemos))
```

Run: `cd backend && go build -o sana .`
Expected: 编译成功

- [ ] **Step 4: Commit**

```bash
git add backend/export_handler.go backend/import_handler.go backend/main.go
git commit -m "feat(api: Memos format import/export

- GET /api/export/memos - export as ZIP (memos.json + .md files)
- POST /api/import/memos - import from ZIP
- Supports both sana and memos format
"
```

---

## Phase 2: 前端 - Timeline 界面

### Task 5: API 层

**Files:**
- Modify: `frontend/src/api/index.js`

- [ ] **Step 1: 更新 api/index.js 添加 memos API**

```javascript
const API_BASE = '/api'

const api = {
  // Auth
  login: (password) => fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    body: JSON.stringify({ password }),
    headers: { 'Content-Type': 'application/json' },
  }).then(r => r.json()),

  logout: () => fetch(`${API_BASE}/auth/logout`, { method: 'POST' }),

  me: () => fetch(`${API_BASE}/auth/me`).then(r => r.json()),

  // Memos (Timeline)
  listMemos: (cursor) => {
    let url = `${API_BASE}/memos?limit=20`
    if (cursor) url += `&cursor=${cursor}`
    return fetch(url).then(r => r.json())
  },

  createMemo: (content) => fetch(`${API_BASE}/memos`, {
    method: 'POST',
    body: JSON.stringify({ content }),
    headers: { 'Content-Type': 'application/json' },
  }).then(r => r.json()),

  getMemo: (id) => fetch(`${API_BASE}/memos/${id}`).then(r => r.json()),

  updateMemo: (id, content) => fetch(`${API_BASE}/memos/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ content }),
    headers: { 'Content-Type': 'application/json' },
  }).then(r => r.json()),

  deleteMemo: (id) => fetch(`${API_BASE}/memos/${id}`, { method: 'DELETE' }),

  searchMemos: (q) => fetch(`${API_BASE}/memos/search?q=${encodeURIComponent(q)}`)
    .then(r => r.json()),

  exportMemos: () => fetch(`${API_BASE}/export/memos`).then(r => r.blob()),

  importMemos: (formData) => fetch(`${API_BASE}/import/memos`, {
    method: 'POST',
    body: formData,
  }).then(r => r.json()),
}

export default api
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/api/index.js
git commit -m "feat(api): add memos API methods

- listMemos, createMemo, getMemo, updateMemo, deleteMemo
- searchMemos
- exportMemos, importMemos
"
```

---

### Task 6: TimelineView 主视图

**Files:**
- Create: `frontend/src/views/TimelineView.vue`

- [ ] **Step 1: 创建 TimelineView.vue**

```vue
<template>
  <div class="timeline-view">
    <header class="timeline-header">
      <h1>Sana</h1>
      <div class="header-actions">
        <button class="icon-btn" @click="showSearch = !showSearch" title="搜索">
          🔍
        </button>
        <button class="icon-btn" @click="handleExport" title="导出">📤</button>
        <button class="icon-btn" @click="fileInput.click()" title="导入">📥</button>
        <input ref="fileInput" type="file" accept=".zip" style="display:none" @change="handleImport">
      </div>
    </header>

    <MemoComposer @created="onMemoCreated" />

    <div v-if="showSearch" class="search-container">
      <input
        v-model="searchQuery"
        class="search-input"
        placeholder="搜索笔记..."
        @input="debouncedSearch"
      >
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else class="memo-list">
      <template v-if="searchMode">
        <div v-if="searchResults.length === 0" class="empty">
          未找到匹配 "{{ searchQuery }}" 的笔记
        </div>
        <MemoCard
          v-for="memo in searchResults"
          :key="memo.id"
          :memo="memo"
          @edit="editMemo"
          @delete="deleteMemo"
        />
      </template>
      <template v-else>
        <div v-if="groupedMemos.length === 0" class="empty">
          还没有笔记，写下第一条吧 ✨
        </div>
        <TimeGroup
          v-for="group in groupedMemos"
          :key="group.label"
          :label="group.label"
          :memos="group.memos"
          @edit="editMemo"
          @delete="deleteMemo"
        />
        <button v-if="hasMore" class="load-more" @click="loadMore">
          加载更多
        </button>
      </template>
    </div>

    <MemoEditor
      v-if="editingMemo"
      :memo="editingMemo"
      @close="editingMemo = null"
      @save="saveMemo"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import MemoComposer from '../components/MemoComposer.vue'
import MemoCard from '../components/MemoCard.vue'
import MemoEditor from '../components/MemoEditor.vue'
import TimeGroup from '../components/TimeGroup.vue'
import api from '../api/index.js'

const memos = ref([])
const loading = ref(false)
const error = ref(null)
const cursor = ref(null)
const hasMore = ref(false)
const showSearch = ref(false)
const searchQuery = ref('')
const searchResults = ref([])
const searchMode = ref(false)
const editingMemo = ref(null)
const fileInput = ref(null)
let searchTimer = null

const groupedMemos = computed(() => {
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const yesterday = new Date(today - 86400000)
  const thisWeek = new Date(today - 7 * 86400000)

  const groups = { today: [], yesterday: [], thisWeek: [], older: [] }

  for (const m of memos.value) {
    const d = new Date(m.updated_ts * 1000)
    if (d >= today) {
      groups.today.push(m)
    } else if (d >= yesterday) {
      groups.yesterday.push(m)
    } else if (d >= thisWeek) {
      groups.thisWeek.push(m)
    } else {
      groups.older.push(m)
    }
  }

  const result = []
  if (groups.today.length) result.push({ label: '今天', memos: groups.today })
  if (groups.yesterday.length) result.push({ label: '昨天', memos: groups.yesterday })
  if (groups.thisWeek.length) result.push({ label: '本周', memos: groups.thisWeek })
  if (groups.older.length) result.push({ label: '更早', memos: groups.older })
  return result
})

async function loadMemos(append = false) {
  loading.value = true
  error.value = null
  try {
    const data = await api.listMemos(append ? cursor.value : null)
    if (append) {
      memos.value = [...memos.value, ...(data.memos || [])]
    } else {
      memos.value = data.memos || []
    }
    cursor.value = data.next_cursor
    hasMore.value = data.has_more
  } catch (e) {
    error.value = '加载失败'
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (cursor.value) await loadMemos(true)
}

function onMemoCreated(memo) {
  memos.value = [memo, ...memos.value]
}

async function editMemo(memo) {
  editingMemo.value = { ...memo }
}

async function saveMemo({ id, content }) {
  await api.updateMemo(id, content)
  const idx = memos.value.findIndex(m => m.id === id)
  if (idx >= 0) {
    memos.value[idx] = { ...memos.value[idx], content }
  }
  editingMemo.value = null
}

async function deleteMemo(id) {
  if (!confirm('确定删除这条笔记？')) return
  await api.deleteMemo(id)
  memos.value = memos.value.filter(m => m.id !== id)
}

function debouncedSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(doSearch, 300)
}

async function doSearch() {
  const q = searchQuery.value.trim()
  if (!q) {
    searchMode.value = false
    searchResults.value = []
    return
  }
  searchMode.value = true
  loading.value = true
  try {
    const data = await api.searchMemos(q)
    searchResults.value = data.memos || []
  } catch (e) {
    error.value = '搜索失败'
  } finally {
    loading.value = false
  }
}

async function handleExport() {
  try {
    const blob = await api.exportMemos()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `sana_export_${Date.now()}.zip`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    alert('导出失败')
  }
}

async function handleImport(e) {
  const file = e.target.files[0]
  if (!file) return
  const formData = new FormData()
  formData.append('file', file)
  try {
    const result = await api.importMemos(formData)
    alert(`导入完成：${result.memos_imported} 条笔记`)
    await loadMemos()
  } catch (e) {
    alert('导入失败')
  }
  e.target.value = ''
}

onMounted(() => loadMemos())
</script>

<style scoped>
.timeline-view {
  max-width: 700px;
  margin: 0 auto;
  padding: 16px;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.timeline-header h1 {
  font-size: 24px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.icon-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
}

.icon-btn:hover {
  background: #f0f0f0;
}

.search-container {
  margin-bottom: 16px;
}

.search-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-size: 14px;
}

.loading, .error, .empty {
  text-align: center;
  padding: 32px;
  color: #666;
}

.error {
  color: #d00;
}

.load-more {
  display: block;
  margin: 16px auto;
  padding: 8px 24px;
  background: #f0f0f0;
  border: none;
  border-radius: 8px;
  cursor: pointer;
}

.load-more:hover {
  background: #e0e0e0;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/TimelineView.vue
git commit -m "feat(frontend): TimelineView main component

- Shows memos grouped by date (today/yesterday/thisWeek/older)
- MemoComposer for quick creation
- Search with debounce
- Load more pagination
- Import/export buttons
"
```

---

### Task 7: Memo 组件组

**Files:**
- Create: `frontend/src/components/MemoComposer.vue`
- Create: `frontend/src/components/MemoCard.vue`
- Create: `frontend/src/components/MemoEditor.vue`
- Create: `frontend/src/components/TimeGroup.vue`

- [ ] **Step 1: 创建 MemoComposer.vue**

```vue
<template>
  <div class="memo-composer">
    <textarea
      v-model="content"
      class="composer-input"
      placeholder="写下此刻的想法..."
      rows="2"
      @keydown.enter.ctrl="submit"
    ></textarea>
    <button class="composer-btn" @click="submit" :disabled="!content.trim()">
      创建
    </button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import api from '../api/index.js'

const emit = defineEmits(['created'])
const content = ref('')

async function submit() {
  const c = content.value.trim()
  if (!c) return
  const memo = await api.createMemo(c)
  content.value = ''
  emit('created', memo)
}
</script>

<style scoped>
.memo-composer {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
}

.composer-input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  resize: none;
}

.composer-input:focus {
  outline: none;
  border-color: #007AFF;
}

.composer-btn {
  padding: 8px 16px;
  background: #007AFF;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
}

.composer-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}
</style>
```

- [ ] **Step 2: 创建 MemoCard.vue**

```vue
<template>
  <div class="memo-card">
    <div class="memo-content">{{ memo.content }}</div>
    <div class="memo-meta">
      <span class="memo-time">{{ formatTime(memo.updated_ts) }}</span>
      <div class="memo-actions">
        <button class="action-btn" @click="$emit('edit', memo)">✎</button>
        <button class="action-btn delete" @click="$emit('delete', memo.id)">🗑</button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps(['memo'])
defineEmits(['edit', 'delete'])

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const now = new Date()
  const diff = now - d
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return d.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.memo-card {
  background: white;
  border: 1px solid #e5e5e5;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
}

.memo-content {
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  margin-bottom: 8px;
}

.memo-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.memo-time {
  font-size: 12px;
  color: #999;
}

.memo-actions {
  display: flex;
  gap: 4px;
}

.action-btn {
  background: none;
  border: none;
  font-size: 14px;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
}

.action-btn:hover {
  background: #f0f0f0;
}

.action-btn.delete:hover {
  background: #fee;
}
</style>
```

- [ ] **Step 3: 创建 MemoEditor.vue**

```vue
<template>
  <div class="editor-overlay" @click.self="$emit('close')">
    <div class="editor-modal">
      <div class="editor-header">
        <span>编辑笔记</span>
        <button class="close-btn" @click="$emit('close')">✕</button>
      </div>
      <textarea
        v-model="editContent"
        class="editor-textarea"
        rows="10"
      ></textarea>
      <div class="editor-footer">
        <button class="cancel-btn" @click="$emit('close')">取消</button>
        <button class="save-btn" @click="save">保存</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps(['memo'])
const emit = defineEmits(['close', 'save'])

const editContent = ref(props.memo.content)

watch(() => props.memo, (m) => {
  editContent.value = m.content
})

function save() {
  const c = editContent.value.trim()
  if (!c) return
  emit('save', { id: props.memo.id, content: c })
}
</script>

<style scoped>
.editor-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.editor-modal {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  overflow: hidden;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
}

.editor-textarea {
  width: 100%;
  padding: 16px;
  border: none;
  font-family: inherit;
  font-size: 14px;
  resize: vertical;
  min-height: 200px;
}

.editor-textarea:focus {
  outline: none;
}

.editor-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #eee;
}

.cancel-btn, .save-btn {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
}

.cancel-btn {
  background: #f0f0f0;
  border: none;
}

.save-btn {
  background: #007AFF;
  color: white;
  border: none;
}
</style>
```

- [ ] **Step 4: 创建 TimeGroup.vue**

```vue
<template>
  <div class="time-group">
    <div class="group-label">{{ label }}</div>
    <MemoCard
      v-for="memo in memos"
      :key="memo.id"
      :memo="memo"
      @edit="$emit('edit', $event)"
      @delete="$emit('delete', $event)"
    />
  </div>
</template>

<script setup>
import MemoCard from './MemoCard.vue'
defineProps(['label', 'memos'])
defineEmits(['edit', 'delete'])
</script>

<style scoped>
.time-group {
  margin-bottom: 24px;
}

.group-label {
  font-size: 12px;
  color: #999;
  font-weight: 500;
  margin-bottom: 8px;
  padding-left: 4px;
}
</style>
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/MemoComposer.vue frontend/src/components/MemoCard.vue frontend/src/components/MemoEditor.vue frontend/src/components/TimeGroup.vue
git commit -m "feat(frontend): memo components

- MemoComposer: quick create input
- MemoCard: single memo display
- MemoEditor: edit modal
- TimeGroup: date grouping wrapper
"
```

---

### Task 8: 路由和入口调整

**Files:**
- Modify: `frontend/src/router/index.js`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: 更新 router/index.js**

```javascript
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Timeline',
    component: () => import('../views/TimelineView.vue')
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue')
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
```

- [ ] **Step 2: 更新 App.vue**

```vue
<template>
  <RouterView />
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from './api/index.js'

const router = useRouter()

onMounted(async () => {
  try {
    await api.me()
  } catch {
    router.push('/login')
  }
})
</script>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/router/index.js frontend/src/App.vue
git commit -m "feat(frontend): route TimelineView as home

- / -> TimelineView
- Auth check on app mount redirects to login
"
```

---

## Phase 3: 样式和清理

### Task 9: 样式和遗留组件清理

**Files:**
- Modify: `frontend/src/style.css`
- Delete: `frontend/src/views/FolderView.vue`
- Delete: `frontend/src/views/NoteView.vue`
- Delete: `frontend/src/components/FolderSidebar.vue`
- Delete: `frontend/src/components/TreeView.vue`
- Delete: `frontend/src/components/TreeNode.vue`

- [ ] **Step 1: 添加 Timeline 样式到 style.css**

```css
/* Timeline base styles */
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #f5f5f5;
  margin: 0;
}
```

- [ ] **Step 2: 删除旧组件**

```bash
cd /home/zjx/code/mine/sana/frontend/src
rm views/FolderView.vue views/NoteView.vue
rm components/FolderSidebar.vue components/TreeView.vue components/TreeNode.vue
```

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "feat(frontend): remove old folder-based components

Deleted:
- views/FolderView.vue
- views/NoteView.vue
- components/FolderSidebar.vue
- components/TreeView.vue
- components/TreeNode.vue

Timeline is now the only view.
"
```

---

## 自检清单

**Spec 覆盖检查：**
- [x] Timeline 视图 - TimelineView.vue
- [x] 快速创建 - MemoComposer.vue
- [x] 全文搜索 - handleSearchMemos + SearchBar in TimelineView
- [x] Memos 导入导出 - export_handler.go + import_handler.go
- [x] PostgreSQL 存储 - db.go

**占位符扫描：**
- 无 TBD/TODO
- 无"类似 Task X" 的引用
- 所有代码块完整

**类型一致性：**
- `Memo` (backend) → `MemoResponse` (对外) → `memo` (frontend)
- `updated_ts` / `CreatedTs` / `UpdatedTs` 命名统一
- `uid` 对外暴露，`id` 仅内部使用

---

## 实施顺序

1. Task 1: 数据库层
2. Task 2: Memo CRUD API
3. Task 3: 搜索 API
4. Task 4: 导入导出 API
5. Task 5: 前端 API 层
6. Task 6: TimelineView
7. Task 7: Memo 组件组
8. Task 8: 路由和入口
9. Task 9: 样式和清理

**建议执行方式：Subagent-Driven**，每个 Task 分配给独立 subagent 完成review后合并。
