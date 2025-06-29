"""
Port Manager - Handles port allocation and availability with process management
"""
import socket
import subprocess
import platform
import psutil
from typing import List, Tuple, Optional, Dict
from loguru import logger as log


class PortManager:
    """Manages port allocation and process management for HTMLnoJS services"""

    @staticmethod
    def find_available_port(start_port: int = 8080, count: int = 1) -> int:
        """Find a single available port"""
        ports = PortManager.find_available_ports(start_port, count)
        return ports[0] if ports else None

    @staticmethod
    def find_available_ports(start_port: int = 8080, count: int = 2) -> List[int]:
        """Find multiple available ports"""
        available_ports = []
        current_port = start_port

        while len(available_ports) < count and current_port < start_port + 1000:
            if PortManager.is_port_available(current_port):
                available_ports.append(current_port)
            current_port += 1

        return available_ports

    @staticmethod
    def is_port_available(port: int) -> bool:
        """Check if a port is available"""
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
                sock.bind(('', port))
                return True
        except OSError:
            return False

    @staticmethod
    def allocate_port_pair(start_port: int = 3000) -> Tuple[int, int]:
        """Allocate a pair of consecutive ports, killing processes if needed"""
        # First try to find available ports
        ports = PortManager.find_available_ports(start_port, 2)
        if len(ports) >= 2:
            return ports[0], ports[1]

        # If not available, kill processes on desired ports and retry
        log.warning(f"Ports {start_port} and {start_port + 1} not available, attempting to free them...")
        PortManager.kill_process_on_port(start_port)
        PortManager.kill_process_on_port(start_port + 1)

        # Wait a moment and try again
        import time
        time.sleep(1)

        if PortManager.is_port_available(start_port) and PortManager.is_port_available(start_port + 1):
            return start_port, start_port + 1

        # Fallback to finding any available ports
        ports = PortManager.find_available_ports(start_port + 2, 2)
        if len(ports) < 2:
            raise RuntimeError("Could not find two available ports")
        return ports[0], ports[1]

    @staticmethod
    def get_process_using_port(port: int) -> Optional[Dict]:
        """Get information about the process using a specific port"""
        try:
            for conn in psutil.net_connections(kind='inet'):
                if conn.laddr.port == port and conn.status == psutil.CONN_LISTEN:
                    try:
                        process = psutil.Process(conn.pid)
                        return {
                            'pid': conn.pid,
                            'name': process.name(),
                            'cmdline': ' '.join(process.cmdline()),
                            'status': conn.status
                        }
                    except (psutil.NoSuchProcess, psutil.AccessDenied):
                        return {'pid': conn.pid, 'name': 'Unknown', 'cmdline': 'Unknown', 'status': conn.status}
        except Exception as e:
            log.error(f"Error getting process info for port {port}: {e}")
        return None

    @staticmethod
    def kill_process_on_port(port: int, force: bool = True) -> bool:
        """Kill the process using a specific port"""
        process_info = PortManager.get_process_using_port(port)
        if not process_info:
            log.debug(f"No process found using port {port}")
            return True

        pid = process_info['pid']
        name = process_info['name']

        log.info(f"Killing process {name} (PID: {pid}) using port {port}")

        try:
            if platform.system() == "Windows":
                # Windows approach
                cmd = ['taskkill', '/PID', str(pid)]
                if force:
                    cmd.append('/F')
                result = subprocess.run(cmd, capture_output=True, text=True)
                success = result.returncode == 0
            else:
                # Unix/Linux/Mac approach
                import signal
                process = psutil.Process(pid)
                if force:
                    process.kill()  # SIGKILL
                else:
                    process.terminate()  # SIGTERM
                process.wait(timeout=5)
                success = True

        except subprocess.CalledProcessError as e:
            log.error(f"Failed to kill process {pid}: {e}")
            success = False
        except psutil.NoSuchProcess:
            log.debug(f"Process {pid} already terminated")
            success = True
        except psutil.TimeoutExpired:
            log.warning(f"Process {pid} didn't terminate within timeout")
            success = False
        except Exception as e:
            log.error(f"Unexpected error killing process {pid}: {e}")
            success = False

        if success:
            log.success(f"Successfully killed process {name} (PID: {pid}) on port {port}")
        else:
            log.error(f"Failed to kill process {name} (PID: {pid}) on port {port}")

        return success

    @staticmethod
    def kill_processes_on_ports(*ports: int) -> Dict[int, bool]:
        """Kill processes on multiple ports"""
        results = {}
        for port in ports:
            results[port] = PortManager.kill_process_on_port(port)
        return results

    @staticmethod
    def force_free_port_pair(start_port: int = 3000) -> Tuple[int, int]:
        """Force free a port pair by killing existing processes"""
        target_ports = [start_port, start_port + 1]

        log.info(f"Force freeing ports {target_ports}")
        results = PortManager.kill_processes_on_ports(*target_ports)

        # Wait for ports to be freed
        import time
        time.sleep(2)

        # Verify ports are now available
        if all(PortManager.is_port_available(port) for port in target_ports):
            log.success(f"Successfully freed ports {target_ports}")
            return tuple(target_ports)
        else:
            log.warning(f"Some ports still not available, finding alternatives...")
            return PortManager.allocate_port_pair(start_port + 2)

    @staticmethod
    def validate_ports(*ports: int) -> bool:
        """Validate that all ports are available"""
        return all(PortManager.is_port_available(port) for port in ports)

    @staticmethod
    def list_port_usage(port_range: Tuple[int, int] = (3000, 3010)) -> Dict[int, Optional[Dict]]:
        """List processes using ports in a range"""
        start_port, end_port = port_range
        usage = {}

        for port in range(start_port, end_port + 1):
            process_info = PortManager.get_process_using_port(port)
            usage[port] = process_info

        return usage

    @staticmethod
    def cleanup_htmlnojs_processes():
        """Clean up all HTMLnoJS-related processes"""
        log.info("Cleaning up HTMLnoJS processes...")

        cleaned = []
        for proc in psutil.process_iter(['pid', 'name', 'cmdline']):
            try:
                cmdline = ' '.join(proc.info['cmdline'] or [])
                name = proc.info['name'] or ''

                # Look for processes that might be HTMLnoJS related
                if any(keyword in cmdline.lower() or keyword in name.lower() for keyword in
                       ['htmlnojs', 'uvicorn', 'main.go', 'htmx_server']):
                    log.info(f"Killing HTMLnoJS process: {name} (PID: {proc.info['pid']})")
                    proc.kill()
                    cleaned.append(proc.info['pid'])

            except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess):
                pass

        if cleaned:
            log.success(f"Cleaned up {len(cleaned)} HTMLnoJS processes: {cleaned}")
        else:
            log.info("No HTMLnoJS processes found to clean up")

        return cleaned