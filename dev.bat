@echo off
REM GPanel 开发环境启动脚本

echo ========================================
echo GPanel 开发环境启动
echo ========================================
echo.

REM 检查 Node.js
where node >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 未找到 Node.js，请先安装 Node.js
    pause
    exit /b 1
)

REM 检查 Go
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 未找到 Go，请先安装 Go
    pause
    exit /b 1
)

echo [1/3] 安装前端依赖...
cd frontend
call npm install
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 前端依赖安装失败
    pause
    exit /b 1
)

echo [2/3] 构建前端...
call npm run build
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 前端构建失败
    pause
    exit /b 1
)

echo [3/3] 复制构建产物...
if not exist "..\core\web\dist" mkdir "..\core\web\dist"
xcopy /E /Y /I dist ..\core\web\dist

echo.
echo ========================================
echo 前端构建完成！
echo ========================================
echo.
echo 现在可以启动后端服务：
echo   cd ..\core
echo   go run main.go frontend.go
echo.
echo 或使用 Makefile 构建：
echo   make build
echo.
pause