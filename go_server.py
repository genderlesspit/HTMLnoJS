"""
Go Server Manager - Handles Go server lifecycle
"""
import asyncio
from pathlib import Path
from typing import Optional


class GoServer:
    """Manages Go server process"""

    def __init__(self, project_dir: str, port: int, python_url: str):
        self.project_dir = Path(project_dir).resolve()
        self.port = port
        self.python_url = python_url
        self._process: Optional[asyncio.subprocess.Process] = None

    @property
    def is_running(self) -> bool:
        """Check if server is running"""
        return self._process is not None and self._process.returncode is None

    @property
    def url(self) -> str:
        """Get server URL"""
        return f"http://localhost:{self.port}"

    async def start(self) -> bool:
        """Start the Go server"""
        if self.is_running:
            return True

        try:
            self._process = await asyncio.create_subprocess_exec(
                "go", "run", "main.go",
                "-directory", str(self.project_dir),
                "-port", str(self.port),
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )

            # Wait for server to start
            for _ in range(30):
                if self._process.returncode is not None:
                    # Process died
                    return False

                if await self._test_health():
                    return True

                await asyncio.sleep(1)

            return False

        except Exception:
            return False

    async def stop(self):
        """Stop the Go server"""
        if self._process:
            self._process.terminate()
            await self._process.wait()
            self._process = None

    async def _test_health(self) -> bool:
        """Test if server is responding"""
        try:
            import aiohttp
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.url}/health", timeout=2) as response:
                    return response.status < 500
        except:
            return False

    def get_status(self) -> dict:
        """Get server status"""
        return {
            "running": self.is_running,
            "port": self.port,
            "url": self.url,
            "project_dir": str(self.project_dir)
        }