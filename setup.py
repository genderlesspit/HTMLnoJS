"""
Setup configuration for HTMLnoJS package
"""
from setuptools import setup, find_packages
from pathlib import Path

# Read README
this_directory = Path(__file__).parent
long_description = (this_directory / "README.md").read_text() if (this_directory / "README.md").exists() else ""

setup(
    name="htmlnojs",
    version="0.1.0",
    author="HTMLnoJS Team",
    author_email="team@htmlnojs.dev",
    description="Async Python orchestrator for HTML-first web applications",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/htmlnojs/python-orchestrator",

    packages=find_packages(),
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Internet :: WWW/HTTP :: HTTP Servers",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: Internet :: WWW/HTTP :: Dynamic Content",
    ],

    python_requires=">=3.8",
    install_requires=[
        "fastapi>=0.68.0",
        "uvicorn[standard]>=0.15.0",
        "aiohttp>=3.8.0",
    ],

    extras_require={
        "dev": [
            "pytest>=6.0",
            "pytest-asyncio>=0.18.0",
            "black>=21.0.0",
            "flake8>=3.9.0",
            "mypy>=0.910",
        ],
        "docs": [
            "mkdocs>=1.2.0",
            "mkdocs-material>=7.0.0",
        ],
    },

    entry_points={
        "console_scripts": [
            "htmlnojs=htmlnojs.cli:main",
        ],
    },

    project_urls={
        "Bug Reports": "https://github.com/htmlnojs/python-orchestrator/issues",
        "Source": "https://github.com/htmlnojs/python-orchestrator",
        "Documentation": "https://htmlnojs.dev/docs",
    },
)