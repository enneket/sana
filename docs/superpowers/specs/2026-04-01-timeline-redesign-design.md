# Sana Timeline 重构设计

## 概述

将 Sana 从文件夹导航模式重构为 Timeline（时间线）模式，参考 Memos 的 Instant Capture 理念。

## 功能清单

1. **Timeline 视图** — 主界面，按更新时间倒序展示所有笔记
2. **快速创建** — 顶部即时输入框，回车创建笔记
3. **全文搜索** — 搜索笔记内容
4. **Memos 格式导入导出** — 支持 Memos 备份格式

## 技术架构

### 后端变更

#### 新增接口

**1. Timeline 列表**
```
GET /api/memos?limit=20&cursor=<timestamp>
```
- Cursor-based 分页，基于 `updated_ts`
- 返回笔记列表，按更新时间倒序
- 沿用现有 JWT 认证

**2. 创建笔记**
```
POST /api/memos
Body: { "content": "笔记内容" }
```
- 笔记内容即 `content`，无需 `title`
- 自动生成 `uid`、`created_ts`、`updated_ts`

**3. 更新笔记**
```
PUT /api/memos/{id}
Body: { "content": "新内容" }
```
- 更新 `content` 和 `updated_ts`

**4. 删除笔记**
```
DELETE /api/memos/{id}
```

**5. 搜索**
```
GET /api/memos/search?q=<keyword>
```
- 模糊匹配 `content` 字段
- 返回所有匹配结果（不分页）

**6. 导出 Memos 格式**
```
GET /api/export/memos
```
- 返回 ZIP 文件
- 包含 `memos.json`（元数据）+ 每个笔记一个 `.md` 文件

**7. 导入 Memos 格式**
```
POST /api/import/memos
```
- 上传 ZIP 文件
- 解析 `memos.json`，导入所有 memo

#### 数据模型变更

**`memo` 结构（数据库模型）：**
```go
type Memo struct {
    ID        int       `json:"id"`        // 数据库自增主键
    UID       string    `json:"uid"`       // 用户可见的唯一标识（用于 API）
    UserID    string    `json:"user_id"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**API 响应格式（对外）：**
```go
type MemoResponse struct {
    ID        string `json:"id"`        // 对外暴露 UID
    Content   string `json:"content"`
    CreatedTs int64  `json:"created_ts"`
    UpdatedTs int64  `json:"updated_ts"`
}
```

**移除：**
- `folder` 相关接口（`/api/folders/*`）
- `noteRecord` 中的 `Title`、`FolderID`、`Filename`

**保留（兼容）：**
- `user` 结构不变
- JWT 认证逻辑不变

#### 存储

**PostgreSQL 数据库**

通过 `DATABASE_URL` 环境变量配置连接：
```
postgres://user:password@host:5432/sana?sslmode=disable
```

**数据库 Schema**

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE memos (
    id SERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_memos_user_id ON memos(user_id);
CREATE INDEX idx_memos_updated_at ON memos(updated_at DESC);
CREATE INDEX idx_memos_content_gin ON memos USING gin(to_tsvector('simple', content));
```

**迁移说明**
- 启动时自动执行 schema 初始化（`CREATE TABLE IF NOT EXISTS`）
- 旧数据（`notes/`、`folders.json`、`notes.json`）保留，暂不迁移
- 未来提供一次性迁移脚本

### 前端变更

#### 路由

| 路由 | 组件 | 说明 |
|------|------|------|
| `/` | `TimelineView` | 主界面 |
| `/login` | `LoginView` | 登录页（保持不变）|

#### TimelineView 布局

```
┌─────────────────────────────────────────┐
│  Sana              [🔍 搜索...] [用户]    │
├─────────────────────────────────────────┤
│  ┌─────────────────────────────────┐   │
│  │ ✏️ 写下此刻的想法...             │   │  ← 快速创建输入框
│  └─────────────────────────────────┘   │
│                                         │
│  ─── 今天 ───                           │
│  ┌─────────────────────────────────┐   │
│  │ 笔记内容摘要...                   │   │
│  │ 2分钟前               ✎  🗑      │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─── 昨天 ───                           │
│  ┌─────────────────────────────────┐   │
│  │ 另一个笔记...                     │   │
│  │ 昨天 14:30            ✎  🗑      │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

#### 组件清单

| 组件 | 职责 |
|------|------|
| `TimelineView` | 主视图，管理状态 |
| `MemoComposer` | 顶部快速创建输入框 |
| `MemoCard` | 单条笔记展示 |
| `MemoEditor` | 笔记编辑弹窗 |
| `TimeGroup` | 按日期分组（今天/昨天/本周/更早）|
| `SearchBar` | 搜索输入框 |

#### 状态管理

- 使用 Vue 3 Composition API + `ref`/`reactive`
- 笔记列表存于 `TimelineView` 局部状态
- 无需 Vuex/Pinia（简单场景）

### Memos 格式规范

#### 导出格式

ZIP 文件结构：
```
memos_20260401.zip
├── memos.json      # 元数据
├── abc123.md       # 笔记正文
├── def456.md
└── ...
```

`memos.json` 结构：
```json
{
  "app": "sana",
  "version": "1.0",
  "exported_at": "2026-04-01T12:00:00Z",
  "memos": [
    {
      "uid": "abc123",
      "content": "笔记内容，支持 #tag 语法",
      "visibility": "private",
      "pinned": false,
      "created_ts": 1234567890,
      "updated_ts": 1234567890
    }
  ]
}
```

#### 导入逻辑

1. 解析 ZIP 中的 `memos.json`
2. 校验格式（`app` 字段必须为 `sana` 或 `memos`）
3. 遍历 `memos` 数组，创建每条笔记
4. 返回导入结果（成功数、失败数）

## 实施计划

| Phase | 内容 | 产出 |
|-------|------|------|
| 1 | 后端 Timeline API | `/api/memos` CRUD 接口 |
| 2 | 前端 Timeline 界面 | TimelineView + MemoCard |
| 3 | 快速创建功能 | MemoComposer |
| 4 | 搜索功能 | `/api/memos/search` + SearchBar |
| 5 | Memos 导入导出 | `/api/export/memos` + `/api/import/memos` |

## 兼容与迁移

- 旧数据（`notes/`、`folders.json`）保持只读
- 未来提供迁移脚本，将旧数据导入 Timeline
- 新接口与旧接口共存一段时间，逐步废弃旧接口

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DATABASE_URL` | PostgreSQL 连接串 | **必须设置** |
| `JWT_SECRET` | JWT 签名密钥 | **必须设置** |
| `PORT` | 服务端口 | `8080` |
| `SANA_PASSWORD` | 登录密码（首次启动创建默认用户） | **必须设置** |
