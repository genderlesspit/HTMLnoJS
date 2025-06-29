"""
HTMLnoJS Core - Main application class that orchestrates all components
"""
import uuid
from pathlib import Path
from typing import Optional, Dict, Any
from dataclasses import dataclass

from .dependency_manager import DependencyManager
from .project_manager import ProjectManager
from .go_server import GoServer
from .htmx_server import HTMXServer
from .port_manager import PortManager
from .instance_registry import InstanceRegistry


@dataclass
class HTMLnoJSConfig:
    """Configuration for HTMLnoJS application"""
    project_dir: str
    go_port: int
    python_port: int
    host: str = "127.0.0.1"


class HTMLnoJS:
    """
    Main HTMLnoJS application class
    Orchestrates Go server, Python HTMX server, and project management
    """

    def __init__(self, project_dir: str = ".", port: Optional[int] = None, alias: Optional[str] = None):
        self.id = alias or str(uuid.uuid4())[:8]
        self.project_dir = Path(project_dir).resolve()

        # Allocate ports
        if port:
            self.go_port = port
            self.python_port = port + 1
        else:
            self.go_port, self.python_port = PortManager.allocate_port_pair(8080)

        self.config = HTMLnoJSConfig(
            project_dir=str(self.project_dir),
            go_port=self.go_port,
            python_port=self.python_port
        )

        # Initialize components
        self.dependency_manager = DependencyManager()
        self.project_manager = ProjectManager(str(self.project_dir))

        python_url = f"http://{self.config.host}:{self.python_port}"
        self.go_server = GoServer(str(self.project_dir), self.go_port, python_url)
        self.htmx_server = HTMXServer(str(self.project_dir), self.python_port)

        # Register instance
        InstanceRegistry.register(self)

    @property
    def is_running(self) -> bool:
        """Check if both servers are running"""
        return self.go_server.is_running and self.htmx_server.is_running

    async def check_dependencies(self) -> bool:
        """Check if dependencies are available"""
        return await self.dependency_manager.check_go_available()

    async def ensure_dependencies(self, timeout: int = 300) -> bool:
        """Ensure dependencies are available"""
        return await self.dependency_manager.ensure_dependencies(timeout)

    async def setup_project(self) -> bool:
        """Setup project structure"""
        return await self.project_manager.ensure_structure()

    async def start_go_server(self) -> bool:
        """Start Go server only"""
        return await self.go_server.start()

    async def start_htmx_server(self) -> bool:
        """Start HTMX server only"""
        return await self.htmx_server.start()

    async def start(self, wait_for_deps: bool = True) -> bool:
        """Start the complete HTMLnoJS system"""
        # Ensure dependencies
        if wait_for_deps:
            if not await self.ensure_dependencies():
                return False

        # Setup project
        if not await self.setup_project():
            return False

        # Start HTMX server first
        if not await self.start_htmx_server():
            return False

        # Start Go server
        if not await self.start_go_server():
            await self.htmx_server.stop()
            return False

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
            "running": self.is_running,
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
def htmlnojs(project_dir: str = ".", port: Optional[int] = None, alias: Optional[str] = None) -> HTMLnoJS:
    """
    Create HTMLnoJS application instance

    Usage:
        app = htmlnojs("./my-project")
        await app.start()
    """
    return HTMLnoJS(project_dir, port, alias)


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