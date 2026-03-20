@echo off
:: Apex Unified Onboarding for Windows

where python >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Python is not detected in your PATH.
    echo Please install Python from python.org to continue.
    pause
    exit /b 1
)

:: 1. Check for local script first (in current dir or scripts/)
set "CORE_SCRIPT=apex-onboard.py"
set "LOCAL_CORE=scripts\apex-onboard.py"

if exist %CORE_SCRIPT% (
    python %CORE_SCRIPT%
) else if exist %LOCAL_CORE% (
    python %LOCAL_CORE%
) else (
    echo [INFO] Onboarding engine not found locally. Fetching...
    curl -sSL https://raw.githubusercontent.com/Segniko/Apex/main/scripts/apex-onboard.py -o apex-onboard.py
    if %ERRORLEVEL% neq 0 (
        echo [ERROR] Failed to fetch onboarding engine. Check your connection or repo visibility.
    ) else (
        python %CORE_SCRIPT%
    )
)

pause
