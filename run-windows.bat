@echo off
setlocal

echo ==========================================
echo        Starting WATCHME v2.0
echo ==========================================
echo.

REM Check if Docker is installed
docker --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not installed or not in your PATH.
    echo Please install Docker Desktop for Windows: 
    echo https://docs.docker.com/desktop/install/windows-install/
    echo.
    pause
    exit /b 1
)

REM Check if Docker daemon is running
docker info >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker daemon is not running.
    echo Please start Docker Desktop and wait for it to initialize, then try again.
    echo.
    pause
    exit /b 1
)

echo [INFO] Starting containers...
docker compose up -d

if errorlevel 1 (
    echo.
    echo [ERROR] Failed to start Docker containers.
    echo.
    pause
    exit /b 1
)

echo.
echo ==========================================
echo    WATCHME is now running!
echo ==========================================
echo.
echo    Open your browser and go to:
echo       http://localhost:3000
echo.
echo    To view logs, run:
echo       docker compose logs -f
echo.
echo    To stop the application, run:
echo       docker compose down
echo.
pause
