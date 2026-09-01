# GoAWD API 接口文档

## 1. 概述

GoAWD 提供 RESTful API，用于与前端和探针进行通信。

**Base URL**: `http://<host>:1337/api`
**认证方式**: Token (通过 Query 参数或 Header)
**响应格式**: JSON

## 2. 通用说明

### 2.1 认证

所有 API 请求需要在 URL 或 Header 中携带 Token：

```
# Query 参数方式
GET /api/v1/info?token=<access_token>

# Header 方式
GET /api/v1/info
Token: <access_token>
```

### 2.2 通用响应格式

```json
{
    "result": "1",
    "message": "ok",
    "data": {}
}
```

### 2.3 分页响应格式

```json
{
    "page": 1,
    "last_page": 10,
    "data": []
}
```

## 3. 探针通信协议 (TCP:8023)

### 3.1 数据格式

探针通过 TCP 发送 JSON 数据，每行一条消息：

```json
{"type": "<message_type>", "data": {}}
```

### 3.2 消息类型

#### 3.2.1 Web 流量上报

```json
{
    "type": "web",
    "data": {
        "script": "index.php",
        "method": "POST",
        "uri": "/api/login?user=admin",
        "remote": "192.168.1.100:12345",
        "header": {
            "Host": "target.com",
            "Content-Type": "application/x-www-form-urlencoded"
        },
        "get": {
            "user": "admin"
        },
        "post": {
            "password": "123456"
        },
        "cookie": {
            "PHPSESSID": "abc123"
        },
        "file": [],
        "buffer": "响应内容"
    }
}
```

**响应**: 返回处理后的 buffer（Base64 编码）

```
base64(buffer)\n
```

#### 3.2.2 PWN 初始化

```json
{
    "type": "pwn",
    "data": {
        "file": "challenge",
        "type": "stdout",
        "pid": 1234,
        "maps": "base64(maps内容)"
    }
}
```

**注意**: PWN 使用流模式，后续数据直接发送，不再包装 JSON。

#### 3.2.3 文件系统事件

```json
{
    "type": "file",
    "data": {
        "path": "/var/www/html/shell.php",
        "event": 256,
        "size": 1024,
        "content": "base64(文件内容)"
    }
}
```

**事件类型 (event)**:
| 值 | 含义 |
|----|------|
| 1 | ACCESS |
| 2 | MODIFY |
| 4 | ATTRIB |
| 8 | CLOSE_WRITE |
| 16 | CLOSE_NOWRITE |
| 32 | OPEN |
| 64 | MOVED_FROM |
| 128 | MOVED_TO |
| 256 | CREATE |
| 512 | DELETE |
| 1024 | DELETE_SELF |
| 2048 | MOVE_SELF |

#### 3.2.4 新进程

```json
{
    "type": "new_process",
    "data": {
        "pid": 1234,
        "ppid": 1,
        "uid": 0,
        "username": "root",
        "cmd": "/bin/bash",
        "param": "-c whoami"
    }
}
```

#### 3.2.5 进程列表

```json
{
    "type": "pid_list",
    "data": [1, 2, 3, 1234]
}
```

#### 3.2.6 心跳

```json
{
    "type": "ping",
    "data": []
}
```

**响应**:

```json
{"type": "pong", "data": []}
```

## 4. REST API (HTTP:1337)

### 4.1 系统信息

#### GET /api/v1/info

获取系统运行状态信息。

**请求参数**: 无

**响应**:
```json
{
    "timestamp_lastupdate": "2024-01-01 12:00:00",
    "count_alert": 10,
    "timestamp_runningtime": "01:30:00"
}
```

#### GET /api/ping

健康检查。

**响应**: `pong`

---

### 4.2 Web 日志

#### GET /api/v1/listweb

获取 Web 日志列表（分页）。

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| count | int | 否 | 每页数量，默认 20 |

**响应**:
```json
{
    "page": 1,
    "last_page": 10,
    "data": [
        {
            "id": "507f1f77bcf86cd799439011",
            "time": "2024-01-01 12:00:00",
            "method": "POST",
            "uri": "/api/login",
            "remote": "192.168.1.100:12345"
        }
    ]
}
```

#### GET /api/v1/webdetail

获取 Web 日志详情。

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 日志 ID |

**响应**:
```json
{
    "id": "507f1f77bcf86cd799439011",
    "time": 1704067200,
    "script": "index.php",
    "method": "POST",
    "uri": "/api/login",
    "remote": "192.168.1.100:12345",
    "header": {},
    "get": {},
    "post": {},
    "cookie": {},
    "file": [],
    "buffer": "响应内容"
}
```

#### GET /api/v1/downloadwebautoscript

下载 Web 日志对应的自动化脚本。

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 日志 ID |

**响应**: PHP 脚本文件

---

### 4.3 PWN 日志

#### GET /api/v1/listpwn

