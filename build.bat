@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo GPanel Build Script
echo ========================================
echo.

set VERSION=dev
for /f "delims=" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
if "%GIT_COMMIT%"=="" set GIT_COMMIT=unknown

set BASE_PATH=%~dp0
set BUILD_PATH=%BASE_PATH%build
set WEB_PATH=%BASE_PATH%frontend
set CORE_PATH=%BASE_PATH%core
set WEB_DIST_PATH=%CORE_PATH%\web\dist
set CORE_MAIN=%CORE_PATH%\main.go
set CORE_NAME=gpanel

where node >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Node.js not found
    pause
    exit /b 1
)

where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go not found
    pause
    exit /b 1
)

echo [1/5] Cleaning build artifacts...
if exist "%BUILD_PATH%" rmdir /s /q "%BUILD_PATH%"
if exist "%WEB_DIST_PATH%" rmdir /s /q "%WEB_DIST_PATH%"
if exist "%WEB_PATH%\dist" rmdir /s /q "%WEB_PATH%\dist"
echo Clean completed

echo.
echo [2/5] Installing frontend dependencies...
cd "%WEB_PATH%"
call npm install --legacy-peer-deps
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Frontend dependencies installation failed
    pause
    exit /b 1
)

echo.
echo [3/5] Building frontend...
call npm run build:pro
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Frontend build failed
    pause
    exit /b 1
)

echo.
echo [4/5] Copying frontend build artifacts...
if not exist "%WEB_DIST_PATH%" mkdir "%WEB_DIST_PATH%"
xcopy /E /Y /I "%WEB_PATH%\dist" "%WEB_DIST_PATH%"
echo Frontend artifacts copied

echo.
echo [5/5] Building backend binaries...

if not exist "%BUILD_PATH%" mkdir "%BUILD_PATH%"

echo Building Windows amd64 version...
cd "%CORE_PATH%"
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags "-s -w -X main.Version=%VERSION% -X main.GitCommit=%GIT_COMMIT%" -o "%BUILD_PATH%\%CORE_NAME%-windows-amd64.exe" main.go frontend.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Windows amd64 build failed
    pause
    exit /b 1
)
echo Windows amd64 build completed

echo Building Linux amd64 version...
cd "%CORE_PATH%"
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags "-s -w -X main.Version=%VERSION% -X main.GitCommit=%GIT_COMMIT%" -o "%BUILD_PATH%\%CORE_NAME%-linux-amd64" main.go frontend.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Linux amd64 build failed
    pause
    exit /b 1
)
echo Linux amd64 build completed

echo.
echo ========================================
echo Build Completed!
echo ========================================
echo.
echo Build artifacts in: %BUILD_PATH%\
echo   - %CORE_NAME%-windows-amd64.exe  ^(Windows version^)
echo   - %CORE_NAME%-linux-amd64       ^(Linux version^)
echo.
echo Version info:
echo   - Version: %VERSION%
echo   - Commit: %GIT_COMMIT%
echo.
pause