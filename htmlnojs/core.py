"""
HTMLnoJS Core - Main application class that orchestrates all components
"""
import uuid
import signal
import asyncio
import atexit
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
        self._shutdown_initiated = False

        self.id = alias or str(uuid.uuid4())[:8]
        self.project_dir = Path(project_dir).resolve()

        self.port_manager = PortManager()
        self.go_port, self.python_port = self.port_manager.allocate_port_pair(port)
        if verbose: log.debug(f"{self}: Initialized ports:\ngo_port={self.go_port}\npython_port={self.python_port}")

        python_url = f"http://localhost:{self.python_port}"

        self.go_server = GoServer(str(self.project_dir), self.go_port, self.python_port)
        self.go_server.verbose = verbose

        self.htmx_server = HTMXServer(
            project_dir=str(self.project_dir),
            port=self.python_port,
            go_port=self.go_port,         # ← your real Go server port
            host="127.0.0.1",             # or localhost, whatever your host is
            verbose=self.verbose
        )

        InstanceRegistry.register(self)

        # Setup signal handlers for graceful shutdown
        self._setup_signal_handlers()

        # Register cleanup at exit
        atexit.register(self._cleanup_sync)

        if verbose: log.success(f"{self}: Successfully initialized!\n{self.__dict__}")

    def _setup_signal_handlers(self):
        """Setup signal handlers for graceful shutdown"""
        def signal_handler(signum, frame):
            if not self._shutdown_initiated:
                log.info(f"{self}: Received signal {signum}, initiating graceful shutdown...")
                self._shutdown_initiated = True
                # Run cleanup in a new event loop if needed
                try:
                    loop = asyncio.get_event_loop()
                    if loop.is_running():
                        # Schedule the cleanup
                        asyncio.create_task(self.stop())
                    else:
                        # Run cleanup directly
                        asyncio.run(self.stop())
                except RuntimeError:
                    # No event loop running, create one
                    asyncio.run(self.stop())
                except Exception as e:
                    log.error(f"{self}: Error during signal cleanup: {e}")
                    self._cleanup_sync()

        # Register signal handlers
        signal.signal(signal.SIGINT, signal_handler)   # Ctrl+C
        signal.signal(signal.SIGTERM, signal_handler)  # Termination signal

        # Windows-specific
        if hasattr(signal, 'SIGBREAK'):
            signal.signal(signal.SIGBREAK, signal_handler)  # Ctrl+Break

    def _cleanup_sync(self):
        """Synchronous cleanup for atexit"""
        if not self._shutdown_initiated:
            self._shutdown_initiated = True
            log.info(f"{self}: Running synchronous cleanup...")
            try:
                # Try to stop servers synchronously if possible
                if hasattr(self.go_server, 'stop_sync'):
                    self.go_server.stop_sync()
                if hasattr(self.htmx_server, 'stop_sync'):
                    self.htmx_server.stop_sync()
            except Exception as e:
                log.error(f"{self}: Error during sync cleanup: {e}")
            finally:
                InstanceRegistry.unregister(self.id)

    def __repr__(self):
        return f"HTMLnoJS.{self.id}"

    async def start(self, wait_for_deps: bool = True) -> bool:
        """Start the complete HTMLnoJS system"""
        if self._shutdown_initiated:
            log.warning(f"{self}: Cannot start - shutdown already initiated")
            return False

        try:
            log.info(f"{self}: Starting servers...")
            await self.go_server.is_running()
            await self.htmx_server.is_running()
            log.success(f"{self}: All servers started successfully!")
            return True
        except Exception as e:
            log.error(f"{self}: Failed to start servers: {e}")
            await self.stop()
            return False

    async def stop(self):
        """Stop all servers"""
        if self._shutdown_initiated:
            return

        self._shutdown_initiated = True
        log.info(f"{self}: Stopping all servers...")

        try:
            # Stop servers concurrently
            tasks = []
            if hasattr(self.go_server, 'stop'):
                tasks.append(self.go_server.stop())
            if hasattr(self.htmx_server, 'stop'):
                tasks.append(self.htmx_server.stop())

            if tasks:
                await asyncio.gather(*tasks, return_exceptions=True)

            log.success(f"{self}: All servers stopped")
        except Exception as e:
            log.error(f"{self}: Error stopping servers: {e}")
        finally:
            InstanceRegistry.unregister(self.id)

    async def wait_for_interrupt(self):
        """Wait for keyboard interrupt - useful for keeping the app running"""
        try:
            log.info(f"{self}: Running... Press Ctrl+C to stop")
            while not self._shutdown_initiated:
                await asyncio.sleep(1)
        except KeyboardInterrupt:
            log.info(f"{self}: Keyboard interrupt received")
        finally:
            await self.stop()

    def run_forever(self):
        """Run the application until interrupted"""
        async def _run():
            await self.start()
            await self.wait_for_interrupt()

        try:
            asyncio.run(_run())
        except KeyboardInterrupt:
            log.info(f"{self}: Application interrupted")
        except Exception as e:
            log.error(f"{self}: Application error: {e}")

    def status(self) -> Dict[str, Any]:
        """Get system status"""
        return {
            "id": self.id,
            "project_dir": str(self.project_dir),
            "go_port": self.go_port,
            "python_port": self.python_port,
            "shutdown_initiated": self._shutdown_initiated,
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
        if not self._shutdown_initiated:
            self._cleanup_sync()


# Factory function - main API entry point
def htmlnojs(project_dir: str = ".", port: int = 3000, alias: Optional[str] = None, verbose: Optional[bool] = False) -> HTMLnoJS:
    """
    Create HTMLnoJS application instance

    Usage:
        app = htmlnojs("./my-project")
        await app.start()

        # Or run forever
        app.run_forever()
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


# Global signal handler for emergency cleanup
def emergency_cleanup():
    """Emergency cleanup for all instances"""
    log.warning("Emergency cleanup - stopping all HTMLnoJS instances")
    try:
        asyncio.run(stop_all())
    except Exception as e:
        log.error(f"Emergency cleanup failed: {e}")

# Register emergency cleanup
atexit.register(emergency_cleanup)