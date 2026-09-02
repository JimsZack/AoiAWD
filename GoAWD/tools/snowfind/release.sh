#!/bin/bash

# 发布脚本 - 自动构建、打包并上传

set -e

VERSION=${1:-"v1.0.0"}
RELEASE_TYPE=${2:-"patch"}  # major, minor, patch

APP_NAME="snowfind"

echo "开始发布 $APP_NAME $VERSION ($RELEASE_TYPE)"

# 检查git状态
if [ -n "$(git status --porcelain)" ]; then
    echo "警告：工作区有未提交的更改"
    read -p "是否继续？(y/N) " -r
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 更新版本信息
echo "更新版本信息..."
if [ -f package.json ]; then
    npm version $RELEASE_TYPE --no-git-tag-version
fi

# 构建所有平台
echo "构建所有平台..."
./build.sh $VERSION

# 检查构建结果
if [ ! -d "release" ] || [ -z "$(ls -A release/*.tar.gz 2>/dev/null)" ]; then
    echo "错误：构建失败或没有生成发布包"
    exit 1
fi

# 创建git tag
echo "创建git标签..."
git add .
git commit -m "Release $VERSION" || echo "没有更改需要提交"
git tag -a "$VERSION" -m "Release $VERSION"

# 推送到远程仓库
echo "推送到远程仓库..."
git push origin main
git push origin "$VERSION"

echo ""
echo "发布完成！"
echo "版本: $VERSION"
echo "生成的文件："
ls -la release/

echo ""
echo "GitHub Actions 工作流将自动："
echo "1. 构建所有平台的二进制文件"
echo "2. 创建 GitHub Release"
echo "3. 上传所有发布包到 Release"
echo ""
echo "请访问 GitHub 查看发布状态："
echo "https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^/.]*\).*/\1/'))/actions"
