# CentOS 7 安装 PostgreSQL 16（不用 Docker）

本文档说明在 **CentOS 7 x86_64** 本机/服务器上安装、启动 PostgreSQL 16，并为 **My-Blog Go 项目** 创建数据库与用户。

> 推荐版本：**PostgreSQL 16**（与项目 `docker-compose.yml`、设计文档一致）。  
> CentOS 7 已 EOL，若 `yum` 源异常，请先按本文「安装前准备」切换 vault 源。

---

## 一、安装前准备

### 1.1 系统要求

| 项 | 建议 |
|----|------|
| 系统 | CentOS 7 x86_64 |
| 内存 | ≥ 1GB（4G 轻量云可运行，建议调小内存参数） |
| 磁盘 | 数据目录建议独立，如 `/data/pgsql/16/data` |
| 权限 | root 或 sudo |

### 1.2 检查端口与旧版本

```bash
ss -lntp | grep 5432
rpm -qa | grep postgres
```

若 5432 已被占用，需先停止旧实例或改端口。

### 1.3 CentOS 7 源不可用（vault）

```bash
sudo sed -i 's|^mirrorlist=|#mirrorlist=|g' /etc/yum.repos.d/CentOS-*.repo
sudo sed -i 's|^#baseurl=|baseurl=|g' /etc/yum.repos.d/CentOS-*.repo
sudo sed -i 's|http://mirror.centos.org|http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*.repo
sudo yum clean all
sudo yum makecache
```

---

## 二、安装 PostgreSQL 16（PGDG 官方源）

### 2.1 安装仓库 RPM

```bash
sudo yum install -y https://download.postgresql.org/pub/repos/yum/reporpms/EL-7-x86_64/pgdg-redhat-repo-latest.noarch.rpm
```

### 2.2 安装服务端与客户端

```bash
sudo yum install -y postgresql16-server postgresql16
```

安装后路径：

| 路径 | 说明 |
|------|------|
| `/usr/pgsql-16/bin/` | `psql`、`pg_dump` 等 |
| `/var/lib/pgsql/16/data/` | 默认数据目录 |
| systemd 服务名 | `postgresql-16` |

### 2.3 初始化（仅首次）

```bash
sudo /usr/pgsql-16/bin/postgresql-16-setup initdb
```

---

## 三、启动与开机自启

```bash
sudo systemctl start postgresql-16
sudo systemctl enable postgresql-16
sudo systemctl status postgresql-16
```

常用：

```bash
sudo systemctl stop postgresql-16
sudo systemctl restart postgresql-16
sudo systemctl reload postgresql-16
```

验证：

```bash
sudo -u postgres /usr/pgsql-16/bin/psql -c "SELECT version();"
ss -lntp | grep 5432
```

---

## 四、创建业务库与用户（对齐 My-Blog）

项目默认连接串见根目录 `.env.example`：

```text
postgres://myblog:<密码>@127.0.0.1:5432/myblog?sslmode=disable
```

**请将 `<密码>` 改为强密码。**

```bash
sudo -u postgres /usr/pgsql-16/bin/psql <<'EOF'
CREATE USER myblog WITH PASSWORD '请改成强密码';
CREATE DATABASE myblog OWNER myblog ENCODING 'UTF8';
GRANT ALL PRIVILEGES ON DATABASE myblog TO myblog;
\c myblog
GRANT ALL ON SCHEMA public TO myblog;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO myblog;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO myblog;
EOF
```

测试连接：

```bash
PGPASSWORD='请改成强密码' /usr/pgsql-16/bin/psql -h 127.0.0.1 -U myblog -d myblog -c "SELECT 1;"
```

---

## 五、访问控制

### 5.1 仅本机访问（推荐默认）

```bash
sudo vi /var/lib/pgsql/16/data/postgresql.conf
```

```ini
listen_addresses = 'localhost'
port = 5432
```

```bash
sudo vi /var/lib/pgsql/16/data/pg_hba.conf
```

