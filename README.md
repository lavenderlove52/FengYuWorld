# My-Blog (Go + Vue 3)

多用户博客/日记系统 MVP，基于设计文档 `GO_BLOG_DETAILED_DESIGN.md` 实现。

## 技术栈

- **后端**: Go, Gin, Gorm, PostgreSQL, goose, JWT, goldmark
- **前端**: Vue 3, Vite, TypeScript, Pinia, Naive UI, Tailwind CSS, md-editor-v3

## 快速开始

### 1. 环境要求

#### go

```bash
# 1. 阿里云镜像
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

### 2. 配置

```bash
cd /root/project/FengYuWorld
cp .env.example backend/.env
vim backend/.env
# 按需修改 DATABASE_URL、JWT 密钥等
#至少改这些：
APP_ENV=production
APP_PORT=8080
DATABASE_URL=postgres://myblog:密码@127.0.0.1:5432/myblog?sslmode=disable
JWT_ACCESS_SECRET=随机至少32字符
JWT_REFRESH_SECRET=另一组随机至少32字符
```

密码里的特殊字符要转成 URL 编码（`%XX`）：

| 字符  | 编码    |
| --- | ----- |
| `@` | `%40` |
| `%` | `%25` |
| `#` | `%23` |
| `:` | `%3A` |
| `/` | `%2F` |
| `?` | `%3F` |
| `&` | `%26` |

后端会从项目根目录或 `backend/` 目录读取 `.env`（通过 godotenv）。

### 3. 启动数据库

```bash
# 第一步：安装编译依赖
yum install -y gcc gcc-c++ make readline-devel zlib-devel openssl-devel
yum install -y libicu-devel
# 第二步：下载源码（用清华镜像）
cd /tmp
wget https://mirrors.tuna.tsinghua.edu.cn/postgresql/source/v16.3/postgresql-16.3.tar.gz

# 第三步：解压编译安装
tar -xzf postgresql-16.3.tar.gz
cd postgresql-16.3
./configure --prefix=/usr/local/pgsql
make -j$(nproc)
make install

# 第四步：创建用户和数据目录
useradd postgres
mkdir -p /data/pgdata
chown postgres:postgres /data/pgdata

# 第五步：初始化数据库
su - postgres -c "/usr/local/pgsql/bin/initdb -D /data/pgdata"

# 第六步：启动数据库
su - postgres -c "/usr/local/pgsql/bin/pg_ctl -D /data/pgdata -l /data/pgdata/logfile start"

# 第七步：验证
su - postgres -c "/usr/local/pgsql/bin/psql -c 'SELECT version();'"


# 第八步 把 '请改成强密码' 换成你的强密码（与后面 .env 一致）：
su - postgres -c "/usr/local/pgsql/bin/psql" <<'EOF'
CREATE USER myblog WITH PASSWORD '请改成强密码';
CREATE DATABASE myblog OWNER myblog ENCODING 'UTF8';
GRANT ALL PRIVILEGES ON DATABASE myblog TO myblog;
\c myblog
GRANT ALL ON SCHEMA public TO myblog;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO myblog;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO myblog;
EOF
# 测试连接：
PGPASSWORD='请改成强密码' /usr/local/pgsql/bin/psql -h 127.0.0.1 -U myblog -d myblog -c "SELECT 1;"
```

### 4. 数据库迁移

```bash
cd /root/project/FengYuWorld/backend
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
go run ./cmd/migrate
# 看到 migrations completed 表示表结构 OK。
# 或: cd backend && go run ./cmd/migrate
```

### 5. 启动后端

```bash
# 临时启动
make backend-run
# 或: cd backend && go run ./cmd/server
cd /root/project/FengYuWorld/backend
go run ./cmd/server
# 另开终端验证：
curl http://127.0.0.1:8080/api/health


# 用 systemd 后台跑
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


systemctl status myblog    # 看状态
journalctl -u myblog -f    # 看日志
systemctl restart myblog   # 重启
systemctl stop myblog      # 停止

# 临时方案：nohup（不想配 systemd 时）
cd /root/project/FengYuWorld/backend
nohup ./bin/server > server.log 2>&1 &

# 查看是否在跑：
ps aux | grep bin/server
tail -f server.log

# 停止：
kill $(pgrep -f bin/server)
```

API 默认监听 `http://localhost:8080`。

### 6. 启动前端

```bash
# 在你 Windows 开发机
cd D:\Code\Go\My-Blog\frontend
npm install
npm run build
scp -r dist root@你的服务器IP:/var/www/blog


# 或: cd frontend && npm install && npm run dev
```

前端默认 `http://localhost:5173`，已配置 `/api` 与 `/uploads` 代理到后端。

### 7. 配置nginx

```bash
mkdir -p /var/www/myblog /etc/nginx/conf.d/disabled
# 上传或复制 deploy/nginx-ip-paths.conf
cp /root/project/FengYuWorld/deploy/nginx-ip-paths.conf /etc/nginx/conf.d/myblog.conf
# 禁用旧的 relay 配置（避免冲突）
mv /etc/nginx/conf.d/jian-guo-yun-relay.conf /etc/nginx/conf.d/disabled/
nginx -t
systemctl reload nginx
# 以后加新服务
# 在 deploy/nginx-ip-paths.conf 里加新的 location，例如：

 location /other-service/ {
     proxy_pass http://127.0.0.1:端口/;
     }
# 原则： 每个服务占一个路径前缀，不要两个服务同时抢 /。
```

### 8. 服务器更新

```bash
# 服务器上更新：
cd /root/project/FengYuWorld
git pull
cd backend
go run ./cmd/migrate
go build -o bin/server ./cmd/server
systemctl restart myblog

#前端传到 /var/www/blog/ 目录下
```

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

前端默认部署在 `**/blog/**` 子路径（见 `frontend/vite.config.ts` 的 `base`）。

| 路径           | 说明                                                      |
| ------------ | ------------------------------------------------------- |
| `/blog/`     | 前端 SPA                                                  |
| `/blog/api/` | 反代 Go API                                               |
| `/uploads/`  | 上传文件                                                    |
| `/webdav/`   | 其他服务（如坚果云 relay，Nginx 示例见 `deploy/nginx-ip-paths.conf`） |

```bash
# 本机构建并上传
cd frontend && npm run build
scp -r dist/* root@YOUR_IP:/var/www/blog/

# 服务器 Nginx
cp deploy/nginx-ip-paths.conf /etc/nginx/conf.d/myblog.conf
nginx -t && systemctl reload nginx
```

访问：`http://YOUR_IP/blog/`  
本地开发：`http://localhost:5173/blog/`