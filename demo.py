#!/usr/bin/env python3
"""
HTMLnoJS Demo - Simple demonstration
"""
import asyncio
import sys
import time
from pathlib import Path

# Add the package to Python path for development
sys.path.insert(0, str(Path(__file__).parent))

from core import htmlnojs


async def main():
    """Simple HTMLnoJS demo"""
    print("🚀 HTMLnoJS Demo")

    # Dead simple usage
    app = htmlnojs("./demo-project")
    await app.start()

    print(f"✅ Started on {app.go_server.url}")

    try:
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        pass
    finally:
        await app.stop()
        print("👋 Stopped")


if __name__ == "__main__":
    asyncio.run(main())