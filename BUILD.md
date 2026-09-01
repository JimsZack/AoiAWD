# GoAWD 构建指南

GoAWD 是 AoiAWD 的 Go 语言重写版，无需 MongoDB 和 PHP 依赖。

## 快速构建

### 方式一：Make（推荐）

```bash
cd GoAWD
make build
```

构建产物在 `bin/` 目录：
- `goawd-server` — 中心服务器
- `goawd-roundworm` — 文件系统/进程探针
- `goawd-guardian` — PWN 探针
- `goawd-tapeworm` — Web 探针（PHP 注入器）

### 方式二：Docker

```bash
docker compose up -d
```

构建的镜像包含所有组件和前端。

## 各组件构建

### 1. 构建 Frontend

```bash
cd Frontend
npm install --ignore-scripts --legacy-peer-deps
npm run build
```

### 2. 构建 GoAWD Server

```bash
cd GoAWD
go build -o bin/goawd-server ./cmd/server
```

### 3. 构建 RoundWorm

```bash
cd GoAWD
CGO_ENABLED=0 go build -o bin/goawd-roundworm ./cmd/roundworm
```

### 4. 构建 Guardian

```bash
cd GoAWD
CGO_ENABLED=0 go build -o bin/goawd-guardian ./cmd/guardian
```

### 5. 构建 TapeWorm

```bash
cd GoAWD
CGO_ENABLED=0 go build -o bin/goawd-tapeworm ./cmd/tapeworm
```

## 交叉编译

```bash
# Linux amd64
make build-linux-amd64

# Linux arm64
make build-linux-arm64
```

## 运行

### 服务器端（选手主机）

```bash
./goawd-server
# 输出 Access Token，用于 Web 面板登录
```

### 探针端（靶机）

```bash
# 文件系统/进程监控
./goawd-roundworm -s <选手IP> -d -w "/tmp;/var/www/html"

# Web 流量监控
./goawd-tapeworm -d /var/www/html -s <选手IP>:8023
```

## 命令行参数

### goawd-server

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-http` | HTTP 服务地址 | `0.0.0.0:1337` |
| `-tcp` | TCP 接收器地址 | `0.0.0.0:8023` |
| `-storage` | 存储后端 (memory/file) | `memory` |
| `-file-path` | JSON 存储路径 | `./goawd.json` |
| `-public` | 前端静态文件目录 | `./public` |
| `-plugins` | 插件目录 | `./plugins` |
| `-token` | 访问令牌 | 随机生成 |

### goawd-roundworm

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-s` | 服务器地址 | `127.0.0.1` |
| `-p` | 服务器端口 | `8023` |
| `-w` | 监控目录 | `/tmp` |
| `-d` | 守护进程模式 | `false` |
| `-i` | 进程扫描间隔 (ms) | `100` |

### goawd-tapeworm

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-d` | Web 根目录 | — |
| `-s` | 服务器地址 | `127.0.0.1:8023` |
| `-m` | 模式 (inject/remove/test) | `inject` |

## 打包

```bash
# 打包二进制版本
./build.sh

# 打包源码
./package_source.sh
```

## 与原版 AoiAWD 的区别

| 维度 | AoiAWD (PHP) | GoAWD (Go) |
|------|-------------|------------|
| 依赖 | MongoDB + PHP | 无外部依赖 |
| 存储 | MongoDB | Memory/File |
| 插件 | PHP include | Go plugin (.so) |
| 部署 | 需要 PHP 环境 | 单二进制文件 |