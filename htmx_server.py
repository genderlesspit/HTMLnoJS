"""
HTMX Server Manager - Handles Python HTMX server lifecycle
"""
import asyncio
from pathlib import Path
from typing import Optional


class HTMXServer:
    """Manages Python HTMX server process"""
    
    def __init__(self, project_dir: str, port: int):
        self.project_dir = Path(project_dir).resolve()
        self.port = port
        self._process: Optional[asyncio.subprocess.Process] = None
        self._server_file = self.project_dir / "htmx_server.py"
    
    @property
    def is_running(self) -> bool:
        """Check if server is running"""
        return self._process is not None and self._process.returncode is None
    
    @property
    def url(self) -> str:
        """Get server URL"""
        return f"http://localhost:{self.port}"
    
    def ensure_server_file(self):
        """Create HTMX server file if it doesn't exist"""
        if self._server_file.exists():
            return
        
        server_content = '''#!/usr/bin/env python3
"""
Standalone HTMX server for HTMLnoJS framework
"""

import sys
import argparse
import asyncio
from pathlib import Path
from typing import Dict, Callable, Any
from fastapi import FastAPI, Request, HTTPException
from fastapi.responses import HTMLResponse, JSONResponse
from fastapi.middleware.cors import CORSMiddleware
import uvicorn


class HTMXServerApp:
    """Standalone HTMX server"""
    
    def __init__(self, project_dir: Path):
        self.project_dir = project_dir
        self.py_htmx_dir = project_dir / "py_htmx"
        self.htmx_handlers: Dict[str, Callable] = {}
        self.app = self._create_app()
    
    def _create_app(self) -> FastAPI:
        """Create FastAPI application"""
        app = FastAPI(title="HTMLnoJS HTMX Server")
        
        app.add_middleware(
            CORSMiddleware,
            allow_origins=["*"],
            allow_credentials=True,
            allow_methods=["*"],
            allow_headers=["*"],
        )
        
        self._register_routes(app)
        self._discover_htmx_handlers()
        
        return app
    
    def _register_routes(self, app: FastAPI):
        """Register routes"""
        
        @app.get("/health")
        async def health_check():
            return {"status": "healthy", "handlers": len(self.htmx_handlers)}
        
        @app.get("/handlers")
        async def list_handlers():
            return {"handlers": list(self.htmx_handlers.keys())}
        
        @app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE"])
        async def handle_htmx_request(path: str, request: Request):
            route_path = f"/{path}"
            
            if route_path not in self.htmx_handlers:
                raise HTTPException(status_code=404, detail=f"Handler not found: {route_path}")
            
            handler = self.htmx_handlers[route_path]
            
            try:
                request_data = await self._prepare_request_data(request)
                result = await self._call_handler(handler, request_data)
                
                if isinstance(result, dict):
                    return JSONResponse(content=result)
                else:
                    return HTMLResponse(content=str(result))
                    
            except Exception as e:
                return HTMLResponse(
                    content=f"<div class='error'>Error in {route_path}: {str(e)}</div>", 
                    status_code=500
                )
    
    async def _prepare_request_data(self, request: Request) -> Dict[str, Any]:
        """Prepare request data for handler"""
        request_data = dict(request.query_params)
        
        if request.method == "POST":
            try:
                if "application/json" in request.headers.get("content-type", ""):
                    json_data = await request.json()
                    request_data.update(json_data)
                else:
                    form_data = await request.form()
                    request_data.update(form_data)
            except Exception:
                pass
        
        # Add headers
        for key, value in request.headers.items():
            request_data[f"HTTP_{key.upper().replace('-', '_')}"] = value
        
        return request_data
    
    async def _call_handler(self, handler: Callable, request_data: Dict[str, Any]) -> Any:
        """Call handler function"""
        if asyncio.iscoroutinefunction(handler):
            return await handler(request_data)
        else:
            loop = asyncio.get_event_loop()
            return await loop.run_in_executor(None, handler, request_data)
    
    def _discover_htmx_handlers(self):
        """Auto-discover htmx_ functions"""
        if not self.py_htmx_dir.exists():
            return
        
        sys.path.insert(0, str(self.py_htmx_dir))
        
        for py_file in self.py_htmx_dir.glob("*.py"):
            if py_file.name.startswith("__"):
                continue
            
            module_name = py_file.stem
            
            try:
                module = __import__(module_name)
                
                for attr_name in dir(module):
                    if attr_name.startswith("htmx_"):
                        func = getattr(module, attr_name)
                        if callable(func):
                            route_name = attr_name[5:]
                            route_path = f"/{module_name}/{route_name}"
                            self.htmx_handlers[route_path] = func
                            
            except Exception as e:
                print(f"Failed to import {module_name}: {e}")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", type=int, default=8081)
    parser.add_argument("--project-dir", type=str, default=".")
    parser.add_argument("--host", type=str, default="127.0.0.1")
    
    args = parser.parse_args()
    
    project_dir = Path(args.project_dir).resolve()
    server = HTMXServerApp(project_dir)
    
    uvicorn.run(server.app, host=args.host, port=args.port, log_level="info")


if __name__ == "__main__":
    main()
'''
        self._server_file.write_text(server_content)
    
    async def start(self) -> bool:
        """Start the HTMX server"""
        if self.is_running:
            return True
        
        self.ensure_server_file()
        
        try:
            self._process = await asyncio.create_subprocess_exec(
                "python", str(self._server_file),
                "--port", str(self.port),
                "--project-dir", str(self.project_dir),
                cwd=str(self.project_dir),
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
            
            # Wait for server to start
            for _ in range(30):
                if self._process.returncode is not None:
                    return False
                
                if await self._test_health():
                    return True
                
                await asyncio.sleep(1)
            
            return False
            
        except Exception:
            return False
    
    async def stop(self):
        """Stop the HTMX server"""
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