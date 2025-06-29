#!/usr/bin/env python3
"""
HTMLnoJS Demo - Simple demonstration
"""
import asyncio
import sys
import signal
import time
from pathlib import Path

# Add the package to Python path for development
sys.path.insert(0, str(Path(__file__).parent))

from htmlnojs.core import htmlnojs


async def main():
    app = htmlnojs("./go-server/debug", verbose=True)
    await app.start()
    # Keep running until interrupted
    while True: time.sleep(1)

if __name__ == "__main__":
    asyncio.run(main())