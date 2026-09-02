#!/bin/bash
echo "开始搜索当前目录下的flag..."
echo

# 自动检测可执行文件
if [ -f ./snowfind ]; then
    ./snowfind .
elif [ -f ./snowfind-linux-* ]; then
    ./snowfind-linux-* .
elif [ -f ./snowfind-darwin-* ]; then
    ./snowfind-darwin-* .
elif [ -f ./snowfind-freebsd-* ]; then
    ./snowfind-freebsd-* .
else
    echo "未找到可执行文件！"
    exit 1
fi

echo
echo "搜索完成！结果已保存到 snowfind-result.txt"
echo "详细日志已保存到 snowfind-output.txt"
read -p "按回车键退出..."
