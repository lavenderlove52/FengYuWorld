# My-Blog (Go + Vue 3)

多用户博客 / 日记系统。Go + Gin 后端，Vue 3 + Naive UI 前端，PostgreSQL 存储。

## 技术栈

- **后端**: Go 1.26, Gin, Gorm, PostgreSQL 16, goose, JWT, goldmark
- **前端**: Vue 3, Vite, TypeScript, Pinia, Naive UI, Tailwind CSS, md-editor-v3

## 目录结构

```text
backend/                Go API（cmd/server 启动、cmd/migrate 迁移、migrations/ 建表脚本）
frontend/               Vue 3 SPA
  src/core/gridplan/    定投计划的盈亏计算、状态与旧数据迁移
deploy/                 nginx 配置样例
docs/                   PostgreSQL 安装指南（Windows / CentOS7 / Docker 三种方式）
docker-compose.yml      本地开发用的 PostgreSQL 16
```

## 环境要求

| 组件 | 版本 | 说明 |
| --- | --- | --- |
| Go | 1.26.2（见 `backend/go.mod`） | 低于此版本编译不过 |
| Node.js | 20+ | 前端构建，`npm install` 用 |
| PostgreSQL | 16 | 本地可用 Docker，服务器见 [docs/](docs/README.md) |

数据库的三种安装方式各有独立文档：Windows 本机、CentOS 7 本机、Windows + Docker。见 [docs/README.md](docs/README.md)。

## 本地开发

### 1. 启动数据库

推荐 Docker（`docker-compose.yml` 已配好）：

```bash
docker compose up -d postgres
```

> 注意：当前 `docker-compose.yml` 的数据卷写死成 `D:\Code\Go\My-Blog\postgres`，如果项目不在这个路径，Docker 可能启动失败或把数据写到别处。改成相对路径 `./postgres` 更稳妥。

不用 Docker 的话，按 [docs/INSTALL_POSTGRESQL_WINDOWS.md](docs/INSTALL_POSTGRESQL_WINDOWS.md) 装本机 PostgreSQL 16。

### 2. 后端配置

```bash
cp .env.example backend/.env
```

至少要改这几项：

```ini
APP_ENV=development
APP_PORT=8080
DATABASE_URL=postgres://myblog:密码@127.0.0.1:5432/myblog?sslmode=disable
JWT_ACCESS_SECRET=随机至少32字符
JWT_REFRESH_SECRET=另一组随机至少32字符
```

后端通过 godotenv 从项目根目录或 `backend/` 读取 `.env`。数据库密码里的特殊字符要 URL 编码：`@` → `%40`，`%` → `%25`，`#` → `%23`，`:` → `%3A`，`/` → `%2F`，`?` → `%3F`，`&` → `%26`。

### 3. 建表并启动后端

```bash
# 建表（幂等，可重复执行）
make migrate

# 启动（前台，Ctrl+C 停止）
make backend-run

# 验证
curl http://127.0.0.1:8080/api/health
```

首次在新机器上跑，国内网络先设代理：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

### 4. 启动前端

```bash
cd frontend
npm install
npm run dev
```

打开 `http://localhost:5173/blog/`。Vite 已配好 `/blog/api` 和 `/uploads` 到 `http://localhost:8080` 的代理，后端不用开 CORS。

## 首次部署到服务器

以下步骤只在**新服务器**上做一次；日常更新见下一节。

### 1. 准备环境

安装 Go 1.26.2（阿里云镜像）：

```bash
curl -LO https://mirrors.aliyun.com/golang/go1.26.2.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.26.2.linux-amd64.tar.gz
cat > /etc/profile.d/go.sh <<'EOF'
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go
export PATH=$PATH:$GOPATH/bin
EOF
source /etc/profile.d/go.sh
go version
```

### 2. 安装 PostgreSQL 16

CentOS 7 上源码编译的完整步骤见 [docs/INSTALL_POSTGRESQL_CENTOS7.md](docs/INSTALL_POSTGRESQL_CENTOS7.md)。装完后要建库建用户：

```sql
CREATE USER myblog WITH PASSWORD '强密码';
CREATE DATABASE myblog OWNER myblog ENCODING 'UTF8';
GRANT ALL PRIVILEGES ON DATABASE myblog TO myblog;
\c myblog
GRANT ALL ON SCHEMA public TO myblog;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO myblog;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO myblog;
```

