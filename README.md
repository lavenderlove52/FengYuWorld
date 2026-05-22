# My-Blog (Go + Vue 3)

多用户博客/日记系统 MVP，基于设计文档 `GO_BLOG_DETAILED_DESIGN.md` 实现。

## 技术栈

- **后端**: Go, Gin, Gorm, PostgreSQL, goose, JWT, goldmark
- **前端**: Vue 3, Vite, TypeScript, Pinia, Naive UI, Tailwind CSS, md-editor-v3

## 快速开始

### 1. 环境要求

- Go 1.22+
- Node.js 18+
- Docker（用于 PostgreSQL，或本地已有 PostgreSQL）

### 2. 配置

```bash
cp .env.example .env
# 按需修改 DATABASE_URL、JWT 密钥等
```

后端会从项目根目录或 `backend/` 目录读取 `.env`（通过 godotenv）。

### 3. 启动数据库

任选一种方式（详见 `docs/`）：

- **Windows 本机安装（推荐，不依赖 Docker/WSL）**： [docs/INSTALL_POSTGRESQL_WINDOWS.md](docs/INSTALL_POSTGRESQL_WINDOWS.md)
- **CentOS 7 服务器**： [docs/INSTALL_POSTGRESQL_CENTOS7.md](docs/INSTALL_POSTGRESQL_CENTOS7.md)
- **Docker**： [docs/INSTALL_POSTGRESQL_DOCKER.md](docs/INSTALL_POSTGRESQL_DOCKER.md)

```bash
make db-up
# 或: docker compose up -d postgres
```

### 4. 数据库迁移

```bash
make migrate
# 或: cd backend && go run ./cmd/migrate
```

### 5. 启动后端

```bash
make backend-run
# 或: cd backend && go run ./cmd/server
```

API 默认监听 `http://localhost:8080`。

### 6. 启动前端

```bash
make frontend-install
make frontend-dev
# 或: cd frontend && npm install && npm run dev
```

前端默认 `http://localhost:5173`，已配置 `/api` 与 `/uploads` 代理到后端。

## 已实现功能

- 用户注册、登录、JWT（Access + Refresh）、当前用户资料
- 文章 CRUD、草稿/发布、公开首页、文章详情、Markdown HTML 缓存
- 文章可见性：public / private / friends / specified + `CanReadPost`
- 好友申请、接受、拒绝、列表、删除
- 分类与标签（用户维度，编辑文章时使用）
- 图片上传（正文图片、头像），本地存储
- Vue 前端：登录、注册、首页、详情、用户中心、写文章、好友、设置

## 待实现（设计文档明确不做第一版）

评论、搜索、管理员后台、COS、Redis、SEO、多主题、友链、站点配置等。

## 目录结构

```text
backend/    Go API
frontend/   Vue 3 SPA
migrations/ PostgreSQL 迁移（位于 backend/migrations）
uploads/    本地上传目录（位于 backend/uploads）
```

## 生产构建

```bash
make backend-build
make frontend-build
```

前端产物在 `frontend/dist`，可由 Nginx 托管；`/api/*` 反代到 Go 服务，`/uploads/*` 指向 `backend/uploads`。
