#!/bin/bash
# Apex Unified Onboarding for Unix/Mac

# Check if Python is installed
if ! command -v python3 &> /dev/null
then
    echo "❌ Python 3 is required for onboarding. Please install it first."
    exit 1
fi

# Download the core script if not present
if [ ! -f "apex-onboard.py" ]; then
    echo "📡 Fetching onboarding engine..."
    curl -sSL https://raw.githubusercontent.com/Segniko/Apex/main/scripts/apex-onboard.py -o apex-onboard.py
fi

# Run the onboarding script
python3 apex-onboard.py
