#!/bin/bash

# snowfind 全平台构建脚本

set -e

APP_NAME="snowfind"
VERSION=${1:-"v1.0.0"}

echo "构建 $APP_NAME $VERSION (全平台版本)"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误：未找到Go语言环境，请先安装Go语言"
    exit 1
fi

# 创建必要的目录
echo "创建目录结构..."
rm -rf bin release
mkdir -p bin release

# 初始化Go模块（如果还没有）
if [ ! -f go.mod ]; then
    echo "初始化Go模块..."
    go mod init snowfind
fi

# 下载依赖
echo "下载依赖..."
go mod tidy

# 运行测试
echo "运行测试..."
go test -v ./... || echo "警告：测试失败，继续构建..."

# 支持的平台
declare -A PLATFORMS=(
    ["linux-amd64"]="linux/amd64"
    ["linux-arm64"]="linux/arm64"
    ["darwin-amd64"]="darwin/amd64"
    ["darwin-arm64"]="darwin/arm64"
    ["windows-amd64"]="windows/amd64"
    ["windows-386"]="windows/386"
    ["freebsd-amd64"]="freebsd/amd64"
)

echo "开始构建所有平台..."

# 构建所有平台
for platform_name in "${!PLATFORMS[@]}"; do
    platform="${PLATFORMS[$platform_name]}"
    GOOS=${platform%/*}
    GOARCH=${platform#*/}
    
    output="bin/${APP_NAME}-${platform_name}"
    if [ "$GOOS" = "windows" ]; then
        output+=".exe"
    fi
    
    echo "构建 $platform_name ($platform)..."
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w -X main.version=$VERSION" -o "$output" main.go
    
    if [ $? -eq 0 ]; then
        echo "✓ 成功构建: $output"
        
        # 创建发布包
        platform_dir="release/${APP_NAME}-${platform_name}-${VERSION}"
        mkdir -p "$platform_dir"
        
        # 复制二进制文件
        cp "$output" "$platform_dir/"
        
        # 复制配置文件
        if [ -f config-complete.json ]; then
            cp config-complete.json "$platform_dir/config.json"
        else
            cp config.json "$platform_dir/"
        fi
        
        # 复制模板文件
        cp release-template/README.md "$platform_dir/"
        
        # 创建运行脚本
        if [ "$GOOS" = "windows" ]; then
            cp release-template/run.bat "$platform_dir/"
        else
            cp release-template/run.sh "$platform_dir/"
            chmod +x "$platform_dir/run.sh"
        fi
        
        # 创建压缩包（Windows使用zip，其他平台使用tar.gz）
        cd release
        if [ "$GOOS" = "windows" ]; then
            if command -v zip &> /dev/null; then
                zip -r "${APP_NAME}-${platform_name}-${VERSION}.zip" "${APP_NAME}-${platform_name}-${VERSION}/"
            else
                echo "警告：未找到zip命令，使用tar代替"
                tar -czf "${APP_NAME}-${platform_name}-${VERSION}.tar.gz" "${APP_NAME}-${platform_name}-${VERSION}/"
            fi
        else
            tar -czf "${APP_NAME}-${platform_name}-${VERSION}.tar.gz" "${APP_NAME}-${platform_name}-${VERSION}/"
        fi
        cd ..
        
        echo "✓ 打包完成: ${APP_NAME}-${platform_name}-${VERSION}"
    else
        echo "✗ 构建失败: $platform"
        exit 1
    fi
done

# 构建当前平台版本（用于本地测试）
echo "构建本地版本..."
go build -ldflags="-s -w -X main.version=$VERSION" -o snowfind main.go

echo ""
echo "构建完成！"
echo ""
echo "生成的二进制文件："
ls -la bin/
echo ""
echo "生成的发布包："
ls -la release/
echo ""
echo "本地测试版本: ./snowfind"
echo ""
echo "使用方法："
echo "  ./snowfind --help                  # 查看帮助"
echo "  ./snowfind --show-encoders         # 查看支持的编码器"
echo "  ./snowfind /path/to/file           # 搜索文件"
echo "  ./snowfind /path/to/directory      # 搜索目录"
echo ""
echo "发布说明："
echo "- 每个平台的压缩包包含可执行文件、配置文件、运行脚本和说明文档"
echo "- Windows 版本为 .zip 格式，其他平台为 .tar.gz 格式"
echo "- 解压后直接运行 run.bat (Windows) 或 run.sh (Linux/Mac) 即可开始搜索"