获取 PWN 日志列表（分页）。

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| count | int | 否 | 每页数量，默认 20 |

**响应**:
```json
{
    "page": 1,
    "last_page": 5,
    "data": [
        {
            "id": "507f1f77bcf86cd799439012",
            "time": "2024-01-01 12:00:00",
            "bin": "challenge",
            "stdin": {"group": 5, "byte": 1024},
            "stdout": {"group": 5, "byte": 2048}
        }
    ]
}
```

#### GET /api/v1/pwndetail

获取 PWN 日志详情。

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 日志 ID |

**响应**:
```json
{
    "id": "507f1f77bcf86cd799439012",
    "time": 1704067200,
    "bin": "challenge",
    "maps": "base64(maps内容)",
    "stdin": {"group": 5, "byte": 1024},
    "stdout": {"group": 5, "byte": 2048},
    "streamlog": [
        {"type": "stdin", "buffer": "base64(data)"},
        {"type": "stdout", "buffer": "base64(data)"}
    ]
}
```

#### GET /api/v1/downloadpwn

下载 PWN 日志数据。

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 日志 ID |
| type | string | 是 | 数据类型: `maps` 或 `stream` |
| part | string | 否 | 流的部分: `all` 或索引 |

**响应**: 二进制数据

#### POST /api/v1/generatepwnbin

生成 Guardian PWN 探针二进制。

**请求参数** (JSON Body):
```json
{
    "binary": "base64(原始ELF)",
    "host": "192.168.1.1",
    "port": 8023
}
```

**响应**: Guardian 二进制文件

---

### 4.4 文件系统日志

#### GET /api/v1/listfilesystem

获取文件系统日志列表（分页）。

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| count | int | 否 | 每页数量，默认 20 |

**响应**:
```json
{
    "page": 1,
    "last_page": 10,
    "data": [
        {
            "id": "507f1f77bcf86cd799439013",
            "time": "2024-01-01 12:00:00",
            "path": "/var/www/html/shell.php",
            "oper": "CREATE",
            "isdir": false,
            "content": "base64(前50字节)"
        }
    ]
}
```

#### GET /api/v1/downloadfile

下载文件内容。

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 日志 ID |

**响应**: 文件二进制内容

---

### 4.5 进程日志

#### GET /api/v1/listprocess

获取进程日志列表（分页）。

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| count | int | 否 | 每页数量，默认 20 |

**响应**:
```json
{
    "page": 1,
    "last_page": 10,
    "data": [
        {
            "id": "507f1f77bcf86cd799439014",
            "time": "2024-01-01 12:00:00",
            "pid": 1234,
            "ppid": 1,
            "uid": 0,
            "user": "root",
            "bin": "/bin/bash",
            "arg": "-c whoami"
        }
    ]
}
```

#### GET /api/v1/listcurrentprocess

获取当前活跃进程列表。

**响应**:
```json
{
    "page": 1,
    "last_page": 1,
    "data": [
        {
            "id": 1234,
            "time": "2024-01-01 12:00:00",
            "pid": 1234,
            "ppid": 1,
            "uid": 0,
            "user": "root",
            "bin": "/bin/bash",
            "arg": ""
        }
    ]
}
```

#### GET /api/v1/currentprocess

获取当前进程 PID 列表。

**响应**:
```json
[1, 2, 3, 1234]
```

---

### 4.6 告警日志

#### GET /api/v1/listalert

获取告警日志列表（分页）。

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| count | int | 否 | 每页数量，默认 20 |

**响应**:
```json
{
    "page": 1,
    "last_page": 5,
    "data": [
        {
            "id": "507f1f77bcf86cd799439015",
            "time": "2024-01-01 12:00:00",
            "type": "Web",
            "plugin": "FlagBuster",
            "message": "发现flag字段，已替换",
            "reference": {
                "page": 1,
                "id": "507f1f77bcf86cd799439011"
            }
        }
    ]
}
```

---

### 4.7 插件管理

#### GET /api/v1/listplugin

获取已加载插件列表。

**响应**:
```json
{
    "result": "1",
    "message": "ok",
    "data": ["FlagBuster.php", "KingWatcher.php", "ZombieKiller.php"]
}
```

#### GET /api/v1/reloadplugin

重新加载插件。

**响应**:
```json
{
    "result": "1",
    "message": "ok"
}
```

---

## 5. WebSocket

### 5.1 连接

```
ws://<host>:1337/websocket
```

### 5.2 消息格式

```json
{
    "operation": "reload",
    "type": "web|pwn|file|process|alert"
}
```

**type 说明**:
| type | 含义 |
|------|------|
| web | Web 日志更新 |
| pwn | PWN 日志更新 |
| file | 文件系统日志更新 |
| process | 进程日志更新 |
| alert | 告警日志更新 |

## 6. 错误码

| HTTP 状态码 | 含义 |
|-------------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 (Token 无效) |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
