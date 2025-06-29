import threading, time
from fastapi import FastAPI, Request
from fastapi.responses import PlainTextResponse, JSONResponse
import uvicorn
from functools import cached_property
from typing import Optional, Dict, Any, List
from loguru import logger as log
import importlib.util
import requests
import pathlib


def create_app_from_registry_map(reg_map: Dict[str, Any], project_dir: pathlib.Path) -> FastAPI:
    """Build FastAPI app using registry map fetched from Go server."""
    app = FastAPI()

    # health endpoint
    @app.get("/health")
    def health():
        log.debug("Serving health check")
        return JSONResponse({"status": "ok"})

    html_routes = reg_map.get("html_routes", [])
    css_routes = reg_map.get("css_routes", [])
    python_routes = reg_map.get("python_routes", [])

    # mount dynamic Python handlers
    for e in python_routes:
        go_route = e.get("route")  # This is "/api/demo/hello"
        fn_name = e.get("function")
        method = e.get("method")

        # Strip /api/ prefix to match what Go server actually calls
        fastapi_route = go_route.replace("/api/", "/", 1) if go_route.startswith("/api/") else go_route

        # Extract module name from the stripped route
        parts = fastapi_route.strip("/").split("/")
        module = parts[0] if len(parts) > 0 else "demo"  # fallback to demo

        file_path = project_dir.joinpath("py_htmx", f"{module}.py")
        log.debug(f"Mounting Python route {go_route} -> FastAPI {fastapi_route} -> {fn_name} from {file_path}")

        try:
            spec = importlib.util.spec_from_file_location(
                pathlib.Path(file_path).stem, file_path.as_posix()
            )
            mod = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(mod)

            # Check if function exists
            if not hasattr(mod, fn_name):
                log.error(f"Function {fn_name} not found in {file_path}")
                continue

            fn = getattr(mod, fn_name)
            log.debug(f"Successfully loaded function {fn_name} from {module}.py")

            # Create handler with proper function binding
            def create_handler(handler_func, func_name):
                async def handler(request: Request):
                    try:
                        log.debug(f"Calling {func_name} with request")
                        log.debug(f"Content-Type: {request.headers.get('content-type', 'None')}")
                        log.debug(f"Method: {request.method}")

                        data = {}

                        # Handle different content types
                        if request.method == "POST":
                            content_type = request.headers.get("content-type", "")
                            log.debug(f"POST request with content-type: {content_type}")

                            if content_type.startswith("application/json"):
                                data = await request.json()
                                log.debug(f"JSON data: {data}")
                            elif content_type.startswith("application/x-www-form-urlencoded"):
                                form_data = await request.form()
                                data = dict(form_data)
                                log.debug(f"Form data: {data}")

                                # If form data is empty, try reading raw body
                                if not data:
                                    try:
                                        body = await request.body()
                                        log.debug(f"Raw body (form empty): {body}")
                                        if body:
                                            from urllib.parse import parse_qs
                                            body_str = body.decode('utf-8')
                                            log.debug(f"Body string: {body_str}")
                                            parsed = parse_qs(body_str)
                                            data = {k: v[0] if v else '' for k, v in parsed.items()}
                                            log.debug(f"Manually parsed data: {data}")
                                    except Exception as e:
                                        log.error(f"Failed to parse raw body: {e}")
                            else:
                                # HTMX default form submission
                                try:
                                    form_data = await request.form()
                                    data = dict(form_data)
                                    log.debug(f"Default form data: {data}")
                                except Exception as e:
                                    log.error(f"Failed to parse form data: {e}")
                                    # Try to read raw body
                                    body = await request.body()
                                    log.debug(f"Raw body: {body}")
                                    if body:
                                        # Parse form data manually
                                        from urllib.parse import parse_qs
                                        body_str = body.decode('utf-8')
                                        parsed = parse_qs(body_str)
                                        data = {k: v[0] if v else '' for k, v in parsed.items()}
                                        log.debug(f"Manually parsed data: {data}")
                        else:
                            # GET request - use query parameters
                            data = dict(request.query_params)
                            log.debug(f"Query params: {data}")

                        log.debug(f"Final data passed to {func_name}: {data}")
                        result = handler_func(data)

                        # Return HTML response
                        from fastapi.responses import HTMLResponse
                        return HTMLResponse(content=result)

                    except Exception as e:
                        log.error(f"Error in handler {func_name}: {e}")
                        return HTMLResponse(
                            content=f'<div class="alert alert-error"><strong>Error:</strong> {str(e)}</div>',
                            status_code=500
                        )
                return handler

            # Create the handler with proper function binding
            route_handler = create_handler(fn, fn_name)

            # Mount at the stripped path that Go server actually calls
            app.add_api_route(fastapi_route, route_handler, methods=[method])
            log.success(f"Successfully mounted {method} {fastapi_route} -> {fn_name}")

        except Exception as e:
            log.error(f"Failed to mount route {fastapi_route}: {e}")

    # human-readable route map
    @app.get("/_routes", response_class=PlainTextResponse)
    def routes_text():
        log.debug("Serving human-readable route map")
        lines = ["=== HTMX FastAPI Route Map ===", ""]
        for h in html_routes:
            lines.append(f"HTML GET {h.get('route')} -> {h.get('name')}")
        lines.append("")
        for c in css_routes:
            deps = c.get('dependencies', [])
            lines.append(f"CSS GET {c.get('route')} -> {c.get('name')} deps={deps}")
        lines.append("")
        for p in python_routes:
            go_route = p.get('route')
            fastapi_route = go_route.replace("/api/", "/", 1) if go_route.startswith("/api/") else go_route
            lines.append(f"PYTHON {p.get('method')} {go_route} -> FastAPI {fastapi_route} -> {p.get('function')}")
        lines.append(f"\nTOTAL ROUTES {reg_map.get('total_routes', len(python_routes))}")
        return "\n".join(lines)

    # machine-readable route map
    @app.get("/_routes.json", response_class=JSONResponse)
    def routes_json():
        log.debug("Serving JSON route map")
        return reg_map

    return app

