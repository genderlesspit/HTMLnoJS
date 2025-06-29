"""
Port Manager - Handles port allocation and availability
"""
import socket
from typing import List, Tuple


class PortManager:
    """Manages port allocation for HTMLnoJS services"""

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
    def allocate_port_pair(start_port: int = 8080) -> Tuple[int, int]:
        """Allocate a pair of consecutive ports"""
        ports = PortManager.find_available_ports(start_port, 2)
        if len(ports) < 2:
            raise RuntimeError("Could not find two available ports")
        return ports[0], ports[1]

    @staticmethod
    def validate_ports(*ports: int) -> bool:
        """Validate that all ports are available"""
        return all(PortManager.is_port_available(port) for port in ports)