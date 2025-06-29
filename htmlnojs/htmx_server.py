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
        route = e.get("route")
        fn_name = e.get("function")
        parts = route.strip("/").split("/")
        module = parts[1] if len(parts) > 1 else None
        file_path = project_dir.joinpath("py_htmx", f"{module}.py")
        log.debug(f"Mounting Python route {route} -> {fn_name} from {file_path}")

        spec = importlib.util.spec_from_file_location(
            pathlib.Path(file_path).stem, file_path.as_posix()
        )
        mod = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(mod)
        fn = getattr(mod, fn_name)

        async def handler(request: Request, fn=fn):
            if request.headers.get("content-type", "").startswith("application/json"):
                data = await request.json()
            else:
                data = dict(await request.form())
            return fn(data)

        app.add_api_route(route, handler, methods=[e.get("method")])

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
            lines.append(f"PYTHON {p.get('method')} {p.get('route')} -> {p.get('function')}")
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
