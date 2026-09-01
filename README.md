# GoAWD — 轻量级 EDR for CTF AWD

GoAWD 是 AoiAWD 的 Go 语言重写版，专为 CTF AWD 模式设计的轻量级 EDR（端点检测与响应）系统。

## 与原版 AoiAWD 的区别

| 维度 | AoiAWD (PHP) | GoAWD (Go) |
|------|-------------|------------|
| 语言 | PHP 7.2 + C | Go 1.21 |
| 存储 | MongoDB | 内置 Memory/File |
| 插件 | PHP include | Go plugin (.so) |
| 依赖 | MongoDB + PHP | 无外部依赖 |
| 部署 | 需要 PHP 环境 | 单二进制文件 |

## 功能特性

- Web 输入输出数据捕获与流量篡改
- PWN 类题目流量包捕获与内存快照
- 服务器进程/文件系统行为监控
- 内置插件: FlagBuster、KingWatcher、ZombieKiller
- WebSocket 实时推送
- 可视化 Web 面板

## 快速开始

### 方式一：直接运行

```bash
# 编译
cd GoAWD
go build -o goawd-server ./cmd/server
go build -o goawd-roundworm ./cmd/roundworm
go build -o goawd-guardian ./cmd/guardian

# 运行服务器
./goawd-server

# 运行探针 (靶机上)
./goawd-roundworm -s <选手IP> -d -w "/tmp;/var/www/html"
```

### 方式二：Docker

```bash
docker compose up -d
```

## 命令行参数

### goawd-server

```
Usage of goawd-server:
  -db string        数据库名称 (default "aoiawd")
  -file-path string JSON 存储文件路径 (default "./goawd.json")
  -http string      HTTP 服务绑定地址 (default "0.0.0.0:1337")
  -plugins string   插件目录 (default "./plugins")
  -public string    前端静态文件目录 (default "./public")
  -storage string   存储后端: memory|file (default "memory")
  -tcp string       TCP 接收器绑定地址 (default "0.0.0.0:8023")
  -token string     访问令牌 (留空则随机生成)
  -v                显示版本
```

### goawd-roundworm

```
Usage of goawd-roundworm:
  -d    守护进程模式
  -h    显示帮助
  -i    进程扫描间隔 (ms) (default 100)
  -p    服务器端口 (default 8023)
  -s    服务器地址 (default 127.0.0.1)
  -w    监控目录，多个用 ; 分隔 (default "/tmp")
```

## 架构

```
选手主机:
  └── goawd-server  (Web UI :1337 + 探针接收 :8023)

靶机:
  ├── goawd-roundworm  (文件系统/进程监控)
  ├── goawd-guardian   (PWN 套壳)
  └── TapeWorm         (Web 流量捕获)
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
| FlagBuster | 检测并替换输出中的 flag |
| KingWatcher | KoH 模式赛点文件监控 |
| ZombieKiller | 不死马行为检测 |

## 项目结构

```
AoiAWD/
├── GoAWD/              # Go 后端
│   ├── cmd/            # 入口
│   ├── internal/       # 内部包
│   ├── pkg/            # 公共包
│   └── plugins/        # 插件目录
├── Frontend/           # Vue 前端
├── Guardian/           # PWN 探针 (原版 C)
├── RoundWorm/          # 文件系统探针 (原版 C)
├── TapeWorm/           # Web 探针 (原版 PHP)
└── AoiAWD/             # 核心服务器 (原版 PHP)
```

## 许可证

GNU AGPL-3.0