import asyncio
import subprocess
import threading
import time
from pathlib import Path
from typing import Optional
from loguru import logger as log
from propcache import cached_property


class GoServer:
    """Manages Go server subprocess using a real threading.Thread"""

    def __init__(self, project_dir: str, port: int, python_port: int):
        self.project_dir = Path(project_dir).resolve()
        self.port = port
        self.python_port = python_port
        self.verbose = False

        self._popen: Optional[subprocess.Popen] = None
        self._thread: Optional[threading.Thread] = None

    def __repr__(self):
        return f"[HTMLnoJS.GoServer]"

    @property
    def url(self) -> str:
        return f"http://localhost:{self.port}"

    @cached_property
    def launcher_path(self) -> Path:
        return Path(__file__).parent / "go_server.ps1"

    @cached_property
    def thread(self) -> threading.Thread:
        """Returns an unstarted thread that runs the go-server.ps1 script"""

        def _target():
            script_path = Path(__file__).parent / "go_server.ps1"
            if not script_path.exists():
                log.error(f"{self}: script not found at {script_path}")
                return

            cmd = [
                "powershell", "-ExecutionPolicy", "Bypass",
                "-File", str(self.launcher_path),
                "-Project", str(self.project_dir),
                "-Port", str(self.port),
                "-FastAPIPort", str(self.python_port)
            ]

            proc = subprocess.Popen(
                cmd,
                cwd=str(self.project_dir),
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
            self._popen = proc

            if self.verbose:
                for line in proc.stdout:
                    log.debug(f"{self}: {line.strip()}")
                for line in proc.stderr:
                    log.debug(f"{self}: {line.strip()}")

            proc.wait()

        return threading.Thread(target=_target, daemon=True)

    async def start(self):
        if self.thread.is_alive():
            if self.verbose:
                log.debug(f"{self}: thread is running -> {self}")
                return

        if self.verbose:
            log.debug(f"{self}: launching thread -> {self}")

        self.thread.start()

        time.sleep(5)
        return

    async def is_running(self) -> bool:
        if not self.thread.is_alive():
            await self.start()

        if self.verbose:
            log.debug(f"{self}: checking {self.url}/health")

        try:
            import aiohttp
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.url}/health", timeout=1) as r:
                    if self.verbose:
                        log.debug(f"{self}: /health returned {r.status}")
                    return r.status < 500
        except Exception:
            if self.verbose:
                log.debug(f"{self}: health check failed")
            return False

    async def stop(self):
        if self._popen and self._popen.poll() is None:
            if self.verbose:
                log.debug(f"{self}: terminating subprocess")
            await asyncio.to_thread(self._popen.terminate)
            await asyncio.to_thread(self._popen.wait)
            log.success(f"{self}: process terminated")
        else:
            if self.verbose:
                log.debug(f"{self}: no active subprocess to stop")

    def get_status(self) -> dict:
        if self.verbose:
            log.debug(f"{self}: getting status")
        return {
            "url": self.url,
            "port": self.port,
            "project_dir": str(self.project_dir),
            "running": self._popen is not None and self._popen.poll() is None
        }
