#!/bin/bash
# AoiAWD Source Code Packaging Script

set -e

echo "=== AoiAWD Source Code Packaging ==="

VERSION=$(date +%Y%m%d_%H%M%S)
PACKAGE_NAME="AoiAWD-source-v${VERSION}"

echo "Creating source package: ${PACKAGE_NAME}"

# 创建临时目录
mkdir -p ${PACKAGE_NAME}

# 复制源码目录
echo "Copying source code..."
cp -r GoAWD ${PACKAGE_NAME}/
cp -r Frontend ${PACKAGE_NAME}/
cp -r Guardian ${PACKAGE_NAME}/
cp -r RoundWorm ${PACKAGE_NAME}/
cp -r TapeWorm ${PACKAGE_NAME}/
cp -r AoiAWD ${PACKAGE_NAME}/
cp -r Readme ${PACKAGE_NAME}/

# 复制配置文件
echo "Copying config files..."
cp docker-compose.yml ${PACKAGE_NAME}/
cp Dockerfile ${PACKAGE_NAME}/
cp docker_AoiAWD_Start.sh ${PACKAGE_NAME}/
cp Dockerfile ${PACKAGE_NAME}/
cp .dockerignore ${PACKAGE_NAME}/
cp LICENSE ${PACKAGE_NAME}/
cp README.md ${PACKAGE_NAME}/
cp BUILD.md ${PACKAGE_NAME}/
cp build.sh ${PACKAGE_NAME}/
cp opencode.json ${PACKAGE_NAME}/
cp codearts_cli.json ${PACKAGE_NAME}/

# 清理不需要的文件
echo "Cleaning unnecessary files..."
cd ${PACKAGE_NAME}
find . -name "node_modules" -type d -exec rm -rf {} + 2>/dev/null || true
find . -name ".git" -type d -exec rm -rf {} + 2>/dev/null || true
find . -name "*.tar.gz" -exec rm -f {} + 2>/dev/null || true
find . -name "*.zip" -exec rm -f {} + 2>/dev/null || true
find . -name ".DS_Store" -exec rm -f {} + 2>/dev/null || true
cd ..

# 创建压缩包
echo "Creating archive..."
tar -czf ${PACKAGE_NAME}.tar.gz ${PACKAGE_NAME}

# 清理临时目录
rm -rf ${PACKAGE_NAME}

echo "=== Packaging Complete ==="
echo "Package: ${PACKAGE_NAME}.tar.gz"
ls -lh ${PACKAGE_NAME}.tar.gz

# 显示包内容摘要
echo ""
echo "Package contents:"
tar -tzf ${PACKAGE_NAME}.tar.gz | head -50
echo "..."
echo "Total files: $(tar -tzf ${PACKAGE_NAME}.tar.gz | wc -l)"