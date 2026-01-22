#!/bin/bash
# Script to build Lambda layer with dependencies

set -e

LAYER_DIR="$(dirname "$0")/layer"
PYTHON_DIR="$LAYER_DIR/python"

echo "Creating Lambda layer directory..."
rm -rf "$LAYER_DIR"
mkdir -p "$PYTHON_DIR"

echo "Installing batch dependencies..."
python3 -m pip install -r "$(dirname "$0")/batch/requirements.txt" \
    -t "$PYTHON_DIR" \
    --platform manylinux2014_x86_64 \
    --target "$PYTHON_DIR" \
    --implementation cp \
    --python-version 3.12 \
    --only-binary=:all: \
    --upgrade \
    --quiet

echo "Cleaning up unnecessary files..."
find "$PYTHON_DIR" -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
find "$PYTHON_DIR" -type d -name "*.dist-info" -exec rm -rf {} + 2>/dev/null || true
find "$PYTHON_DIR" -type d -name "tests" -exec rm -rf {} + 2>/dev/null || true
find "$PYTHON_DIR" -type f -name "*.pyc" -delete 2>/dev/null || true

echo "Lambda layer created successfully!"
echo "Layer directory: $LAYER_DIR"
