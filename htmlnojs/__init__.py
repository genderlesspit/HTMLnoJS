"""
HTMLnoJS - Async Python orchestrator for HTML-first web applications
"""

from .core import htmlnojs, HTMLnoJS, get, list_instances, stop_all
from .go_server import GoServer
from .htmx_server import HTMXServer
from .port_manager import PortManager
from .instance_registry import InstanceRegistry

__version__ = "0.1.0"
__author__ = "HTMLnoJS Team"
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
    "ProjectManager",
    "GoServer",
    "HTMXServer",
    "PortManager",
    "InstanceRegistry",
]