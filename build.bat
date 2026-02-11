@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo GPanel Build Script
echo ========================================
echo.

@REM set VERSION=dev
set VERSION=release
set BASE_PATH=%~dp0
set BUILD_PATH=%BASE_PATH%build
set WEB_PATH=%BASE_PATH%frontend
set CORE_PATH=%BASE_PATH%core
set AGENT_PATH=%BASE_PATH%agent
set WEB_DIST_PATH=%CORE_PATH%\web\dist
set CORE_MAIN=%CORE_PATH%\main.go
set AGENT_MAIN=%AGENT_PATH%\main.go
set CORE_NAME=gpanel
set AGENT_NAME=gpanel-agent

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

echo [1/6] Cleaning build artifacts...
if exist "%BUILD_PATH%" rmdir /s /q "%BUILD_PATH%"
if exist "%WEB_DIST_PATH%" rmdir /s /q "%WEB_DIST_PATH%"
if exist "%WEB_PATH%\dist" rmdir /s /q "%WEB_PATH%\dist"
if exist "%CORE_PATH%\routes\web\dist" rmdir /s /q "%CORE_PATH%\routes\web\dist"
if exist "%WEB_PATH%\.vite" rmdir /s /q "%WEB_PATH%\.vite"
echo Clean completed

echo.
echo [2/6] Installing frontend dependencies...
cd "%WEB_PATH%"
call npm install --legacy-peer-deps
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Frontend dependencies installation failed
    pause
    exit /b 1
)

echo.
echo [3/6] Building frontend...
call npm run build:pro
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Frontend build failed
    pause
    exit /b 1
)

echo.
echo [4/6] Copying frontend build artifacts...
set ROUTES_WEB_DIST=%CORE_PATH%\routes\web\dist
if not exist "%ROUTES_WEB_DIST%" mkdir "%ROUTES_WEB_DIST%"
xcopy /E /Y /I "%WEB_PATH%\dist" "%ROUTES_WEB_DIST%"
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to copy frontend artifacts
    pause
    exit /b 1
)
echo Frontend artifacts copied to %ROUTES_WEB_DIST%
echo Verifying copied files...
if not exist "%ROUTES_WEB_DIST%\index.html" (
    echo [ERROR] index.html not found in %ROUTES_WEB_DIST%
    pause
    exit /b 1
)
echo Verification completed

echo.
echo [5/6] Building Core backend binaries...

if not exist "%BUILD_PATH%" mkdir "%BUILD_PATH%"

echo Building Windows amd64 Core version...
cd "%CORE_PATH%"
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags "-s -w" -o "%BUILD_PATH%\windows-amd64\%CORE_NAME%.exe" main.go frontend.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Windows amd64 Core build failed
    pause
    exit /b 1
)
echo Windows amd64 Core build completed

echo Building Windows gpctl...
cd "%BASE_PATH%\tools"
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags "-s -w" -o "%BUILD_PATH%\windows-amd64\gpctl.exe" gpctl.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Windows gpctl build failed
    pause
    exit /b 1
)
echo Windows gpctl build completed

echo Building Linux amd64 Core version...
cd "%CORE_PATH%"
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags "-s -w" -o "%BUILD_PATH%\linux-amd64\%CORE_NAME%" main.go frontend.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Linux amd64 Core build failed
    pause
    exit /b 1
)
echo Linux amd64 Core build completed

echo Building Linux gpctl...
cd "%BASE_PATH%\tools"
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags "-s -w" -o "%BUILD_PATH%\linux-amd64\gpctl" gpctl.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Linux gpctl build failed
    pause
    exit /b 1
)
echo Linux gpctl build completed

echo.
echo [6/6] Building Agent binaries...

echo Updating Agent dependencies...
cd "%AGENT_PATH%"
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Agent dependencies update failed
    pause
    exit /b 1
)

echo Building Windows amd64 Agent version...
cd "%AGENT_PATH%"
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags "-s -w" -o "%BUILD_PATH%\windows-amd64\%AGENT_NAME%.exe" main.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Windows amd64 Agent build failed
    pause
    exit /b 1
)
echo Windows amd64 Agent build completed

echo Building Linux amd64 Agent version...
cd "%AGENT_PATH%"
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags "-s -w" -o "%BUILD_PATH%\linux-amd64\%AGENT_NAME%" main.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Linux amd64 Agent build failed
    pause
    exit /b 1
)
echo Linux amd64 Agent build completed

echo.
echo ========================================
echo Build Completed!
echo ========================================
echo.
echo Build artifacts in: %BUILD_PATH%\
echo   - %CORE_NAME%.exe       ^(Windows Core^)
echo   - %AGENT_NAME%.exe      ^(Windows Agent^)
echo   - gpctl.exe             ^(Windows gpctl^)
echo   - %CORE_NAME%           ^(Linux Core^)
echo   - %AGENT_NAME%          ^(Linux Agent^)
echo   - gpctl                 ^(Linux gpctl^)
echo.
echo Usage:
echo   1. Start Agent: %AGENT_NAME%
echo   2. Start Core:  %CORE_NAME%
echo.
pause