#!/bin/bash

echo "=========================================="
echo "      🎬 Starting WATCHME v2.0 (Mac)"
echo "=========================================="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "[ERROR] Docker is not installed."
    echo "Please install Docker Desktop for Mac:"
    echo "https://docs.docker.com/desktop/install/mac-install/"
    echo ""
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    echo "[ERROR] Docker daemon is not running."
    echo "Please open the 'Docker' application from your Applications folder and wait for it to start."
    echo ""
    exit 1
fi

echo "[INFO] Starting containers..."
# Support both `docker compose` and `docker-compose`
if docker compose version &> /dev/null; then
    docker compose up -d
elif docker-compose version &> /dev/null; then
    docker-compose up -d
else
    echo "[ERROR] Docker Compose plugin is not installed."
    echo "Please install it or ensure Docker Desktop is fully updated."
    echo ""
    exit 1
fi

if [ $? -ne 0 ]; then
    echo ""
    echo "[ERROR] Failed to start Docker containers."
    echo ""
    exit 1
fi

echo ""
echo "=========================================="
echo "  ✅ WATCHME is now running!"
echo "=========================================="
echo ""
echo "  🌐 Open your browser and go to:"
echo "     http://localhost:3000"
echo ""
echo "  To view logs, run:"
echo "     docker compose logs -f"
echo ""
echo "  To stop the application, run:"
echo "     docker compose down"
echo ""
