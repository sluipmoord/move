#!/bin/bash
# Build script for Move Reminder
# Fixes the duplicate Objective-C library warning on macOS

echo "Building Move Reminder..."
CGO_LDFLAGS="-Wl,-no_warn_duplicate_libraries" go build -o main main.go
echo "Build complete! Run with: ./main"