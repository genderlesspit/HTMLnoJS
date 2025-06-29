"""
Instance Registry - Manages HTMLnoJS application instances
"""
from typing import Dict, Optional, List, TYPE_CHECKING

if TYPE_CHECKING:
    from .core import HTMLnoJS


class InstanceRegistry:
    """Registry for managing HTMLnoJS instances"""

    _instances: Dict[str, 'HTMLnoJS'] = {}

    @classmethod
    def register(cls, instance: 'HTMLnoJS') -> None:
        """Register an instance"""
        cls._instances[instance.id] = instance

    @classmethod
    def unregister(cls, instance_id: str) -> None:
        """Unregister an instance"""
        cls._instances.pop(instance_id, None)

    @classmethod
    def get(cls, instance_id: str) -> Optional['HTMLnoJS']:
        """Get instance by ID"""
        return cls._instances.get(instance_id)

    @classmethod
    def list_all(cls) -> List['HTMLnoJS']:
        """List all registered instances"""
        return list(cls._instances.values())

    @classmethod
    def list_running(cls) -> List['HTMLnoJS']:
        """List all running instances"""
        return [instance for instance in cls._instances.values() if instance.is_running]

    @classmethod
    def find_by_port(cls, port: int) -> Optional['HTMLnoJS']:
        """Find instance by Go server port"""
        for instance in cls._instances.values():
            if instance.go_port == port:
                return instance
        return None

    @classmethod
    def find_by_project(cls, project_dir: str) -> Optional['HTMLnoJS']:
        """Find instance by project directory"""
        from pathlib import Path
        target_dir = Path(project_dir).resolve()

        for instance in cls._instances.values():
            if instance.project_dir == target_dir:
                return instance
        return None

    @classmethod
    async def stop_all(cls) -> None:
        """Stop all registered instances"""
        instances = list(cls._instances.values())
        for instance in instances:
            await instance.stop()

    @classmethod
    def clear(cls) -> None:
        """Clear the registry"""
        cls._instances.clear()

    @classmethod
    def get_status_summary(cls) -> dict:
        """Get summary status of all instances"""
        instances = list(cls._instances.values())
        return {
            "total_instances": len(instances),
            "running_instances": len([i for i in instances if i.is_running]),
            "instances": [
                {
                    "id": instance.id,
                    "project_dir": str(instance.project_dir),
                    "ports": {"go": instance.go_port, "python": instance.python_port},
                    "running": instance.is_running
                }
                for instance in instances
            ]
        }