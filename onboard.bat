@echo off
:: Apex Unified Onboarding for Windows

where python >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Python is not detected in your PATH.
    echo Please install Python from python.org to continue.
    pause
    exit /b 1
)

if not exist apex-onboard.py (
    echo [INFO] Fetching onboarding engine...
    curl -sSL https://raw.githubusercontent.com/Segniko/Apex/main/scripts/apex-onboard.py -o apex-onboard.py
)

python apex-onboard.py
pause
