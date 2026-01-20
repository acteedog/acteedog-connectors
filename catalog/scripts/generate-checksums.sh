#!/bin/bash
set -e

# Usage: ./generate-checksums.sh <connector-dir>
# Example: ./generate-checksums.sh connectors/github/1.0.0

if [ -z "$1" ]; then
  echo "Usage: $0 <plugin-dir>"
  echo "Example: $0 connectors/github/1.0.0"
  exit 1
fi

CONNECTOR_DIR="$1"

if [ ! -d "$CONNECTOR_DIR" ]; then
  echo "Error: Directory $CONNECTOR_DIR does not exist"
  exit 1
fi

if [ ! -f "$CONNECTOR_DIR/plugin.wasm" ]; then
  echo "Error: plugin.wasm not found in $CONNECTOR_DIR"
  exit 1
fi

echo "Generating checksums for $CONNECTOR_DIR..."

# Generate SHA256 checksum
CHECKSUM=$(sha256sum "$CONNECTOR_DIR/plugin.wasm" | awk '{print $1}')

# Get file size
SIZE=$(stat -c%s "$CONNECTOR_DIR/plugin.wasm")

# Write checksums file
cat > "$CONNECTOR_DIR/checksums.txt" <<EOF
SHA256: sha256:$CHECKSUM
Size: $SIZE bytes
EOF

echo "✓ Checksums written to $CONNECTOR_DIR/checksums.txt"
echo "  SHA256: sha256:$CHECKSUM"
echo "  Size: $SIZE bytes"
