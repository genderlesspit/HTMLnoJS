"""
Dependency Manager - Handles Go installation and validation
"""
import shutil
import asyncio
from pathlib import Path
from typing import Optional


class DependencyManager:
    """Manages Go dependencies for HTMLnoJS"""

    def __init__(self):
        self._go_validated = False

    async def check_go_available(self) -> bool:
        """Check if Go is available and working"""
        if self._go_validated:
            return True

        if shutil.which("go") is None:
            return False

        try:
            process = await asyncio.create_subprocess_exec(
                "go", "version",
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
            stdout, stderr = await process.communicate()

            if process.returncode == 0:
                self._go_validated = True
                return True
        except Exception:
            pass

        return False

    async def install_go_dependencies(self, deps_script: Optional[Path] = None) -> bool:
        """Install Go dependencies using PowerShell script"""
        if deps_script is None:
            deps_script = Path("dependencies.ps1")

        if not deps_script.exists():
            return False

        try:
            process = await asyncio.create_subprocess_exec(
                "powershell", "-ExecutionPolicy", "Bypass",
                "-File", str(deps_script),
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
            stdout, stderr = await process.communicate()
            return process.returncode == 0
        except Exception:
            return False

    async def wait_for_go(self, timeout: int = 300) -> bool:
        """Wait for Go to become available"""
        start_time = asyncio.get_event_loop().time()

        while asyncio.get_event_loop().time() - start_time < timeout:
            if await self.check_go_available():
                return True
            await asyncio.sleep(5)

        return False

    async def ensure_dependencies(self, timeout: int = 300) -> bool:
        """Ensure all dependencies are available, installing if needed"""
        if await self.check_go_available():
            return True

        # Try to install
        if await self.install_go_dependencies():
            return await self.wait_for_go(timeout)

        return False