class HTMXServer:
    """Runs FastAPI HTMX server in background, waiting on Go server."""

    def __init__(self, project_dir: str, port: int, go_port: int = 3000, host: str = "127.0.0.1", verbose: bool = True):
        self.project_dir = pathlib.Path(project_dir)
        self.port = port                # HTMX FastAPI server port
        self.go_port = go_port          # Go server port
        self.host = host                # Host for both servers
        self.verbose = verbose
        # construct base URLs from host and ports
        self.base_go_url = f"http://{self.host}:{self.go_port}"
        self.base_htmx_url = f"http://{self.host}:{self.port}"

        self._server: Optional[uvicorn.Server] = None
        if self.verbose: log.debug(f"HTMXServer config: go_url={self.base_go_url}, htmx_url={self.base_htmx_url}")

    def __repr__(self) -> str:
        return f"<HTMXServer go_url={self.base_go_url} htmx_url={self.base_htmx_url} dir={self.project_dir}>"

    @cached_property
    def thread(self) -> threading.Thread:
        def run():
            # wait for Go server health
            for i in range(20):
                try:
                    r = requests.get(f"{self.base_go_url}/health", timeout=1)
                    if r.status_code < 500:
                        if self.verbose: log.debug(f"Go server healthy: {r.status_code}")
                        break
                except Exception as err:
                    if self.verbose: log.debug(f"Waiting for Go server (attempt {i+1}): {err}")
                time.sleep(0.5)

            # fetch registry
            try:
                resp = requests.get(f"{self.base_go_url}/_routes.json", timeout=2)
                resp.raise_for_status()
                reg_map = resp.json()
                if self.verbose: log.debug(f"Loaded registry keys: {list(reg_map.keys())}")
            except Exception as err:
                log.error(f"Failed to load registry: {err}")
                reg_map = {}

            app = create_app_from_registry_map(reg_map, self.project_dir)
            cfg = uvicorn.Config(app, host=self.host, port=self.port, log_level="info")
            self._server = uvicorn.Server(cfg)
            self._started = True
            if self.verbose: log.debug(f"Starting HTMXServer on {self.base_htmx_url}")
            self._server.run()
            if self.verbose: log.debug("HTMXServer shutdown")

        t = threading.Thread(target=run, daemon=True)
        if self.verbose: log.debug("HTMXServer thread launched")
        return t

    async def is_running(self) -> bool:
        if not self.thread.is_alive():
            self.thread.start()
            
        try:
            import aiohttp
            async with aiohttp.ClientSession() as s:
                r = await s.get(f"{self.base_htmx_url}/health", timeout=2)
                ok = r.status < 500
                if self.verbose: log.debug(f"Health check at {self.base_htmx_url}/health: {r.status}")
                return ok
        except Exception as err:
            log.warning(f"Health check error: {err}")
            return False

    async def stop(self) -> None:
        if self._server and self._server.started:
            if self.verbose: log.debug("Stopping HTMXServer")
            await self._server.shutdown()
            self._started = False
            log.success("HTMXServer stopped")

    def get_status(self) -> Dict[str, Any]:
        status = {
            "go_url": self.base_go_url,
            "htmx_url": self.base_htmx_url,
            "running": self._started
        }
        if self.verbose: log.debug(f"Status: {status}")
        return status
