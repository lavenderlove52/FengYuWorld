# Windows 使用 Docker 启动 PostgreSQL（可选）

本文档说明在 **Windows + Docker Desktop** 下，用项目自带 `docker-compose.yml` 启动 PostgreSQL 16。  
若 WSL 更新失败或不想用 Docker，请改用 [Windows 本机安装](./INSTALL_POSTGRESQL_WINDOWS.md)。

---

## 一、前提

- 已安装 **Docker Desktop**
- WSL2 已更新（否则会提示 `WSL needs updating`，需执行 `wsl --update`）
- 项目路径：`D:\Code\Go\My-Blog`

---

## 二、数据挂载到本机目录（推荐）

编辑项目根目录 `docker-compose.yml`，将命名卷改为绑定挂载，例如：

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: myblog-postgres
    environment:
      POSTGRES_USER: myblog
      POSTGRES_PASSWORD: myblog
      POSTGRES_DB: myblog
    ports:
      - "5432:5432"
    volumes:
      - D:/Data/myblog/postgres:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U myblog -d myblog"]
      interval: 5s
      timeout: 5s
      retries: 5
```

说明：

- Windows 路径使用 **`D:/Data/...`** 正斜杠。
- 先创建空目录：`D:\Data\myblog\postgres`
- 删除文件末尾的 `volumes: postgres_data:` 段（若已改用绑定挂载）。

若使用项目内目录（仓库内已有 `postgres/` 数据时）：

```yaml
volumes:
  - ./postgres:/var/lib/postgresql/data
```

---

## 三、启动命令

PowerShell：

```powershell
cd D:\Code\Go\My-Blog

# 若使用 D:\Data\myblog\postgres
New-Item -ItemType Directory -Force -Path D:\Data\myblog\postgres

docker compose up -d postgres
docker compose ps
docker compose logs -f postgres
```

停止（保留数据）：

```powershell
docker compose down
```

---

## 四、连接串

`backend\.env`：

```env
DATABASE_URL=postgres://myblog:myblog@localhost:5432/myblog?sslmode=disable
```

与 `docker-compose.yml` 中 `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` 一致。

迁移与启动：

```powershell
cd D:\Code\Go\My-Blog\backend
go run ./cmd/migrate
go run ./cmd/server
```

---

## 五、注意

- 本机已安装 PostgreSQL 时，不要与容器同时占用 **5432**。
- 绑定挂载目录不要混用不同大版本 PostgreSQL 的数据文件。

---

## 相关文档

- [Windows 本机安装（不用 Docker）](./INSTALL_POSTGRESQL_WINDOWS.md)
- [CentOS 7 安装](./INSTALL_POSTGRESQL_CENTOS7.md)
