"""
HTMLnoJS Core - Main application class that orchestrates all components
"""
import uuid
from pathlib import Path
from typing import Optional, Dict, Any
from dataclasses import dataclass

from async_property import async_cached_property, async_property

from .go_server import GoServer
from .htmx_server import HTMXServer
from .port_manager import PortManager
from .instance_registry import InstanceRegistry

from loguru import logger as log

class HTMLnoJS:
    """
    Main HTMLnoJS application class
    Orchestrates Go server, Python HTMX server, and project management
    """

    def __init__(self, project_dir: str = ".", port: int = 3000, alias: Optional[str] = None, verbose: Optional[bool] = False):
        self.verbose = verbose

        self.id = alias or str(uuid.uuid4())[:8]
        self.project_dir = Path(project_dir).resolve()

        self.port_manager = PortManager()
        self.go_port, self.python_port = self.port_manager.allocate_port_pair(port)
        if verbose: log.debug(f"{self}: Initialized ports:\ngo_port={self.go_port}\npython_port={self.python_port}")

        python_url = f"http://localhost:{self.python_port}"

        self.go_server = GoServer(str(self.project_dir), self.go_port, self.python_port)
        self.go_server.verbose = verbose

        self.htmx_server = HTMXServer(str(self.project_dir), self.python_port)
        self.htmx_server.verbose = verbose

        InstanceRegistry.register(self)

        if verbose: log.success(f"{self}: Successfully initialized!\n{self.__dict__}")

    def __repr__(self):
        return f"HTMLnoJS.{self.id}"

    async def start(self, wait_for_deps: bool = True) -> bool:
        """Start the complete HTMLnoJS system"""

        await self.go_server.is_running()
        await self.htmx_server.is_running()

        return True

    async def stop(self):
        """Stop all servers"""
        await self.go_server.stop()
        await self.htmx_server.stop()

    def status(self) -> Dict[str, Any]:
        """Get system status"""
        return {
            "id": self.id,
            "project_dir": str(self.project_dir),
            "go_port": self.go_port,
            "python_port": self.python_port,
            "go_server": self.go_server.get_status(),
            "htmx_server": self.htmx_server.get_status(),
            "urls": {
                "go": self.go_server.url,
                "htmx": self.htmx_server.url,
                "routes": f"{self.go_server.url}/_routes",
                "health": f"{self.go_server.url}/health"
            }
        }

    async def __aenter__(self):
        """Async context manager entry"""
        await self.start()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit"""
        await self.stop()

    def __del__(self):
        """Cleanup on deletion"""
        InstanceRegistry.unregister(self.id)


# Factory function - main API entry point
def htmlnojs(project_dir: str = ".", port: int = 3000, alias: Optional[str] = None, verbose: Optional[bool] = False) -> HTMLnoJS:
    """
    Create HTMLnoJS application instance

    Usage:
        app = htmlnojs("./my-project")
        await app.start()
    """
    return HTMLnoJS(project_dir, port, alias, verbose)


# Convenience functions
def get(alias: str) -> Optional[HTMLnoJS]:
    """Get HTMLnoJS instance by alias"""
    return InstanceRegistry.get(alias)


def list_instances():
    """List all HTMLnoJS instances"""
    return InstanceRegistry.list_all()


async def stop_all():
    """Stop all HTMLnoJS instances"""
    await InstanceRegistry.stop_all()