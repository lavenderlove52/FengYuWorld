# Windows 本机安装 PostgreSQL 16（不用 Docker）

本文档说明在 **Windows 10/11** 上安装、启动 PostgreSQL 16，配置 **My-Blog Go 项目**（`D:\Code\Go\My-Blog`）的数据库连接与迁移。

> 若 Docker Desktop / WSL 更新卡住，**优先用本文方式**，不依赖 WSL。  
> 推荐版本：**PostgreSQL 16.x**（与项目设计一致）。

---

## 一、安装前准备

### 1.1 需要已安装的其它环境

| 软件 | 用途 |
|------|------|
| Go 1.22+ | 运行后端、`go run ./cmd/migrate` |
| Node.js 18+ | 前端 `npm run dev`（与数据库无关） |

### 1.2 检查 5432 是否被占用

PowerShell：

```powershell
netstat -ano | findstr :5432
```

若已被 Docker 或其它 PostgreSQL 占用，需先停止对应服务，或安装时改用其它端口（不推荐，需同步改 `.env`）。

---

## 二、下载并安装

### 2.1 官方安装包（推荐）

1. 打开：https://www.postgresql.org/download/windows/
2. 进入 **EDB installer**（或 Stack Builder 同发行方安装器）
3. 下载 **PostgreSQL 16.x** Windows x86-64
4. 运行安装向导

### 2.2 安装向导建议选项

| 步骤 | 建议 |
|------|------|
| Components | PostgreSQL Server、pgAdmin 4、Command Line Tools |
| Data Directory | 默认 `C:\Program Files\PostgreSQL\16\data`，或自定义如 `D:\Data\pgsql\16\data` |
| Password | 为 **超级用户 `postgres`** 设置强密码（务必记住） |
| Port | **5432**（与项目默认一致） |
| Locale | 默认或 Chinese, China |

安装完成后，PostgreSQL 会注册为 **Windows 服务**，一般自动启动。

### 2.3 使用 winget（可选）

```powershell
winget install PostgreSQL.PostgreSQL.16
```

安装后确认服务名与安装路径，下文以 `C:\Program Files\PostgreSQL\16` 为例。

---

## 三、确认服务已启动

### 3.1 服务管理器

1. `Win + R` → `services.msc`
2. 找到 **postgresql-x64-16**（名称可能略有不同）
3. 状态应为 **正在运行**，启动类型 **自动**

PowerShell（管理员）：

```powershell
Get-Service *postgres*
Start-Service postgresql-x64-16   # 若未运行
```

### 3.2 命令行验证

将 PostgreSQL 的 `bin` 加入 PATH（安装程序通常已加）：

```text
C:\Program Files\PostgreSQL\16\bin
```

PowerShell：

```powershell
psql --version
psql -U postgres -c "SELECT version();"
```

按提示输入安装时设置的 **postgres** 超级用户密码。

---

## 四、创建业务库与用户

与项目 `.env.example` 对齐：

```text
DATABASE_URL=postgres://myblog:<密码>@localhost:5432/myblog?sslmode=disable
```

### 4.1 使用 psql（推荐）

PowerShell：

```powershell
psql -U postgres
```

在 `psql` 里执行（**请修改密码**）：

```sql
CREATE USER myblog WITH PASSWORD 'myblog_password_here';
CREATE DATABASE myblog OWNER myblog ENCODING 'UTF8';
GRANT ALL PRIVILEGES ON DATABASE myblog TO myblog;
\c myblog
GRANT ALL ON SCHEMA public TO myblog;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO myblog;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO myblog;
\q
```

测试：

```powershell
psql -h localhost -U myblog -d myblog -c "SELECT 1;"
```

### 4.2 使用 pgAdmin 4（图形界面）

1. 打开 pgAdmin，连接本地 Server（postgres 用户）
2. 登录 → 右键 **Login/Group Roles** → Create → Role：名 `myblog`，密码页设置密码
3. 右键 **Databases** → Create → Database：名 `myblog`，Owner 选 `myblog`

---

## 五、配置 Go 项目

### 5.1 复制环境变量