确保包含：

```text
local   all             all                                     peer
host    all             all             127.0.0.1/32            md5
host    all             all             ::1/128                 md5
```

```bash
sudo systemctl reload postgresql-16
```

### 5.2 允许内网远程（可选）

`postgresql.conf`：

```ini
listen_addresses = '*'
```

`pg_hba.conf`（示例：仅允许 10.0.0.0/24 访问库 myblog）：

```text
host    myblog    myblog    10.0.0.0/24    md5
```

防火墙：

```bash
sudo firewall-cmd --permanent --add-service=postgresql
sudo firewall-cmd --reload
```

云厂商安全组仅放行可信 IP，**不要**对公网 0.0.0.0/0 开放 5432。

---

## 六、4G 内存简要调优（可选）

编辑 `/var/lib/pgsql/16/data/postgresql.conf`：

```ini
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB
max_connections = 50
```

```bash
sudo systemctl restart postgresql-16
```

---

## 七、数据目录迁到独立盘（可选）

```bash
sudo systemctl stop postgresql-16
sudo mkdir -p /data/pgsql/16/data
sudo chown -R postgres:postgres /data/pgsql
sudo rsync -av /var/lib/pgsql/16/data/ /data/pgsql/16/data/
sudo mv /var/lib/pgsql/16/data /var/lib/pgsql/16/data.bak
sudo ln -s /data/pgsql/16/data /var/lib/pgsql/16/data
sudo systemctl start postgresql-16
```

新机器更推荐：按 PGDG 文档设置 `PGDATA` 后直接在新路径 `initdb`。

---

## 八、与 Go 项目联调

在部署后端的服务器上：

```bash
cd /path/to/My-Blog/backend
export DATABASE_URL='postgres://myblog:你的密码@127.0.0.1:5432/myblog?sslmode=disable'
go run ./cmd/migrate
go run ./cmd/server
```

检查表：

```bash
sudo -u postgres /usr/pgsql-16/bin/psql -d myblog -c "\dt"
```

---

## 九、日常运维

```bash
# 备份
sudo -u postgres /usr/pgsql-16/bin/pg_dump -Fc myblog > /backup/myblog_$(date +%F).dump

# 查看连接
sudo -u postgres /usr/pgsql-16/bin/psql -c \
  "SELECT pid, usename, datname, client_addr, state FROM pg_stat_activity;"
```

---

## 十、常见问题

| 现象 | 处理 |
|------|------|
| `yum` 无法安装 PGDG | `yum repolist \| grep pgdg`；检查 vault 源 |
| 启动失败 | `journalctl -u postgresql-16 -n 50` |
| `connection refused` | 服务是否启动、`listen_addresses`、防火墙、安全组 |
| 密码错误 | 用户密码、`pg_hba.conf` 认证方式 |
| schema 权限不足 | 重做第四节 `GRANT ON SCHEMA public` |

---

## 十一、一键安装顺序（汇总）

```bash
sudo yum install -y https://download.postgresql.org/pub/repos/yum/reporpms/EL-7-x86_64/pgdg-redhat-repo-latest.noarch.rpm
sudo yum install -y postgresql16-server postgresql16
sudo /usr/pgsql-16/bin/postgresql-16-setup initdb
sudo systemctl enable postgresql-16
sudo systemctl start postgresql-16
sudo -u postgres /usr/pgsql-16/bin/psql -c "CREATE USER myblog WITH PASSWORD '强密码';"
sudo -u postgres /usr/pgsql-16/bin/psql -c "CREATE DATABASE myblog OWNER myblog ENCODING 'UTF8';"
```

---

## 相关文档

- [Windows 本机安装 PostgreSQL](./INSTALL_POSTGRESQL_WINDOWS.md)
- [使用 Docker 启动数据库](./INSTALL_POSTGRESQL_DOCKER.md)（若已编写）
- 项目根目录 `README.md`
