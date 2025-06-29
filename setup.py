from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="pygops",
    version="0.1.0",
    description="Python wrapper for Go applications with PowerShell launcher",
    long_description=long_description,
    long_description_content_type="text/markdown",
    author="PyGoPS Team",
    packages=find_packages(),
    include_package_data=True,  # This tells setuptools to use MANIFEST.in
    install_requires=[
        "loguru>=0.6.0",
        "aiohttp>=3.8.0",
    ],
    python_requires=">=3.8",
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: Microsoft :: Windows",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: System :: Systems Administration",
    ],
    keywords="go golang powershell launcher wrapper server",
)