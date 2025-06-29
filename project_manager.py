"""
Project Manager - Handles project directory structure
"""
import asyncio
from pathlib import Path
from typing import Optional


class ProjectManager:
    """Manages HTMLnoJS project structure"""

    def __init__(self, project_dir: str):
        self.project_dir = Path(project_dir).resolve()
        self.py_htmx_dir = self.project_dir / "py_htmx"
        self.templates_dir = self.project_dir / "templates"
        self.css_dir = self.project_dir / "css"

    def exists(self) -> bool:
        """Check if project structure exists"""
        return self.py_htmx_dir.exists()

    def create_structure(self) -> bool:
        """Create basic project structure"""
        try:
            self.py_htmx_dir.mkdir(parents=True, exist_ok=True)
            self.templates_dir.mkdir(exist_ok=True)
            self.css_dir.mkdir(exist_ok=True)

            # Create example handler
            example_handler = self.py_htmx_dir / "example.py"
            if not example_handler.exists():
                example_handler.write_text('''def htmx_hello(request_data):
    """Simple hello world HTMX handler"""
    name = request_data.get('name', 'World')
    return f'<div class="greeting">Hello, {name}!</div>'

def htmx_get_time(request_data):
    """Get current time"""
    import datetime
    now = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    return f'<div class="time">Current time: {now}</div>'
''')

            return True
        except Exception:
            return False

    async def setup_via_go(self, timeout: int = 10) -> bool:
        """Try to setup project using Go main.go"""
        try:
            process = await asyncio.create_subprocess_exec(
                "go", "run", "main.go",
                "-directory", str(self.project_dir),
                "-port", "0",
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )

            try:
                await asyncio.wait_for(process.communicate(), timeout=timeout)
            except asyncio.TimeoutError:
                process.terminate()
                await process.wait()

            return self.exists()
        except Exception:
            return False

    async def ensure_structure(self) -> bool:
        """Ensure project structure exists, creating if needed"""
        if self.exists():
            return True

        # Try Go setup first
        if await self.setup_via_go():
            return True

        # Fallback to manual creation
        return self.create_structure()