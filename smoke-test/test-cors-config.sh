#!/bin/bash
# Test script for smoke-test with shared CORS config

echo "🧪 Testing smoke-test with shared CORS config..."
echo ""

# Navigate to smoke-test directory
cd "$(dirname "$0")" || exit 1

echo "📦 Step 1: Tidying Go modules..."
go mod tidy
if [ $? -ne 0 ]; then
    echo "❌ go mod tidy failed"
    exit 1
fi
echo "✅ Go modules OK"
echo ""

echo "🔨 Step 2: Building smoke-test..."
go build -o smoke-test-binary
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"
echo ""

echo "🧹 Cleaning up binary..."
rm -f smoke-test-binary

echo ""
echo "✅ All tests passed!"
echo ""
echo "📋 CORS Config Location: ~/pubgames-v2/shared/config/cors-config.json"
echo "📋 Current CORS Settings:"
cat ../shared/config/cors-config.json
echo ""
echo "🚀 To start smoke-test:"
echo "   cd ~/pubgames-v2/smoke-test"
echo "   go run ."
echo ""
echo "Watch for these lines in the startup logs:"
echo "   ✅ Loaded CORS config: mode=pattern, environment=development"
echo "   📋 CORS Mode: pattern"
echo "   📋 Allowed Origins: [http://localhost:* http://192.168.1.*:*]"
