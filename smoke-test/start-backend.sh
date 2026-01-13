#!/bin/bash
# Quick start script for smoke-test with CORS config

cd ~/pubgames-v2/smoke-test || exit 1

echo "🚀 Starting Smoke Test Backend (Port 30011)..."
echo "📋 Using shared CORS config from: ~/pubgames-v2/shared/config/cors-config.json"
echo ""

go run .
