# Sana

简洁的自托管笔记应用，Timeline-first 设计，支持 Markdown 写作。

## 功能

- **Timeline 视图** — 时间线展示所有笔记，按更新时间倒序
- **快速创建** — 顶部即时输入框，回车即创建
- **全文搜索** — 快速搜索笔记内容
- **Memos 导入导出** — 支持 Memos 格式备份
- **自托管** — Docker 部署，数据完全自己掌控
- **密码登录** — 无注册流程，单密码访问

## 技术栈

- 后端：Go + PostgreSQL + JWT
- 前端：Vue 3 + Vite

## 快速部署

### Docker Compose（推荐）

```bash
git clone https://github.com/enneket/sana.git
cd sana/docker

JWT_SECRET=your-secret SANA_PASSWORD=your-password docker compose up -d
```

访问 `http://localhost:8080`，使用 `SANA_PASSWORD` 设置的密码登录。

### 手动运行

```bash
# 安装 PostgreSQL 并创建数据库

# 后端
cd backend
go build -o sana .
DATABASE_URL=postgres://user:pass@localhost:5432/sana SANA_PASSWORD=你的密码 JWT_SECRET=随机密钥 ./sana

# 前端
cd frontend
npm install
npm run dev
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SANA_PASSWORD` | 登录密码 | **必须设置** |
| `JWT_SECRET` | JWT 签名密钥 | **必须设置** |
| `DATABASE_URL` | PostgreSQL 连接串 | **必须设置** |
| `PORT` | 服务端口 | `8080` |

## 数据存储

PostgreSQL 数据库存储所有笔记数据。启动时自动创建表结构。

## 项目结构

```
sana/
├── backend/          # Go 后端
│   ├── main.go       # 入口
│   ├── db.go         # PostgreSQL 连接
│   ├── memo_handler.go     # Memo CRUD API
│   ├── search_handler.go   # 搜索 API
│   ├── export_handler.go   # 导出
│   └── import_handler.go  # 导入
├── frontend/         # Vue 3 前端
│   └── src/
│       ├── views/
│       │   ├── TimelineView.vue  # 主界面
│       │   └── Login.vue         # 登录页
│       └── components/
│           ├── MemoComposer.vue   # 快速创建
│           ├── MemoCard.vue      # 笔记卡片
│           ├── MemoEditor.vue    # 编辑弹窗
│           └── TimeGroup.vue     # 日期分组
├── docker/          # Docker 部署
│   ├── Dockerfile
│   └── docker-compose.yml
└── docs/            # 设计文档
```

## API

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/login` | 登录 |
| POST | `/api/auth/logout` | 登出 |
| GET | `/api/auth/me` | 当前用户 |

### 笔记 (Memos)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/memos` | 笔记列表（分页） |
| POST | `/api/memos` | 创建笔记 |
| GET | `/api/memos/:id` | 获取单条笔记 |
| PUT | `/api/memos/:id` | 更新笔记 |
| DELETE | `/api/memos/:id` | 删除笔记 |
| GET | `/api/memos/search?q=` | 搜索笔记 |

### 导入导出

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/export/memos` | 导出 Memos 格式 ZIP |
| POST | `/api/import/memos` | 导入 Memos 格式 ZIP |
