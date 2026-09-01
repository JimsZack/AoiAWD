# GoAWD

GoAWD 是 [AoiAWD](https://github.com/Aodzip/AoiAWD) 的 Golang 重写版本，专为 CTF AWD (Attack With Defense) 模式设计的轻量级 EDR (Endpoint Detection and Response) 系统。

## 特性

- **零外部依赖**: 纯 Go 标准库实现，`go build` 即可编译
- **单二进制部署**: 无需 PHP 运行时，编译后直接运行
- **低内存占用**: ~15MB（原版 PHP ~50MB）
- **跨平台编译**: 支持 Linux/macOS/Windows 交叉编译
- **静态编译**: 探针可编译为静态二进制，无外部依赖
- **双存储后端**: 内存存储（默认）+ JSON 文件持久化（可选）

## 架构

```
┌─────────────────────────────────────────────────────────┐
│                    GoAWD Frontend                        │
│               (Vue.js, 保持不变)                         │
└─────────────────────────┬───────────────────────────────┘
                          │ HTTP / WebSocket
┌─────────────────────────▼───────────────────────────────┐
│                    GoAWD Core Server                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐ │
│  │ HTTP Server  │  │ TCP Receiver│  │ WebSocket Hub   │ │
│  │ (ServeMux)  │  │ (port 8023) │  │ (real-time)     │ │
│  │ (port 1337) │  │             │  │                 │ │
│  └──────┬──────┘  └──────┬──────┘  └────────┬────────┘ │
│         │                │                   │          │
│         └────────┬───────┴───────────────────┘          │
│                  │                                      │
│         ┌────────▼────────┐  ┌─────────────────┐       │
│         │ Plugin Manager  │  │  Storage (mem)  │       │
│         └─────────────────┘  └─────────────────┘       │
└─────────────────────────────────────────────────────────┘

探针 (运行在靶机上):
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  RoundWorm-Go    │  │  Guardian-Go     │  │  TapeWorm (PHP)  │
│  (文件/进程监控)  │  │ (PWN 流量劫持)  │  │  (Web 流量监控)  │
│  (syscall)       │  │  (管道模式)      │  │  (可选)          │
└────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘
         └─────────────────────┴──────────────────────┘
                          │ TCP → Core:8023
```

## 快速开始

### 环境要求

- Go 1.21+

### 编译

```bash
make build
# 或直接
go build -o bin/goawd-server ./cmd/server
go build -o bin/goawd-roundworm ./cmd/roundworm
go build -o bin/goawd-guardian ./cmd/guardian
```

### 运行服务器

```bash
./bin/goawd-server \
    -http "0.0.0.0:1337" \
    -tcp "0.0.0.0:8023" \
    -token "your-token" \
    -storage memory
```

### 命令行参数

#### GoAWD Server

```bash
./goawd-server [OPTIONS]

Options:
  -http string     HTTP 服务监听地址 (default "0.0.0.0:1337")
  -tcp string      TCP 接收器监听地址 (default "0.0.0.0:8023")
  -token string    访问令牌 (留空则随机生成)
  -storage string  存储后端: memory|file (default "memory")
  -file-path string JSON 文件路径 (default "./goawd.json", file 后端使用)
  -plugins string  插件目录 (default "./plugins")
  -public string   前端静态文件目录 (default "./public")
  -v               显示版本
```

#### RoundWorm 探针

```bash
./goawd-roundworm [OPTIONS]

Options:
  -s string    服务器地址 (default "127.0.0.1")
  -p int       服务器端口 (default 8023)
  -w string    监控目录，分号分隔 (default "/tmp")
  -i duration  进程扫描间隔 (default 100ms)
  -d           守护进程模式
```

#### Guardian 探针

```bash
./goawd-guardian [OPTIONS]

Options:
  -binary string  PWN 二进制路径 (必填)
  -host string    服务器地址 (default "127.0.0.1")
  -port int       服务器端口 (default 8023)
```

### Docker 部署

```bash
docker-compose up -d
```

## 内置插件

- **FlagBuster**: 检测 Web/PWN 输出中的 flag 字段，替换为随机 flag 并告警
- **KingWatcher**: 监控赛点文件（默认 `/flag`，可通过 `GOAWD_KING_FILE` 环境变量配置）
- **ZombieKiller**: 检测不死马行为（文件删除后快速重建）

## 文档

详细文档请参阅 [docs/](docs/) 目录：

- [项目概述](docs/01-项目概述.md)
- [架构设计](docs/02-架构设计.md)
- [API 接口文档](docs/03-API接口文档.md)
- [数据结构定义](docs/04-数据结构定义.md)
- [模块实现指南](docs/05-模块实现指南.md)
- [部署文档](docs/06-部署文档.md)
- [迁移指南](docs/07-迁移指南.md)
- [开发计划](docs/08-开发计划.md)

## 与 AoiAWD 的区别

| 特性 | AoiAWD | GoAWD |
|------|--------|-------|
| 语言 | PHP + C | Go |
| 部署 | 需要 PHP 运行时 | 单二进制文件 |
| 外部依赖 | MongoDB + PHP | 零外部依赖 |
| 启动速度 | ~2s | <100ms |
| 内存占用 | ~50MB | ~15MB |
| 插件系统 | 运行时加载 (.php) | 编译时链接 (.go) |
| 探针编译 | 需要编译环境 | 静态编译，无依赖 |
| 存储后端 | MongoDB | 内存 / JSON 文件 |

## 兼容性

- **API 兼容**: 所有 REST API 接口与原版保持一致
- **协议兼容**: TCP 探针通信协议保持一致
- **前端兼容**: Vue.js 前端无需修改

## 许可证

GNU Affero General Public License v3.0

## 致谢

感谢 [AoiAWD](https://github.com/Aodzip/AoiAWD) 项目作者 Aodzip 的优秀设计。