验证：`PGPASSWORD='强密码' psql -h 127.0.0.1 -U myblog -d myblog -c "SELECT 1;"`

### 3. 拉代码、配置、迁移

```bash
cd /root/project/FengYuWorld
git pull
cp .env.example backend/.env
vim backend/.env          # 按上一节改 DATABASE_URL 和两个 JWT 密钥

cd backend
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
go run ./cmd/migrate      # 输出 migrations completed 即建表成功
go build -o bin/server ./cmd/server
```

### 4. 创建第一个管理员账号

**不能跳过。** 注册要求「邀请码 + 邮箱验证码」，而邀请码只能由已登录的管理员在后台生成（`invite_codes.created_by` 外键指向 `users.id`）。全新库里 `users` 是空表，没有管理员就发不出邀请码，正常流程永远注册不出第一个账号；项目也没有内置 seed 命令。所以第一个管理员只能直接写库。

登录服务器，用 `psql` 执行下面这段 SQL（纯 SQL，密码哈希由 `pgcrypto` 的 `crypt()` 现场生成）：

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (username, email, password_hash, nickname, role, status, can_site_pin, can_custom_js)
VALUES (
  'admin',                                     -- 登录名，3~50 字符
  'admin@your-domain.com',                     -- 登录邮箱，唯一
  crypt('换成你的强密码', gen_salt('bf', 12)),  -- bcrypt 12 轮
  '站长',                                       -- 昵称
  'super_admin',                               -- 第一个号必须给 super_admin
  'active',                                    -- locked 会被登录接口直接拒绝
  true,                                        -- 可设置首页置顶
  true                                         -- 可用自定义 JS
);

-- 顺手生成一个邀请码，之后注册普通用户要用
INSERT INTO invite_codes (code, created_by, expires_at)
SELECT 'WELCOME2026', id, now() + interval '30 days'
FROM users WHERE username = 'admin';
```

执行方式（`-v ON_ERROR_STOP=1` 保证中途出错就停，不会半截写入）：

```bash
PGPASSWORD='数据库密码' psql -h 127.0.0.1 -U myblog -d myblog -v ON_ERROR_STOP=1 -f bootstrap-admin.sql
```

或用 heredoc 直接粘：

```bash
PGPASSWORD='数据库密码' psql -h 127.0.0.1 -U myblog -d myblog -v ON_ERROR_STOP=1 <<'SQL'
CREATE EXTENSION IF NOT EXISTS pgcrypto;
INSERT INTO users (username, email, password_hash, nickname, role, status, can_site_pin, can_custom_js)
VALUES ('admin', 'admin@your-domain.com', crypt('换成你的强密码', gen_salt('bf', 12)),
        '站长', 'super_admin', 'active', true, true);
