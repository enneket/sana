# Sana

简洁的自托管笔记应用，支持文件夹管理和 Markdown 写作。

## 功能

- **文件夹管理** — 二级树形目录，支持右键新建/删除
- **Markdown 编辑** — 实时编辑 + 预览模式，1 秒自动保存
- **自托管** — Docker 部署，数据完全自己掌控
- **密码登录** — 无注册流程，单密码访问

## 技术栈

- 后端：Go + SQLite + JWT
- 前端：Vue 3 + Vite

## 快速部署

### Docker Compose（推荐）

```bash
# 克隆项目
git clone https://github.com/enneket/sana.git
cd sana

# 启动
docker-compose up -d
```

访问 `http://localhost:8080`，使用环境变量 `SANA_PASSWORD` 设置的密码登录，默认密码为 `sana123`。

### 手动运行

```bash
# 后端
cd backend
go build -o sana .
SANA_PASSWORD=你的密码 JWT_SECRET=随机密钥 ./sana

# 前端
cd frontend
npm install
npm run dev
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SANA_PASSWORD` | 登录密码 | `sana123` |
| `JWT_SECRET` | JWT 签名密钥 | `change-me-in-production` |

## 数据存储

- **元数据**：`sana.db`（SQLite）— 用户、文件夹、笔记关系
- **笔记正文**：`notes/` 目录下的 `.md` 文件

## 开发

```bash
# 前端开发
cd frontend
npm install
npm run dev

# 后端开发
cd backend
go run .
```

## 项目结构

```
sana/
├── backend/          # Go 后端
│   ├── main.go      # 入口
│   ├── handlers/    # API 接口
│   └── Sana.db     # SQLite 数据库
├── frontend/        # Vue 3 前端
│   └── src/
│       ├── components/
│       │   ├── TreeView.vue   # 树形目录
│       │   └── TreeNode.vue   # 递归文件夹节点
│       └── views/
│           ├── NoteView.vue   # 笔记编辑器
│           ├── Login.vue      # 登录页
│           └── Layout.vue     # 主布局
└── notes/          # 笔记文件（Markdown）
```

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/login` | 登录 |
| POST | `/api/auth/logout` | 登出 |
| GET | `/api/folders` | 获取文件夹列表 |
| POST | `/api/folders` | 创建文件夹 |
| PUT | `/api/folders/:id` | 重命名文件夹 |
| DELETE | `/api/folders/:id` | 删除文件夹 |
| GET | `/api/notes?folder_id=` | 获取文件夹下笔记 |
| POST | `/api/notes` | 创建笔记 |
| GET | `/api/notes/:id` | 获取笔记内容 |
| PUT | `/api/notes/:id` | 更新笔记 |
| DELETE | `/api/notes/:id` | 删除笔记 |
