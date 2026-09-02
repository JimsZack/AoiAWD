@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

set APP_NAME=snowfind
set VERSION=%1
if "%VERSION%"=="" set VERSION=v1.0.0

echo 构建 %APP_NAME% %VERSION% (全平台版本)

REM 检查Go环境
go version >nul 2>&1
if errorlevel 1 (
    echo 错误：未找到Go语言环境，请先安装Go语言
    exit /b 1
)

REM 创建目录
echo 创建目录结构...
if exist bin rmdir /s /q bin
if exist release rmdir /s /q release
mkdir bin
mkdir release

REM 下载依赖
echo 下载依赖...
go mod tidy

REM 运行测试
echo 运行测试...
go test -v .\... || echo 警告：测试失败，继续构建...

REM 平台列表
set PLATFORMS=linux-amd64:linux/amd64 linux-arm64:linux/arm64 darwin-amd64:darwin/amd64 darwin-arm64:darwin/arm64 windows-amd64:windows/amd64 windows-386:windows/386 freebsd-amd64:freebsd/amd64

echo 开始构建所有平台...

for %%p in (%PLATFORMS%) do (
    for /f "tokens=1,2 delims=:" %%a in ("%%p") do (
        set PLATFORM_NAME=%%a
        set PLATFORM=%%b
        
        for /f "tokens=1,2 delims=/" %%c in ("!PLATFORM!") do (
            set GOOS=%%c
            set GOARCH=%%d
            
            set OUTPUT=bin\!APP_NAME!-!PLATFORM_NAME!
            if "!GOOS!"=="windows" set OUTPUT=!OUTPUT!.exe
            
            echo 构建 !PLATFORM_NAME! (!PLATFORM!)...
            set CGO_ENABLED=0
            set GOOS=!GOOS!
            set GOARCH=!GOARCH!
            go build -ldflags="-s -w -X main.version=%VERSION%" -o "!OUTPUT!" main.go
            
            if !errorlevel! equ 0 (
                echo ✓ 成功构建: !OUTPUT!
                
                REM 创建发布包
                set PLATFORM_DIR=release\!APP_NAME!-!PLATFORM_NAME!-%VERSION%
                mkdir "!PLATFORM_DIR!"
                
                REM 复制文件
                copy "!OUTPUT!" "!PLATFORM_DIR!\" >nul
                
                if exist config-complete.json (
                    copy config-complete.json "!PLATFORM_DIR!\config.json" >nul
                ) else (
                    copy config.json "!PLATFORM_DIR!\" >nul
                )
                
                copy release-template\README.md "!PLATFORM_DIR!\" >nul
                
                REM 创建运行脚本
                if "!GOOS!"=="windows" (
                    copy release-template\run.bat "!PLATFORM_DIR!\" >nul
                ) else (
                    copy release-template\run.sh "!PLATFORM_DIR!\" >nul
                )
                
                REM 创建压缩包（Windows使用zip，其他平台使用tar.gz）
                cd release
                if "!GOOS!"=="windows" (
                    if exist "%ProgramFiles%\7-Zip\7z.exe" (
                        "%ProgramFiles%\7-Zip\7z.exe" a -tzip "!APP_NAME!-!PLATFORM_NAME!-%VERSION%.zip" "!APP_NAME!-!PLATFORM_NAME!-%VERSION%\" >nul
                    ) else (
                        powershell -command "Compress-Archive -Path '!APP_NAME!-!PLATFORM_NAME!-%VERSION%' -DestinationPath '!APP_NAME!-!PLATFORM_NAME!-%VERSION%.zip'" >nul
                    )
                ) else (
                    if exist "%ProgramFiles%\7-Zip\7z.exe" (
                        "%ProgramFiles%\7-Zip\7z.exe" a -tgzip "!APP_NAME!-!PLATFORM_NAME!-%VERSION%.tar.gz" "!APP_NAME!-!PLATFORM_NAME!-%VERSION%\" >nul
                    ) else (
                        echo 警告：未找到7-Zip，Linux平台建议使用tar.gz格式
                        powershell -command "Compress-Archive -Path '!APP_NAME!-!PLATFORM_NAME!-%VERSION%' -DestinationPath '!APP_NAME!-!PLATFORM_NAME!-%VERSION%.zip'" >nul
                    )
                )
                cd ..
                
                echo ✓ 打包完成: !APP_NAME!-!PLATFORM_NAME!-%VERSION%
            ) else (
                echo ✗ 构建失败: !PLATFORM_NAME!
                exit /b 1
            )
        )
    )
)

REM 构建本地版本
echo 构建本地版本...
set CGO_ENABLED=0
set GOOS=
set GOARCH=
go build -ldflags="-s -w -X main.version=%VERSION%" -o snowfind.exe main.go

echo.
echo 构建完成！
echo.
echo 生成的二进制文件：
dir bin\
echo.
echo 生成的发布包：
dir release\
echo.
echo 本地测试版本: snowfind.exe
echo.
echo 使用方法：
echo   snowfind.exe --help                  # 查看帮助
echo   snowfind.exe --show-encoders         # 查看支持的编码器
echo   snowfind.exe /path/to/file           # 搜索文件
echo   snowfind.exe /path/to/directory      # 搜索目录
echo.
echo 发布说明：
echo - 每个平台的压缩包包含可执行文件、配置文件、运行脚本和说明文档
echo - Windows 版本为 .zip 格式，其他平台为 .tar.gz 格式
echo - 解压后直接运行 run.bat (Windows) 或 run.sh (Linux/Mac) 即可开始搜索
