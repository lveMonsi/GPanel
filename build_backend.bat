@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo GPanel Backend Build Script
echo ========================================
echo.

@REM set VERSION=dev
set VERSION=release
for /f "delims=" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
if "%GIT_COMMIT%"=="" set GIT_COMMIT=unknown

set BASE_PATH=%~dp0
set BUILD_PATH=%BASE_PATH%build
set CORE_PATH=%BASE_PATH%core
set AGENT_PATH=%BASE_PATH%agent
set CORE_NAME=gpanel
set AGENT_NAME=gpanel-agent

where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go not found
    pause
    exit /b 1
)

echo [1/4] Cleaning build artifacts...
if exist "%BUILD_PATH%" rmdir /s /q "%BUILD_PATH%"
echo Clean completed

echo.
echo [2/4] Building Core binaries...

if not exist "%BUILD_PATH%" mkdir "%BUILD_PATH%"
if not exist "%BUILD_PATH%\windows-amd64" mkdir "%BUILD_PATH%\windows-amd64"
if not exist "%BUILD_PATH%\linux-amd64" mkdir "%BUILD_PATH%\linux-amd64"

echo Building Windows amd64 Core version...
cd "%CORE_PATH%"
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags "-s -w -X main.Version=%VERSION% -X main.GitCommit=%GIT_COMMIT%" -o "%BUILD_PATH%\windows-amd64\%CORE_NAME%.exe" main.go frontend.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Windows amd64 Core build failed
    pause
    exit /b 1
)
echo Windows amd64 Core build completed

echo Building Linux amd64 Core version...
cd "%CORE_PATH%"
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags "-s -w -X main.Version=%VERSION% -X main.GitCommit=%GIT_COMMIT%" -o "%BUILD_PATH%\linux-amd64\%CORE_NAME%" main.go frontend.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Linux amd64 Core build failed
    pause
    exit /b 1
)
echo Linux amd64 Core build completed

echo.
echo [3/4] Building Agent binaries...

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
go build -trimpath -ldflags "-s -w -X main.Version=%VERSION% -X main.GitCommit=%GIT_COMMIT%" -o "%BUILD_PATH%\windows-amd64\%AGENT_NAME%.exe" main.go
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
go build -trimpath -ldflags "-s -w -X main.Version=%VERSION% -X main.GitCommit=%GIT_COMMIT%" -o "%BUILD_PATH%\linux-amd64\%AGENT_NAME%" main.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Linux amd64 Agent build failed
    pause
    exit /b 1
)
echo Linux amd64 Agent build completed

echo.
echo [4/4] Building gpctl...

echo Building Windows gpctl...
cd "%BASE_PATH%\tools"
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags "-s -w -X main.Version=%VERSION% -X main.GitCommit=%GIT_COMMIT%" -o "%BUILD_PATH%\windows-amd64\gpctl.exe" gpctl.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Windows gpctl build failed
    pause
    exit /b 1
)
echo Windows gpctl build completed

echo Building Linux gpctl...
cd "%BASE_PATH%\tools"
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags "-s -w -X main.Version=%VERSION% -X main.GitCommit=%GIT_COMMIT%" -o "%BUILD_PATH%\linux-amd64\gpctl" gpctl.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Linux gpctl build failed
    pause
    exit /b 1
)
echo Linux gpctl build completed

echo.
echo ========================================
echo Backend Build Completed!
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
echo Version info:
echo   - Version: %VERSION%
echo   - Commit: %GIT_COMMIT%
echo.
pause