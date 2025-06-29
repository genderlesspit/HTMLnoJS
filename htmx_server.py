import os
import shutil
import subprocess
import asyncio
import uuid
import socket
from pathlib import Path
from typing import Optional, Dict, Any
from dataclasses import dataclass


@dataclass
class ServerConfig:
    """Configuration for HTMLnoJS servers"""
    project_dir: str
    go_port: int
    python_port: int
    host: str = "127.0.0.1"


class HTMLnoJS:
    """
    Python orchestrator for HTMLnoJS framework
    Manages Go server and coordinates with separate HTMX server
    """

    instances: Dict[str, 'HTMLnoJS'] = {}

    def __init__(self, project_dir: str = ".", port: Optional[int] = None, alias: Optional[str] = None):
        self.id = alias or str(uuid.uuid4())[:8]
        self.project_dir = Path(project_dir).resolve()

        # Find available ports
        self.go_port = port or self._find_available_port(8080)
        self.python_port = self.go_port + 1

        self.config = ServerConfig(
            project_dir=str(self.project_dir),
            go_port=self.go_port,
            python_port=self.python_port
        )

        self._go_process = None
        self._python_process = None
        self._dependency_check_complete = False

        # Store instance
        HTMLnoJS.instances[self.id] = self

    @classmethod
    def create(cls, project_dir: str = ".", port: Optional[int] = None, alias: Optional[str] = None) -> 'HTMLnoJS':
        """Factory method to create HTMLnoJS instance"""
        return cls(project_dir, port, alias)

    @classmethod
    def get(cls, alias: str) -> Optional['HTMLnoJS']:
        """Get existing instance by alias"""
        return cls.instances.get(alias)

    def _find_available_port(self, start_port: int = 8080) -> int:
        """Find an available port starting from start_port"""
        for port in range(start_port, start_port + 100):
            try:
                with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
                    sock.bind(('', port))
                    return port
            except OSError:
                continue
        raise RuntimeError("No available ports found")

    async def _install_go_dependencies(self) -> bool:
        """Install Go dependencies asynchronously"""
        print("📦 Installing Go dependencies...")

        try:
            # Run PowerShell dependencies script if it exists
            deps_script = Path("dependencies.ps1")
            if deps_script.exists():
                print("🔧 Running dependencies.ps1...")
                process = await asyncio.create_subprocess_exec(
                    "powershell", "-ExecutionPolicy", "Bypass",
                    "-File", str(deps_script),
                    stdout=asyncio.subprocess.PIPE,
                    stderr=asyncio.subprocess.PIPE
                )
                stdout, stderr = await process.communicate()

                if process.returncode == 0:
                    print("✅ Dependencies script completed")
                    return True
                else:
                    print(f"❌ Dependencies script failed: {stderr.decode()}")
                    return False
            else:
                print("❌ dependencies.ps1 not found")
                return False

        except Exception as e:
            print(f"❌ Dependencies installation error: {e}")
            return False

    async def _wait_for_go(self, timeout: int = 300) -> bool:
        """Wait for Go to be available"""
        print("⏳ Waiting for Go installation...")
        start_time = asyncio.get_event_loop().time()

        while asyncio.get_event_loop().time() - start_time < timeout:
            if shutil.which("go") is not None:
                try:
                    process = await asyncio.create_subprocess_exec(
                        "go", "version",
                        stdout=asyncio.subprocess.PIPE,
                        stderr=asyncio.subprocess.PIPE
                    )
                    stdout, stderr = await process.communicate()

                    if process.returncode == 0:
                        print(f"✅ Go is ready: {stdout.decode().strip()}")
                        return True
                except Exception:
                    pass

            print("⏳ Go not ready yet, waiting...")
            await asyncio.sleep(5)

        print("❌ Timeout waiting for Go installation")
        return False

    async def check_dependencies(self) -> bool:
        """Check if all dependencies are available"""
        if self._dependency_check_complete:
            return True

        print("🔍 Checking dependencies...")

        # Check Go
        go_available = shutil.which("go") is not None
        if go_available:
            try:
                process = await asyncio.create_subprocess_exec(
                    "go", "version",
                    stdout=asyncio.subprocess.PIPE,
                    stderr=asyncio.subprocess.PIPE
                )
                stdout, stderr = await process.communicate()

                if process.returncode == 0:
                    print(f"✅ Go: {stdout.decode().strip()}")
                else:
                    go_available = False
            except Exception:
                go_available = False

        if not go_available:
            print("❌ Go not found")

        if go_available:
            self._dependency_check_complete = True
            print("✅ All dependencies satisfied!")
        else:
            print("❌ Some dependencies missing")

        return go_available

    async def wait_for_dependencies(self, timeout: int = 300) -> bool:
        """Wait for dependencies to be installed"""
        if await self.check_dependencies():
            return True

        print(f"⏳ Installing dependencies (timeout: {timeout}s)...")

        # Install dependencies
        if not await self._install_go_dependencies():
            return False

        # Wait for Go to be available
        return await self._wait_for_go(timeout)

    async def setup_project(self) -> bool:
        """Setup project directory structure"""
        print(f"🏗️ Setting up project structure in {self.project_dir}")

        # Ensure dependencies are available first
        if not await self.wait_for_dependencies():
            print("❌ Cannot setup project - dependencies not available")
            return False

        try:
            process = await asyncio.create_subprocess_exec(
                "go", "run", "main.go",
                "-directory", str(self.project_dir),
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )

            stdout, stderr = await asyncio.wait_for(
                process.communicate(), timeout=60
            )

            if process.returncode == 0:
                print("✅ Project structure created")
                if stdout.decode().strip():
                    print(stdout.decode())
                return True
            else:
                print(f"❌ Setup failed: {stderr.decode()}")
                return False

        except asyncio.TimeoutError:
            print("❌ Setup timed out")
            return False
        except Exception as e:
            print(f"❌ Setup error: {e}")
            return False

    async def start_go_server(self) -> bool:
        """Start the Go server"""
        if self._go_process and self._go_process.returncode is None:
            print(f"✅ Go server already running on port {self.go_port}")
            return True

        # Ensure dependencies are available
        if not await self.wait_for_dependencies():
            print("❌ Cannot start Go server - dependencies not available")
            return False

        print(f"🚀 Starting Go server on port {self.go_port}...")

        try:
            self._go_process = await asyncio.create_subprocess_exec(
                "go", "run", "main.go",
                "-directory", str(self.project_dir),
                "-port", str(self.go_port),
                "-python=false",  # We'll manage Python server separately
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )

            # Wait for server to start
            for i in range(30):
                if await self._test_server(self.go_port, "/health"):
                    print(f"✅ Go server ready at http://localhost:{self.go_port}")
                    return True
                print(f"⏳ Waiting for Go server... ({i+1}/30)")
                await asyncio.sleep(1)

            print("❌ Go server failed to start")
            return False

        except Exception as e:
            print(f"❌ Failed to start Go server: {e}")
            return False

    async def start_python_server(self) -> bool:
        """Start the Python HTMX server"""
        if self._python_process and self._python_process.returncode is None:
            print(f"✅ Python server already running on port {self.python_port}")
            return True

        print(f"🐍 Starting Python HTMX server on port {self.python_port}...")

        # Create the htmx_server.py if it doesn't exist
        htmx_server_path = self.project_dir / "htmx_server.py"
        if not htmx_server_path.exists():
            self._create_htmx_server_file(htmx_server_path)

        try:
            self._python_process = await asyncio.create_subprocess_exec(
                "python", str(htmx_server_path),
                "--port", str(self.python_port),
                "--project-dir", str(self.project_dir),
                cwd=str(self.project_dir),
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )

            # Wait for server to start
            for i in range(30):
                if await self._test_server(self.python_port, "/docs"):
                    print(f"✅ Python server ready at http://localhost:{self.python_port}")
                    return True
                print(f"⏳ Waiting for Python server... ({i+1}/30)")
                await asyncio.sleep(1)

            print("❌ Python server failed to start")
            return False

        except Exception as e:
            print(f"❌ Failed to start Python server: {e}")
            return False

    def _create_htmx_server_file(self, htmx_server_path: Path):
        """Create the separate htmx_server.py file"""
        htmx_server_content = '''#!/usr/bin/env python3
"""
Separate HTMX server for HTMLnoJS framework
This server handles Python-based HTMX endpoints
"""

import sys
import argparse
from pathlib import Path
from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse
import uvicorn


def create_app(project_dir: Path) -> FastAPI:
    """Create FastAPI application with HTMX handlers"""
    app = FastAPI(title="HTMLnoJS HTMX Server")
    
    # Add py_htmx directory to Python path
    py_htmx_dir = project_dir / "py_htmx"
    if py_htmx_dir.exists():
        sys.path.insert(0, str(py_htmx_dir))
    
    # Auto-discover htmx_ functions
    htmx_handlers = {}
    
    if py_htmx_dir.exists():
        for py_file in py_htmx_dir.glob("*.py"):
            if py_file.name.startswith("__"):
                continue
            
            module_name = py_file.stem
            try:
                module = __import__(module_name)
                
                for attr_name in dir(module):
                    if attr_name.startswith("htmx_"):
                        func = getattr(module, attr_name)
                        if callable(func):
                            route_name = attr_name[5:]  # Remove htmx_ prefix
                            route_path = f"/{module_name}/{route_name}"
                            htmx_handlers[route_path] = func
                            print(f"📍 Registered: {route_path} -> {attr_name}")
            except Exception as e:
                print(f"❌ Failed to import {module_name}: {e}")
    
    @app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE"])
    async def handle_htmx(path: str, request: Request):
        route_path = f"/{path}"
        
        if route_path in htmx_handlers:
            handler = htmx_handlers[route_path]
            
            # Prepare request data
            request_data = dict(request.query_params)
            
            if request.method == "POST":
                form_data = await request.form()
                request_data.update(form_data)
            
            # Add headers
            for key, value in request.headers.items():
                request_data[f"HTTP_{key.upper().replace('-', '_')}"] = value
            
            try:
                result = handler(request_data)
                return HTMLResponse(content=str(result))
            except Exception as e:
                return HTMLResponse(
                    content=f"<div class='error'>Error in {route_path}: {e}</div>", 
                    status_code=500
                )
        
        return HTMLResponse(
            content=f"<div class='error'>Handler not found: {route_path}</div>", 
            status_code=404
        )
    
    return app


def main():
    parser = argparse.ArgumentParser(description="HTMLnoJS HTMX Server")
    parser.add_argument("--port", type=int, default=8081, help="Server port")
    parser.add_argument("--project-dir", type=str, default=".", help="Project directory")
    parser.add_argument("--host", type=str, default="127.0.0.1", help="Server host")
    
    args = parser.parse_args()
    
    project_dir = Path(args.project_dir).resolve()
    app = create_app(project_dir)
    
    print(f"🐍 Starting HTMX server on {args.host}:{args.port}")
    print(f"📁 Project directory: {project_dir}")
    
    uvicorn.run(app, host=args.host, port=args.port, log_level="info")


if __name__ == "__main__":
    main()
'''
        htmx_server_path.write_text(htmx_server_content.strip())
        print(f"📄 Created {htmx_server_path}")

    async def _test_server(self, port: int, path: str = "/") -> bool:
        """Test if server is responding"""
        try:
            import aiohttp
            async with aiohttp.ClientSession() as session:
                async with session.get(f"http://localhost:{port}{path}", timeout=2) as response:
                    return response.status < 500
        except:
            return False

    async def start(self, wait_for_deps: bool = True) -> bool:
        """Start the complete HTMLnoJS system"""
        print(f"🎯 Starting HTMLnoJS system [{self.id}]")
        print(f"📁 Project: {self.project_dir}")
        print(f"🌐 Ports: Go={self.go_port}, Python={self.python_port}")

        # Wait for dependencies if requested
        if wait_for_deps:
            if not await self.wait_for_dependencies():
                print("❌ Failed to install dependencies")
                return False

        # Setup project if needed
        if not (self.project_dir / "py_htmx").exists():
            if not await self.setup_project():
                return False

        # Start servers
        if not await self.start_go_server():
            return False

        if not await self.start_python_server():
            return False

        print("🎉 HTMLnoJS system ready!")
        print(f"🔗 Visit: http://localhost:{self.go_port}")
        print(f"📊 Routes: http://localhost:{self.go_port}/_routes")
        print(f"🐍 Python API: http://localhost:{self.python_port}/docs")

        return True

    async def stop(self):
        """Stop all servers"""
        print(f"🛑 Stopping HTMLnoJS system [{self.id}]")

        if self._go_process:
            self._go_process.terminate()
            await self._go_process.wait()
            self._go_process = None

        if self._python_process:
            self._python_process.terminate()
            await self._python_process.wait()
            self._python_process = None

        print("✅ All servers stopped")

    def status(self) -> Dict[str, Any]:
        """Get system status"""
        return {
            "id": self.id,
            "project_dir": str(self.project_dir),
            "go_port": self.go_port,
            "python_port": self.python_port,
            "go_running": self._go_process and self._go_process.returncode is None,
            "python_running": self._python_process and self._python_process.returncode is None,
            "dependencies_ready": self._dependency_check_complete,
            "go_url": f"http://localhost:{self.go_port}",
            "python_url": f"http://localhost:{self.python_port}",
        }

    async def __aenter__(self):
        """Async context manager entry"""
        await self.start()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit"""
        await self.stop()


# Convenience aliases
htmlnojs = HTMLnoJS.create
create = HTMLnoJS.create
get = HTMLnoJS.get


async def main():
    """CLI usage"""
    import sys

    if len(sys.argv) < 2:
        print("Usage: python htmlnojs.py <project_dir> [port]")
        sys.exit(1)

    project_dir = sys.argv[1]
    port = int(sys.argv[2]) if len(sys.argv) > 2 else None

    async with HTMLnoJS(project_dir, port) as system:
        try:
            print("Press Ctrl+C to stop...")
            while True:
                await asyncio.sleep(1)
        except KeyboardInterrupt:
            print("\n👋 Shutting down...")


if __name__ == "__main__":
    asyncio.run(main())