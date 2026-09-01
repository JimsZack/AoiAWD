#!/bin/bash
# GoAWD Build Script

set -e

echo "=== Building GoAWD Backend ==="
cd GoAWD
go build -v -o bin/goawd-server ./cmd/server
CGO_ENABLED=0 go build -v -o bin/goawd-roundworm ./cmd/roundworm
CGO_ENABLED=0 go build -v -o bin/goawd-guardian ./cmd/guardian
cd ..

echo "=== Building Frontend ==="
cd Frontend
npm install --ignore-scripts --legacy-peer-deps
npm run build
cd ..

echo "=== Creating Package ==="
VERSION=$(date +%Y%m%d_%H%M%S)
PACKAGE_NAME="GoAWD-v${VERSION}"

mkdir -p ${PACKAGE_NAME}
cp -r GoAWD/bin ${PACKAGE_NAME}/
cp -r Frontend/dist ${PACKAGE_NAME}/frontend
cp docker-compose.yml ${PACKAGE_NAME}/
cp Dockerfile ${PACKAGE_NAME}/
cp docker_AoiAWD_Start.sh ${PACKAGE_NAME}/
cp README.md ${PACKAGE_NAME}/

tar -czf ${PACKAGE_NAME}.tar.gz ${PACKAGE_NAME}
rm -rf ${PACKAGE_NAME}

echo "=== Build Complete ==="
echo "Package: ${PACKAGE_NAME}.tar.gz"
ls -lh ${PACKAGE_NAME}.tar.gz