@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo GPanel Frontend Build Script
echo ========================================
echo.

set BASE_PATH=%~dp0
set WEB_PATH=%BASE_PATH%frontend
set CORE_PATH=%BASE_PATH%core
set WEB_DIST_PATH=%CORE_PATH%\routes\web\dist

where node >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Node.js not found
    pause
    exit /b 1
)

echo [1/3] Cleaning frontend artifacts...
if exist "%WEB_DIST_PATH%" rmdir /s /q "%WEB_DIST_PATH%"
if exist "%WEB_PATH%\dist" rmdir /s /q "%WEB_PATH%\dist"
if exist "%WEB_PATH%\.vite" rmdir /s /q "%WEB_PATH%\.vite"
echo Clean completed

echo.
echo [2/3] Installing frontend dependencies...
cd "%WEB_PATH%"
call npm install --legacy-peer-deps
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Frontend dependencies installation failed
    pause
    exit /b 1
)
echo Dependencies installed

echo.
echo [3/3] Building frontend...
call npm run build:pro
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Frontend build failed
    pause
    exit /b 1
)
echo Frontend build completed

echo.
echo [4/4] Copying frontend build artifacts...
if not exist "%WEB_DIST_PATH%" mkdir "%WEB_DIST_PATH%"
xcopy /E /Y /I "%WEB_PATH%\dist" "%WEB_DIST_PATH%"
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to copy frontend artifacts
    pause
    exit /b 1
)
echo Frontend artifacts copied to %WEB_DIST_PATH%

echo Verifying copied files...
if not exist "%WEB_DIST_PATH%\index.html" (
    echo [ERROR] index.html not found in %WEB_DIST_PATH%
    pause
    exit /b 1
)
echo Verification completed

echo.
echo ========================================
echo Frontend Build Completed!
echo ========================================
echo.
echo Frontend artifacts copied to: %WEB_DIST_PATH%
echo.
pause