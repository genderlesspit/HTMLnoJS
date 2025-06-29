import json, threading, time
from pathlib import Path
from fastapi import FastAPI, Request
from fastapi.responses import PlainTextResponse, JSONResponse
import uvicorn
from functools import cached_property
from typing import Optional, Dict, Any
from loguru import logger as log
import importlib.util


def create_app(registry_path: str = "registry.json") -> FastAPI:
    app = FastAPI()
    reg = json.loads(Path(registry_path).read_text())
    # dynamic Python handlers
    for e in reg:
        if e.get("type") != "python":
            continue
        spec = importlib.util.spec_from_file_location(
            Path(e["file"]).stem,
            e["file"],
        )
        mod = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(mod)
        fn = getattr(mod, e["function"])

        async def handler(request: Request, fn=fn):
            if request.headers.get("content-type", "").startswith("application/json"):
                data = await request.json()
            else:
                data = dict(await request.form())
            return fn(data)

        app.add_api_route(e["route"], handler, methods=[e["method"]])

    # human-readable route map
    @app.get("/_routes", response_class=PlainTextResponse)
    def routes_text():
        lines = ["=== HTMX FastAPI Route Map ===", ""]
        for e in reg:
            t = e.get("type", "").upper()
            m = e.get("method", "")
            r = e.get("route", "")
            f = e.get("function") if t == "PYTHON" else e.get("file")
            lines.append(f"{t} {m} {r} -> {f}")
        lines.append(f"\nTOTAL ROUTES {len(reg)}")
        return "\n".join(lines)

    # machine-readable route map
    @app.get("/_routes.json", response_class=JSONResponse)
    def routes_json():
        return reg

    return app


# module-level app
app = create_app()


class HTMXServer:
    def __init__(self, project_dir: str, port: int, verbose: bool = True):
        self.project_dir = Path(project_dir)
        self.port = port
        self.verbose = verbose
        self._server: Optional[uvicorn.Server] = None
        self._started = False

    def __repr__(self) -> str:
        return f"<HTMXServer port={self.port} dir={self.project_dir}>"

    @cached_property
    def thread(self) -> threading.Thread:
        def run():
            cfg = uvicorn.Config(app, host="127.0.0.1", port=self.port, log_level="info")
            self._server = uvicorn.Server(cfg)
            self._started = True
            if self.verbose:
                log.debug(f"Starting HTMXServer on http://127.0.0.1:{self.port}")
            self._server.run()

        t = threading.Thread(target=run, daemon=True)
        t.start()
        time.sleep(1)
        return t

    async def start(self) -> None:
        if not self.thread.is_alive() and self.verbose:
            log.debug("launching HTMXServer thread")
        _ = self.thread

    async def is_running(self) -> bool:
        if not self._started:
            await self.start()
        try:
            import aiohttp
            async with aiohttp.ClientSession() as s:
                r = await s.get(f"http://127.0.0.1:{self.port}/health", timeout=2)
                return r.status < 500
        except Exception:
            return False

    async def stop(self) -> None:
        if self._server and self._server.started:
            if self.verbose:
                log.debug("shutting down HTMXServer")
            await self._server.shutdown()
            self._started = False
            log.success("HTMXServer stopped")

    def get_status(self) -> Dict[str, Any]:
        return {"url": f"http://127.0.0.1:{self.port}", "port": self.port, "running": self._started}