INSERT INTO invite_codes (code, created_by, expires_at)
SELECT 'WELCOME2026', id, now() + interval '30 days' FROM users WHERE username = 'admin';
SQL
```

几个必须注意的点：

- `role` 只有 `user` / `admin` / `super_admin`。`admin` 能管邀请码和用户权限，但站点级设置（`/api/super-admin/*`）只认 `super_admin`，所以第一个号用 `super_admin`。
- `password_hash` 必须是 bcrypt（`$2a$` / `$2b$` / `$2y$` 前缀后端都认）。写明文、md5、sha256 一律登录失败，而且不会报错，只是返回「登录失败」。
- 密码 6~72 字节，bcrypt 上限。
- 如果服务器上的 PostgreSQL 没带 `pgcrypto`，`CREATE EXTENSION` 会失败，改成先本地算哈希再填整串 `$2b$12$...`：

  ```bash
  python -c "import bcrypt;print(bcrypt.hashpw(b'你的密码', bcrypt.gensalt(12)).decode())"
  ```

- `username`、`email` 都是 UNIQUE，重复插入会直接报错，不会覆盖已有账号。

验证：

```sql
SELECT id, username, email, role, status FROM users;
SELECT code, is_used, expires_at FROM invite_codes;
```

然后打开 `/blog/login` 登录。注意登录强制图形验证码：5 位、区分大小写、5 分钟过期、**验证一次即作废**，所以只能提交一次，失败要刷新图片。如果提示是「登录失败」而不是「验证码错误或已过期」，说明验证码过了、是账号密码没对上。

登录后侧边栏会多出「🔒 邀请码」「🔒 首页置顶」。在「邀请码」页生成新码发给别人，他们才能注册。改密码目前只有「忘记密码」一条路（依赖 `.env` 的 `SMTP_*`），后台没有改密码接口——所以 `WELCOME2026` 这类固定邀请码用掉后记得删掉。

### 5. 用 systemd 常驻

```bash
cat > /etc/systemd/system/myblog.service <<'EOF'
[Unit]
Description=My-Blog API
After=network.target

[Service]
Type=simple
WorkingDirectory=/root/project/FengYuWorld/backend
EnvironmentFile=/root/project/FengYuWorld/backend/.env
ExecStart=/root/project/FengYuWorld/backend/bin/server
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now myblog
systemctl status myblog --no-pager
```

常用命令：`systemctl restart myblog` 重启、`journalctl -u myblog -f` 跟日志、`systemctl stop myblog` 停止。

### 6. 配置 nginx

```bash
mkdir -p /var/www/blog /etc/nginx/conf.d/disabled
cp deploy/nginx-ip-paths.conf /etc/nginx/conf.d/myblog.conf
# 若与本机已有的 relay 配置抢 80 端口，先禁用旧的
mv /etc/nginx/conf.d/jian-guo-yun-relay.conf /etc/nginx/conf.d/disabled/ 2>/dev/null
nginx -t && systemctl reload nginx
```

路径分配原则：每个服务占一个前缀，不要抢 `/`。要加新服务就在 `deploy/nginx-ip-paths.conf` 里加一个 `location /xxx/ { proxy_pass http://127.0.0.1:端口/; }`。

### 7. 上传前端

```bash
# 本地
cd frontend && npm install && npm run build
scp -r dist/* root@124.223.116.8:/var/www/blog/
```

访问 `http://124.223.116.8/blog/`。

## 日常发布

日常发布只做两件事之一：**只更前端**（纯静态文件）或 **更后端**（重新编译 + 重启服务）。两者互不依赖，不要每次都全量走一遍。

### 部署拓扑

先搞清楚每个东西落在哪，后面所有命令都围绕这张表：

| 内容 | 本地位置 | 服务器位置 | 更新后要不要重启 |
| --- | --- | --- | --- |
| 前端产物 | `frontend/dist/` | `/var/www/blog/` | 不用，替换文件即可 |
| 后端二进制 | 服务器上编译 | `/root/project/FengYuWorld/backend/bin/server` | 要，`systemctl restart myblog` |
| 后端配置 | `backend/.env` | 同名路径 | 要重启才生效 |
| 数据库结构 | `backend/migrations/` | 由 `cmd/migrate` 写入 | 要，在执行 build 前后跑一次 |
| 上传文件 | `backend/uploads/` | 同名路径 | 不用，注意别被覆盖 |

Nginx 只做路径分流，**只有改 `deploy/nginx-ip-paths.conf` 时才需要 reload**：

```text
/blog/        → /var/www/blog/        前端 SPA（index.html 禁缓存，assets 长期缓存）
/blog/api/    → 127.0.0.1:8080/api/   Go API
/uploads/     → 127.0.0.1:8080        上传文件
/webdav/      → 127.0.0.1:10000/      坚果云 relay（其他服务示例）
```

### 只更新前端

前端是纯静态资源，流程是「本地打包 → 覆盖上传 → 浏览器强刷」，服务器上不重启任何进程。

```bash
# 1. 本地打包（在 frontend/ 目录下，产物是 dist/）
cd frontend
npm install          # 首次或依赖变化时
npm run build

# 2. 上传覆盖（注意目标路径是 /var/www/blog/，因为 nginx 是 root /var/www + location /blog/）
scp -r dist/* root@124.223.116.8:/var/www/blog/

# 3. 验证
curl -sI http://124.223.116.8/blog/ | head -3
```

浏览器打开 `http://124.223.116.8/blog/`，按 **Ctrl+F5** 强刷一次。

几个容易踩的点：

- **必须重新 build 再传**。直接传旧 `dist/` 等于把旧版本又发一遍，前端源码改了不 build 是不会生效的。
- **目标路径不能写错**。写 `/var/www/` 会覆盖掉别的站点，写 `/var/www/blog`（末尾没斜杠）在部分环境会变成把 dist 目录整个塞进去，统一用 `dist/*` 加 `/var/www/blog/`。
- **旧 assets 不会自动清理**。带哈希的 `assets/*.js` 会一直堆积，隔一段时间清一次：

  ```bash
  ssh root@124.223.116.8 "ls -t /var/www/blog/assets/*.js | tail -n +200 | xargs -r rm -f"
  ```

  只删最老的、保留最近 200 个是安全的：`index.html` 只引用当前这一版哈希，旧文件已无引用。

- **本地未提交的改动会被一起打进去**。`npm run build` 不看 git 状态，`dist/` 里是什么就发什么。上线前先 `git status` 确认一遍，不想发的先 `git stash`。

### 只更新后端

后端是编译型服务，流程是「拉代码 → 迁移 → 编译 → 重启」。

```bash
# 全部在服务器上执行
ssh root@124.223.116.8

cd /root/project/FengYuWorld
git pull

# 1. 数据库迁移（有新 migration 时；无变更时重复执行也安全）
cd backend
go run ./cmd/migrate

# 2. 编译
go build -o bin/server ./cmd/server

# 3. 重启
systemctl restart myblog

# 4. 验证
systemctl status myblog --no-pager
curl http://127.0.0.1:8080/api/health
journalctl -u myblog -n 50 --no-pager     # 有异常看这里
```

注意事项：

- **迁移放在编译之前、重启之前**。新代码可能依赖新字段，顺序反了会先崩一次。
- **`.env` 改了也要重启**，`systemd` 的 `EnvironmentFile` 只在启动时读一次。
- **不要在本地 Windows 上 `go build` 后传二进制**，除非显式交叉编译；直接本地编出来的 `.exe` 在服务器上跑不了：

  ```bash
  GOOS=linux GOARCH=amd64 go build -o bin/server ./cmd/server
  ```

  更省事的做法还是上面那样在服务器上编。
- **服务器 Go 版本**要和 `go.mod` 要求一致，升级 Go 后要重新 `go mod download`。

### 前后端都要更新

顺序建议 **先后端、后前端**：

```bash
# 服务器
cd /root/project/FengYuWorld && git pull
cd backend && go run ./cmd/migrate && go build -o bin/server ./cmd/server
systemctl restart myblog
curl http://127.0.0.1:8080/api/health

# 本地
cd frontend && npm run build
scp -r dist/* root@124.223.116.8:/var/www/blog/
```

先更后端的好处是：新接口先就位，旧前端仍能用；反过来先发前端，新前端调新接口会短暂 404。

### 回滚

前端回滚最快，把上一版 `dist` 重新传一次即可，建议本地保留最近几版的备份：

```bash
# 本地打包后顺手留一份
cp -r dist dist.bak.$(date +%Y%m%d-%H%M)
```

后端回滚走 git：

```bash
cd /root/project/FengYuWorld
git log --oneline -5
git checkout <上一个正常的提交>
cd backend && go build -o bin/server ./cmd/server
systemctl restart myblog
# 确认没问题后再回到分支：git checkout main
```

数据库迁移通常不可逆，回滚代码前先确认新 migration 是否已写入生产数据。

### 只有改了 nginx 配置才需要 reload

```bash
cp deploy/nginx-ip-paths.conf /etc/nginx/conf.d/myblog.conf
nginx -t && systemctl reload nginx
```

`reload` 是平滑重载，不会断已有连接；`restart` 会短暂中断，能用 `reload` 就别用 `restart`。

## 已实现功能

- 用户注册、登录、JWT（Access + Refresh）、当前用户资料
- 文章 CRUD、草稿/发布、公开首页、文章详情、Markdown HTML 缓存
- 文章可见性：public / private / friends / specified + `CanReadPost`
- 好友申请、接受、拒绝、列表、删除，好友聊天（WebSocket）
- 分类与标签（用户维度，编辑文章时使用）、文章目录树
- 图片上传（正文图片、头像），本地存储
- 手机推送（ntfy）、公开推送页
- Vue 前端：登录、注册、首页、详情、用户中心、写文章、好友、设置、主页定制
- 倒金字塔网格定投计划（`/dashboard/grid-plans`）：每份计划是独立资金池，数据存后端、一人可有多份；支持网格/权重/预算/回撤闸门、当前市值与盈亏、资产占比饼图、收益变化曲线，以及旧版 localStorage 计划一键导入

## 暂不实现

评论、全文搜索、管理员后台、对象存储（COS）、Redis、SEO、多主题、友链、站点配置。这些在设计文档里明确不属于第一版。