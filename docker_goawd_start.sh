#!/bin/bash
set -e

IMAGE_NAME="goawd"
CONTAINER_NAME="goawd"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

printf "${GREEN}[+] GoAWD Docker 启动脚本${NC}\n"

# 检查镜像是否存在
if docker image ls | grep -q "$IMAGE_NAME"; then
    printf "${YELLOW}[!] 镜像已存在，是否重新构建？(y/N) ${NC}"
    read -r choice
    if [ "$choice" = 'y' ] || [ "$choice" = 'Y' ]; then
        printf "${GREEN}[+] 重新构建镜像...${NC}\n"
        docker compose down 2>/dev/null || true
        docker compose build
        printf "${GREEN}[+] 构建完成${NC}\n"
    fi
else
    printf "${GREEN}[+] 首次构建镜像...${NC}\n"
    docker compose build
    printf "${GREEN}[+] 构建完成${NC}\n"
fi

# 启动容器
printf "${GREEN}[+] 启动容器...${NC}\n"
docker compose up -d

# 等待容器启动
sleep 3

# 检查容器状态
if docker ps | grep -q "$CONTAINER_NAME"; then
    printf "${GREEN}[+] 容器启动成功${NC}\n"
    
    # 获取访问令牌
    TOKEN=$(docker logs "$CONTAINER_NAME" 2>&1 | grep -i "token\|access" | tail -1 | awk '{print $NF}')
    if [ -n "$TOKEN" ]; then
        printf "${GREEN}[+] 访问令牌: ${TOKEN}${NC}\n"
    fi
    
    printf "${GREEN}[+] 服务地址: http://localhost:1337${NC}\n"
    printf "${GREEN}[+] TCP 端口: 8023${NC}\n"
    printf "${GREEN}[+] 全部完成!${NC}\n"
else
    printf "${RED}[-] 容器启动失败${NC}\n"
    docker compose logs
    exit 1
fi
