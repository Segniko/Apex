@echo off
setlocal
SET "PROJECT_ROOT=%~dp0"

echo [1/3] Launching Infrastructure (Postgres, Prometheus, Grafana)...
docker-compose up -d

echo [2/3] Starting Go Production Receiver...
if "%GEMINI_API_KEY%"=="" (
    echo [ERROR] GEMINI_API_KEY is not set in this session.
    echo Please set it first: set GEMINI_API_KEY=your_key
    start "Apex Receiver (Port 8081)" cmd /k "cd /d %PROJECT_ROOT% && go run cmd/server/main.go"
) else (
    echo [INFO] GEMINI_API_KEY detected, passing to receiver...
    start "Apex Receiver (Port 8081)" cmd /k "cd /d %PROJECT_ROOT% && set GEMINI_API_KEY=%GEMINI_API_KEY% && go run cmd/server/main.go"
)

echo [3/3] Starting Next.js Dashboard (Port 3000)...
start "Apex Dashboard" cmd /k "cd /d %PROJECT_ROOT%\dashboard && npm run dev"

echo.
echo ========================================================
echo   🚀 APEX MONITORING ENGINE IS ONLINE
echo ========================================================
echo   - Dashboard:   http://localhost:3000
echo   - Metrics:     http://localhost:8081/metrics
echo   - Prometheus:  http://localhost:9090
echo   - Grafana:     http://localhost:3001 (admin/admin)
echo ========================================================
echo.
pause
