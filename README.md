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

- 后端：Go + SQLite + JWT
- 前端：Vue 3 + Vite

## 快速部署

### Docker Compose（推荐）

```bash
git clone https://github.com/enneket/sana.git
cd sana/docker
cp .env.example .env
# 编辑 .env，填入 JWT_SECRET 和 SANA_PASSWORD
docker compose up -d
```

访问 `http://localhost:5560`，使用 `SANA_PASSWORD` 设置的密码登录。

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SANA_PASSWORD` | 登录密码 | **必须设置** |
| `JWT_SECRET` | JWT 签名密钥 | **必须设置** |
| `PORT` | 服务端口 | `8080` |
| `SQLITE_PATH` | 数据库路径 | `/data/sana.db` |

## 数据存储

SQLite 数据库存储在 `/data/sana.db`（docker-compose 中挂载为 `sana_data` 目录）。

## 项目结构

```
sana/
├── backend/          # Go 后端
├── frontend/         # Vue 3 前端
│   └── src/
│       ├── views/
│       │   ├── TimelineView.vue  # 主界面
│       │   └── Login.vue         # 登录页
│       └── components/
│           ├── SanaComposer.vue  # 快速创建
│           ├── SanaCard.vue      # 笔记卡片
│           ├── SanaEditor.vue    # 编辑弹窗
│           └── TimeGroup.vue     # 日期分组
├── docker/          # Docker 部署
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── .env.example
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
