# GoAWD — 轻量级 EDR for CTF AWD

<p align="center">
  <img src="https://img.shields.io/badge/version-2.0.0-blue.svg" alt="version">
  <img src="https://img.shields.io/badge/license-AGPL--3.0-green.svg" alt="license">
  <img src="https://img.shields.io/badge/go-1.21+-00ADD8.svg" alt="go version">
  <img src="https://img.shields.io/badge/vue-2.x-brightgreen.svg" alt="vue version">
</p>

GoAWD 是专为 CTF AWD（Attack With Defense）模式设计的轻量级 EDR（端点检测与响应）系统。

## 功能特性

- **零外部依赖**：纯 Go 标准库实现，无需 PHP、MongoDB 等外部服务
- **单二进制部署**：编译后直接运行，内存占用 ~15MB
- **Web 流量监控**：捕获 HTTP 请求/响应，支持实时流量篡改
- **PWN 流量捕获**：劫持 PWN 程序 stdin/stdout，记录内存映射
- **文件系统监控**：基于 inotify 的实时文件变更检测
- **进程行为监控**：扫描 /proc 检测新进程创建
- **内置插件系统**：FlagBuster、KingWatcher、ZombieKiller
- **WebSocket 实时推送**：数据变更即时通知前端
- **可视化 Web 面板**：现代化 Vue.js 管理界面

## 快速开始

### 方式一：直接运行

```bash
# 编译所有组件
cd GoAWD
make build

# 运行服务器
./bin/goawd-server -storage file -file-path "./goawd.json"

# 运行探针 (靶机上)
./bin/goawd-roundworm -s <选手IP> -d -w "/tmp;/var/www/html"
```

### 方式二：Docker

```bash
# 构建并启动
docker compose up -d

# 或使用启动脚本
chmod +x docker_goawd_start.sh
./docker_goawd_start.sh
```

## 界面预览

### 登录页面

采用现代化设计，渐变背景 + 毛玻璃效果，支持 Access Token 登录。

### Dashboard

实时显示系统状态、报警统计、插件列表等关键信息。

### Web 日志

支持按 HTTP 方法、IP 地址过滤，双击查看请求详情。

### PWN 日志

实时捕获 PWN 程序的 stdin/stdout 流量，支持 Hex Dump 视图。

## 命令行参数

### goawd-server

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-http` | `0.0.0.0:1337` | HTTP 服务监听地址 |
| `-tcp` | `0.0.0.0:8023` | TCP 接收器监听地址 |
| `-storage` | `memory` | 存储后端: memory/file |
| `-file-path` | `./goawd.json` | JSON 文件路径 (file 存储) |
| `-token` | 随机生成 | API 访问令牌 |
| `-plugins` | `./plugins` | 插件目录 |
| `-public` | `./public` | 前端静态文件目录 |

### goawd-roundworm

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-s` | `127.0.0.1` | 服务器地址 |
| `-p` | `8023` | 服务器端口 |
| `-w` | `/tmp` | 监控目录，多个用 ; 分隔 |
| `-i` | `100` | 进程扫描间隔 (ms) |
| `-d` | `false` | 守护进程模式 |

### goawd-guardian

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-binary` | (必填) | PWN 二进制路径 |
| `-host` | `127.0.0.1` | 服务器地址 |
| `-port` | `8023` | 服务器端口 |

## 架构

```
选手主机:
  └── goawd-server  (Web UI :1337 + 探针接收 :8023)

靶机:
  ├── goawd-roundworm  (文件系统/进程监控)
  ├── goawd-guardian   (PWN 流量劫持)
  └── goawd-tapeworm   (PHP Web 流量注入)
```

## 探针通信协议

探针通过 TCP 连接到服务器，使用 JSON 格式通信（换行分隔）：

| 类型 | 方向 | 数据 |
|------|------|------|
| `web` | 探针→服务器 | Web 请求/响应数据 |
| `pwn` | 探针→服务器 | PWN 程序 stdin/stdout + 内存映射 |
| `new_process` | 探针→服务器 | 新进程信息 |
| `pid_list` | 探针→服务器 | 当前存活 PID 列表 |
| `file` | 探针→服务器 | 文件系统事件 |
| `ping` | 探针→服务器 | 心跳 |

## 插件开发

插件使用 Go 编译为 `.so` 文件，放到 `plugins/` 目录即可热加载。

```go
package main

import "plugin"

func init() {
    // 注册钩子
}
```

## 内置插件

| 插件 | 功能 |
|------|------|
| FlagBuster | 检测并替换 Web/PWN 输出中的 flag |
| KingWatcher | KoH 模式赛点文件监控 |
| ZombieKiller | 不死马行为检测 |

## 项目结构

```
GoAWD/
├── cmd/                # 入口程序
│   ├── server/         # 核心服务器
│   ├── roundworm/      # 文件/进程探针
│   ├── guardian/       # PWN 探针
│   └── tapeworm/       # PHP Web 注入器
├── internal/           # 内部包
│   ├── api/            # REST API
│   ├── config/         # 配置管理
│   ├── core/           # 核心逻辑
│   ├── plugin/         # 插件管理器
│   ├── storage/        # 存储后端
│   ├── types/          # 数据结构
│   └── ws/             # WebSocket
├── pkg/                # 公共包
│   ├── inotify/        # 文件系统监控
│   ├── proc/           # 进程扫描
│   └── tcpclient/      # TCP 客户端
├── plugins/            # 内置插件
└── docs/               # 项目文档

Frontend/
├── src/
│   ├── views/          # 页面组件
│   ├── api/            # API 接口
│   ├── styles/         # 样式文件
│   └── vuex/           # 状态管理
└── static/             # 静态资源
```

## 测试

```bash
# 运行所有测试
cd GoAWD
go test ./...

# 查看测试覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 文档

详细文档请参阅 `GoAWD/docs/` 目录：

- [项目概述](GoAWD/docs/01-项目概述.md)
- [架构设计](GoAWD/docs/02-架构设计.md)
- [API 接口文档](GoAWD/docs/03-API接口文档.md)
- [数据结构定义](GoAWD/docs/04-数据结构定义.md)
- [模块实现指南](GoAWD/docs/05-模块实现指南.md)
- [部署文档](GoAWD/docs/06-部署文档.md)
- [迁移指南](GoAWD/docs/07-迁移指南.md)
- [开发计划](GoAWD/docs/08-开发计划.md)

## 许可证

GNU AGPL-3.0
