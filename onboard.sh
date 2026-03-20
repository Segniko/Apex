#!/bin/bash
# Apex Unified Onboarding for Unix/Mac

# Check if Python is installed
if ! command -v python3 &> /dev/null
then
    echo "❌ Python 3 is required for onboarding. Please install it first."
    exit 1
fi

# 1. Check for local script first (in current dir or scripts/)
if [ -f "apex-onboard.py" ]; then
    python3 apex-onboard.py
elif [ -f "scripts/apex-onboard.py" ]; then
    python3 scripts/apex-onboard.py
else
    echo "📡 Onboarding engine not found locally. Fetching..."
    curl -sSL https://raw.githubusercontent.com/Segniko/Apex/main/scripts/apex-onboard.py -o apex-onboard.py
    if [ $? -eq 0 ]; then
        python3 apex-onboard.py
    else
        echo "❌ Failed to fetch onboarding engine. Check your connection or repo visibility."
        exit 1
    fi
fi
