# Snowfind - 多编码格式搜索工具

一个强大的多编码格式文本搜索工具，专为CTF、安全分析和文件搜索而设计。

## 特性

- 🔍 支持多种编码格式搜索
- 🚀 高性能并发搜索
- 📁 递归目录扫描
- 🎯 智能模式匹配
- 📝 详细结果报告
- ⚙️ 灵活配置选项

## 快速开始

### 解压文件
- **Linux/Mac/FreeBSD**: `tar -xzf snowfind-*.tar.gz`
- **Windows**: 使用WinRAR、7-Zip或Windows资源管理器解压.zip文件

### 一键运行
- **Windows**: 双击 `run.bat`
- **Linux/Mac**: 执行 `chmod +x run.sh && ./run.sh`

### 命令行使用

```bash
# 搜索当前目录
./snowfind .

# 搜索指定目录
./snowfind /path/to/directory

# 使用配置文件
./snowfind --config config.json /path/to/search

# 查看帮助
./snowfind --help

# 显示支持的编码器
./snowfind --show-encoders
```

## 支持的编码格式

- **ASCII**: 标准ASCII编码
- **Base64**: Base64编码
- **Hex**: 十六进制编码
- **Binary**: 二进制编码
- **Unicode**: Unicode编码
- **URL**: URL编码
- **HTML**: HTML实体编码

## 配置文件说明

`config.json` 包含以下主要配置：

```json
{
    "patterns": ["flag", "key", "password", "secret"],
    "match": ["flag{", "ctf{", "key:", "pass:"],
    "max_workers": 8,
    "max_processes": 2,
    "log_level": "INFO",
    "output_config": {
        "save_results": true,
        "result_file": "snowfind-result.txt",
        "output_file": "snowfind-output.txt"
    }
}
```

### 主要配置项

- `patterns`: 搜索模式列表
- `match`: 精确匹配模式
- `max_workers`: 最大工作线程数
- `exclude_files`: 排除的文件列表
- `exclude_exts`: 排除的文件扩展名
- `suspicious_exts`: 可疑文件扩展名

## 输出文件

搜索完成后会生成两个文件：

1. **snowfind-result.txt**: 搜索结果摘要
2. **snowfind-output.txt**: 详细搜索日志

## 使用示例

### CTF Flag 搜索
```bash
# 搜索常见的flag格式
./snowfind --pattern "flag{" --pattern "ctf{" ./

# 搜索Base64编码的flag
./snowfind --encoder base64 --pattern "flag" ./
```

### 安全分析
```bash
# 搜索敏感信息
./snowfind --pattern "password" --pattern "secret" --pattern "api_key" ./

# 搜索配置文件
./snowfind --pattern "config" --pattern "database" ./config/
```

### 自定义搜索
```bash
# 使用正则表达式
./snowfind --regex --pattern "flag\{[a-f0-9]+\}" ./

# 限制文件大小
./snowfind --max-size 10MB --pattern "flag" ./
```

## 高级功能

### 时间过滤
```bash
# 只搜索最近7天修改的文件
./snowfind --time-filter --days 7 --pattern "flag" ./
```

### 可疑文件扫描
```bash
# 启用可疑文件扫描
./snowfind --suspicious-scan --pattern "flag" ./
```

### 交互模式
```bash
# 启动交互模式
./snowfind --interactive
```

## 性能调优

- 增加 `max_workers` 提高搜索速度
- 使用 `exclude_files` 和 `exclude_exts` 排除不必要的文件
- 调整 `buffer_size` 优化内存使用
- 设置 `max_file_size` 限制大文件处理

## 故障排除

### 权限问题
```bash
# 给运行脚本添加执行权限
chmod +x run.sh
chmod +x snowfind-*
```

### 编码问题
- 确保终端支持UTF-8编码
- Windows用户可能需要使用管理员权限运行

### 性能问题
- 减少并发数 (`max_workers`)
- 增加文件排除规则
- 使用更具体的搜索模式

## 更新日志

查看 [GitHub Releases](https://github.com/your-repo/snowfind/releases) 获取最新版本和更新日志。

## 技术支持

如遇问题请提交 [GitHub Issue](https://github.com/your-repo/snowfind/issues)。

---

**注意**: 本工具仅供合法的安全研究和教育用途使用，请遵守当地法律法规。
