#!/bin/bash

echo "🧪 Testing Sweepstakes Restructured Build..."
echo ""

# Check if files exist
echo "✅ Checking files..."
files=(
    "sweepstakes-main.go"
    "sweepstakes-handlers.go"
    "sweepstakes-models.go"
    "sweepstakes-database.go"
    "sweepstakes-auth.go"
    "go.mod"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✓ $file"
    else
        echo "  ✗ $file (MISSING)"
        exit 1
    fi
done

echo ""
echo "📦 Running go mod tidy..."
go mod tidy

echo ""
echo "🔨 Building application..."
go build -o sweepstakes *.go

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Build successful!"
    echo ""
    echo "File sizes:"
    ls -lh sweepstakes-*.go | awk '{print "  " $9 ": " $5}'
    echo ""
    echo "Total lines of code:"
    wc -l sweepstakes-*.go | tail -1
    echo ""
    echo "🚀 Ready to run: ./sweepstakes"
else
    echo ""
    echo "❌ Build failed!"
    exit 1
fi
