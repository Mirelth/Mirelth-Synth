#!/bin/sh
set -e

echo "=========================================="
echo "  Synth — TypeScript Demo"
echo "=========================================="
echo ""

cd /opt/synth-sdk

echo ">>> Running TypeScript SDK demo"
echo "------------------------------------------"
npx tsx /opt/synth-demo/ts-basic.ts
echo ""

echo "=========================================="
echo "  TypeScript demo complete!"
echo "=========================================="
