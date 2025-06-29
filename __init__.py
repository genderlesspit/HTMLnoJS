"""
HTMLnoJS - Async Python orchestrator for HTML-first web applications
"""

from .core import htmlnojs, HTMLnoJS, get, list_instances, stop_all
from .dependency_manager import DependencyManager
from .project_manager import ProjectManager
from .go_server import GoServer
from .htmx_server import HTMXServer
from .port_manager import PortManager
from .instance_registry import InstanceRegistry

__version__ = "0.1.0"
__author__ = "genderlesspit"
__description__ = "Async Python orchestrator for HTML-first web applications"

# Main API exports
__all__ = [
    # Main API
    "htmlnojs",
    "HTMLnoJS",
    "get",
    "list_instances",
    "stop_all",

    # Components (for advanced usage)
    "DependencyManager",
    "ProjectManager",
    "GoServer",
    "HTMXServer",
    "PortManager",
    "InstanceRegistry",
]


# Simple demo runner
if __name__ == "__main__":
    import asyncio

    async def demo():
        print("🎯 HTMLnoJS Quick Demo")
        print("=" * 30)

        # Basic usage demo
        app = htmlnojs("./demo-project", port=8080)
        print(f"📦 Created app: {app.id}")
        print(f"📊 Config: {app.config}")

        # Show component status
        deps_ok = await app.check_dependencies()
        print(f"📦 Dependencies: {'✅' if deps_ok else '❌'}")

        project_ok = await app.setup_project()
        print(f"📁 Project setup: {'✅' if project_ok else '❌'}")

        print(f"🌐 URLs: {app.status()['urls']}")
        print("\n💡 Use: async with htmlnojs('./project') as app: ...")

    asyncio.run(demo())