```powershell
cd D:\Code\Go\My-Blog
copy .env.example backend\.env
```

编辑 `D:\Code\Go\My-Blog\backend\.env`：

```env
DATABASE_URL=postgres://myblog:myblog_password_here@localhost:5432/myblog?sslmode=disable
JWT_ACCESS_SECRET=change-me-access-secret-min-32-chars-long
JWT_REFRESH_SECRET=change-me-refresh-secret-min-32-chars-long
```

`JWT_*` 请改为足够长的随机字符串。

### 5.2 执行迁移

```powershell
cd D:\Code\Go\My-Blog\backend
go run ./cmd/migrate
```

成功后再启动后端：

```powershell
go run ./cmd/server
```

另开终端启动前端：

```powershell
cd D:\Code\Go\My-Blog\frontend
npm install
npm run dev
```

浏览器访问：http://localhost:5173

---

## 六、与 Docker 二选一

| 方式 | 说明 |
|------|------|
| **本机 PostgreSQL（本文）** | 不依赖 WSL；数据在 `C:\Program Files\PostgreSQL\16\data` 或你自定义目录 |
| **Docker Compose** | 见项目根目录 `docker-compose.yml`，需 Docker Desktop + WSL2 正常 |

**不要同时**让本机 PostgreSQL 与 Docker 容器都监听 `5432`，会冲突。

若已用 Docker 且数据在 `D:\Code\Go\My-Blog\postgres\`，那是容器数据目录，与本机安装器的数据目录不是同一套，勿混用。

---

## 七、IDEA / 开发机常用操作

```powershell
# 迁移
cd D:\Code\Go\My-Blog\backend
go run ./cmd/migrate

# 后端
go run ./cmd/server

# 前端
cd D:\Code\Go\My-Blog\frontend
npm run dev
```

IDEA 打开目录：`D:\Code\Go\My-Blog\backend`，SDK 选 **Go**，不要选 Java JDK。

---

## 八、防火墙与远程访问

- **仅本机开发**：默认 `localhost` 即可，一般无需改防火墙。
- **局域网其它机器连库**：需在 `postgresql.conf` 中设置 `listen_addresses`，并修改 `pg_hba.conf`；Windows 防火墙放行 5432。**生产环境勿对公网开放 5432。**

配置文件位置（默认）：

```text
C:\Program Files\PostgreSQL\16\data\postgresql.conf
C:\Program Files\PostgreSQL\16\data\pg_hba.conf
```

修改后重启服务：

```powershell
Restart-Service postgresql-x64-16
```

---

## 九、备份与卸载注意

### 9.1 备份

```powershell
pg_dump -U myblog -Fc -f D:\Backup\myblog.dump myblog
```

### 9.2 卸载

控制面板卸载 PostgreSQL 会提示是否删除数据目录；**删前请先备份** `data` 目录或 `pg_dump`。

---

## 十、常见问题

| 现象 | 处理 |
|------|------|
| `connection refused` on 5432 | 检查 Windows 服务是否运行；`netstat -ano \| findstr 5432` |
| `password authentication failed` | 核对 `myblog` 密码与 `DATABASE_URL` |
| `go run ./cmd/migrate` 失败 | 确认 `backend\.env` 存在且 `DATABASE_URL` 正确 |
| 与 Docker 冲突 | 停 Docker 容器：`docker compose down`，或改其一端口 |
| `psql` 不是内部命令 | 将 `PostgreSQL\16\bin` 加入系统 PATH |
| 安装后无法连接 | 是否用了安装时的 postgres 密码；pg_hba 是否允许本地 md5 |

---

## 十一、检查清单

```powershell
psql --version
psql -h localhost -U myblog -d myblog -c "SELECT 1;"
cd D:\Code\Go\My-Blog\backend
go run ./cmd/migrate
go run ./cmd/server
```

全部通过即表示 Windows 本机 PostgreSQL 与项目已打通。

---

## 相关文档

- [CentOS 7 安装 PostgreSQL](./INSTALL_POSTGRESQL_CENTOS7.md)
- 项目根目录 `README.md`、`.env.example`
