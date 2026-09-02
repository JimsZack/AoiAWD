@echo off
chcp 65001 >nul
echo 开始搜索当前目录下的flag...
echo.

REM 自动检测可执行文件
if exist snowfind.exe (
    snowfind.exe .
) else if exist snowfind-windows-*.exe (
    for %%f in (snowfind-windows-*.exe) do (
        %%f .
        goto :done
    )
) else (
    echo 未找到可执行文件！
    pause
    exit /b 1
)

:done
echo.
echo 搜索完成！结果已保存到 snowfind-result.txt
echo 详细日志已保存到 snowfind-output.txt
pause
