import os
import sys
import json
import shutil
import subprocess
import webbrowser
from pathlib import Path

# --- Constants ---
GITHUB_RAW_BASE = "https://raw.githubusercontent.com/Segniko/Apex/main/"
CONFIG_PATH = Path.home() / ".apex_config.json"
DEFAULT_INGEST_URL = "https://apex-addis.vercel.app/api/ingest" # Cloud default
LOCAL_INGEST_URL = "http://localhost:8081/ingest"

def clear_screen():
    os.system('cls' if os.name == 'nt' else 'clear')

def print_header():
    print("\033[1;33m" + "="*60)
    print("   🚀 APEX_MONITORING: TACTICAL_ONBOARDING_INITIATED")
    print("="*60 + "\033[0m\n")

def get_ingest_key():
    if CONFIG_PATH.exists():
        try:
            with open(CONFIG_PATH, 'r') as f:
                config = json.load(f)
                return config.get("api_key")
        except:
            pass
    
    print("🔑 No API Key detected. Visit \033[1;34mhttps://apex-addis.vercel.app\033[0m to get one.")
    key = input("👉 Enter your INGEST_KEY: ").strip()
    if key:
        with open(CONFIG_PATH, 'w') as f:
            json.dump({"api_key": key}, f)
        print(f"✅ Key saved to {CONFIG_PATH}")
    return key

def detect_language():
    files = os.listdir(".")
    if "go.mod" in files:
        return "go"
    if "package.json" in files:
        return "node"
    if "requirements.txt" in files or any(f.endswith(".py") for f in files):
        return "python"
    return "unknown"

def setup_agent(lang, key):
    print(f"\n🛠  Setting up \033[1;32m{lang.upper()}\033[0m agent...")
    
    agent_files = {
        "python": "agents/python/agent.py",
        "node": "agents/node/agent.js"
    }
    
    output_files = {
        "python": "apex_agent.py",
        "node": "apexAgent.js"
    }

    if lang == "go":
        print("💡 For Go projects, add the agent to your \033[1;36mgo.mod\033[0m:")
        print("\033[1;37m" + "   go get github.com/Segniko/Apex/pkg/agent" + "\033[0m")
        return

    # Try local first, then remote
    remote_url = GITHUB_RAW_BASE + agent_files[lang]
    output_path = output_files[lang]
    
    print(f"📡 Fetching agent DNA from \033[1;34m{remote_url}\033[0m...")
    
    try:
        import urllib.request
        with urllib.request.urlopen(remote_url) as response:
            with open(output_path, 'wb') as out_file:
                out_file.write(response.read())
        print(f"✅ Created \033[1;36m{output_path}\033[0m in your current directory.")
    except Exception as e:
        print(f"❌ Failed to fetch agent: {e}")
        return

    if lang == "python":
        print("\n📦 \033[1;33mINSTALL DEPENDENCIES:\033[0m")
        print("   pip install requests zstandard")
        print("\n💡 \033[1;37mUSAGE:\033[0m")
        print(f"""
from apex_agent import ApexAgent
agent = ApexAgent(ingest_url="https://apex-addis.vercel.app/api/ingest", api_key="{key}")
""")
    elif lang == "node":
        print("\n📦 \033[1;33mINSTALL DEPENDENCIES:\033[0m")
        print("   npm install uuid fzstd")
        print("\n💡 \033[1;37mUSAGE:\033[0m")
        print(f"""
const {{ ApexAgent }} = require('./apexAgent');
const agent = new ApexAgent("https://apex-addis.vercel.app/api/ingest", "{key}");
""")

def launch_hud():
    print("\n🖥  Launching \033[1;33mAPEX_HUD\033[0m...")
    
    # 1. Open Web Dashboard
    print("🌐 Opening Web Dashboard: \033[1;34mhttps://apex-addis.vercel.app/dashboard\033[0m")
    webbrowser.open("https://apex-addis.vercel.app/dashboard")

    # 2. Try to launch TUI
    print("⌨️  Attempting to launch Terminal HUD...")
    
    if shutil.which("go"):
        # Run via go
        print("🚀 Go detected. Running TUI...")
        cmd = ["go", "run", "github.com/Segniko/Apex/cmd/apex-tui@latest"]
        try:
            subprocess.Popen(cmd, creationflags=subprocess.CREATE_NEW_CONSOLE if os.name == "nt" else 0)
        except Exception as e:
            print(f"⚠️  Could not start TUI automatically: {e}")
            print("💡 Manual launch: go run github.com/Segniko/Apex/cmd/apex-tui@latest")
    elif shutil.which("docker"):
        # Run via docker
        print("🐳 Go not detected. Using Docker for TUI...")
        print("💡 Run: docker run -it --rm -v ~/.apex_config.json:/root/.apex_config.json segniko/apex-tui")
    else:
        print("❌ Neither Go nor Docker detected. Please visit the Web Dashboard.")

def main():
    clear_screen()
    print_header()
    
    key = get_ingest_key()
    if not key:
        print("❌ Onboarding aborted: No API Key.")
        return

    lang = detect_language()
    if lang != "unknown":
        setup_agent(lang, key)
    else:
        print("❓ Could not automatically detect project language. Skipping agent setup.")

    input("\n🎯 Setup complete. Press \033[1;33m[ENTER]\033[0m to launch the HUD...")
    launch_hud()

if __name__ == "__main__":
    main()
