"""
HTMLnoJS CLI - Command line interface for development
"""
import asyncio
import argparse
import signal
from pathlib import Path
from typing import Optional

from .core import htmlnojs, list_instances, stop_all
from .instance_registry import InstanceRegistry


class HTMLnoJSCLI:
    """Command line interface for HTMLnoJS"""

    def __init__(self):
        self.running = True

    async def run_server(self, project_dir: str, port: Optional[int] = None, alias: Optional[str] = None):
        """Run HTMLnoJS server"""
        print(f"🚀 Starting HTMLnoJS server")
        print(f"📁 Project: {Path(project_dir).resolve()}")

        # Setup signal handlers
        def signal_handler():
            print("\n🛑 Shutdown signal received...")
            self.running = False

        loop = asyncio.get_event_loop()
        for sig in [signal.SIGTERM, signal.SIGINT]:
            loop.add_signal_handler(sig, signal_handler)

        try:
            async with htmlnojs(project_dir, port, alias) as app:
                print(f"✅ Server ready!")
                print(f"🌐 Visit: {app.go_server.url}")
                print(f"📊 Routes: {app.go_server.url}/_routes")
                print(f"🐍 HTMX API: {app.htmx_server.url}/docs")
                print(f"💚 Health: {app.go_server.url}/health")
                print("\nPress Ctrl+C to stop...")

                # Keep running until signal
                while self.running:
                    await asyncio.sleep(1)

        except KeyboardInterrupt:
            print("\n🛑 Keyboard interrupt received...")
        except Exception as e:
            print(f"❌ Error: {e}")
        finally:
            print("👋 Goodbye!")

    async def status_command(self):
        """Show status of all instances"""
        summary = InstanceRegistry.get_status_summary()

        print("📊 HTMLnoJS Status")
        print("=" * 30)
        print(f"Total instances: {summary['total_instances']}")
        print(f"Running instances: {summary['running_instances']}")

        if summary['instances']:
            print("\nInstances:")
            for instance in summary['instances']:
                status = "🟢" if instance['running'] else "🔴"
                print(f"  {status} {instance['id']}")
                print(f"    Project: {instance['project_dir']}")
                print(f"    Ports: Go={instance['ports']['go']}, Python={instance['ports']['python']}")
        else:
            print("\nNo instances found")

    async def stop_command(self, instance_id: Optional[str] = None):
        """Stop instances"""
        if instance_id:
            instance = InstanceRegistry.get(instance_id)
            if instance:
                await instance.stop()
                print(f"✅ Stopped instance: {instance_id}")
            else:
                print(f"❌ Instance not found: {instance_id}")
        else:
            await stop_all()
            print("✅ Stopped all instances")

    async def create_project(self, project_dir: str):
        """Create new HTMLnoJS project"""
        print(f"🏗️ Creating HTMLnoJS project: {project_dir}")

        app = htmlnojs(project_dir)

        try:
            # Check dependencies
            deps_ok = await app.check_dependencies()
            if not deps_ok:
                print("📦 Installing dependencies...")
                deps_ok = await app.ensure_dependencies()

            if deps_ok:
                print("✅ Dependencies ready")

                # Setup project
                project_ok = await app.setup_project()
                if project_ok:
                    print("✅ Project structure created")
                    print(f"📁 Created: {app.project_manager.py_htmx_dir}")
                    print(f"📁 Created: {app.project_manager.templates_dir}")
                    print(f"📁 Created: {app.project_manager.css_dir}")
                    print("\n🎯 Next steps:")
                    print(f"  1. Add Python handlers to: {app.project_manager.py_htmx_dir}")
                    print(f"  2. Run: htmlnojs serve {project_dir}")
                else:
                    print("❌ Failed to create project structure")
            else:
                print("❌ Dependencies not available")

        finally:
            await app.stop()


def main():
    """Main CLI entry point"""
    parser = argparse.ArgumentParser(description="HTMLnoJS Development Server")
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Serve command
    serve_parser = subparsers.add_parser("serve", help="Start development server")
    serve_parser.add_argument("project_dir", nargs="?", default=".", help="Project directory")
    serve_parser.add_argument("--port", type=int, help="Server port")
    serve_parser.add_argument("--alias", help="Instance alias")

    # Status command
    subparsers.add_parser("status", help="Show instance status")

    # Stop command
    stop_parser = subparsers.add_parser("stop", help="Stop instances")
    stop_parser.add_argument("instance_id", nargs="?", help="Instance ID to stop (all if not specified)")

    # Create command
    create_parser = subparsers.add_parser("create", help="Create new project")
    create_parser.add_argument("project_dir", help="Project directory to create")

    args = parser.parse_args()

    cli = HTMLnoJSCLI()

    try:
        if args.command == "serve" or args.command is None:
            # Default to serve
            project_dir = getattr(args, 'project_dir', '.')
            port = getattr(args, 'port', None)
            alias = getattr(args, 'alias', None)
            asyncio.run(cli.run_server(project_dir, port, alias))

        elif args.command == "status":
            asyncio.run(cli.status_command())

        elif args.command == "stop":
            instance_id = getattr(args, 'instance_id', None)
            asyncio.run(cli.stop_command(instance_id))

        elif args.command == "create":
            asyncio.run(cli.create_project(args.project_dir))

        else:
            parser.print_help()

    except KeyboardInterrupt:
        print("\n👋 Goodbye!")
    except Exception as e:
        print(f"💥 Error: {e}")


if __name__ == "__main__":
    main()