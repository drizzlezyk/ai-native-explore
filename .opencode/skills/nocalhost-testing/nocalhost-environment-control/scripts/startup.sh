#!/bin/bash
set -e

if [ -d "/vault/secrets" ] && [ "$(ls -A /vault/secrets 2>/dev/null)" ]; then
    echo "Backing up vault secrets..."
    mkdir -p /vault/backup 2>/dev/null || true
    cp -r /vault/secrets/* /vault/backup/ 2>/dev/null || true
else
    if [ -d "/vault/backup" ] && [ "$(ls -A /vault/backup 2>/dev/null)" ]; then
        echo "Restoring vault secrets from backup..."
        mkdir -p /vault/secrets 2>/dev/null || true
        cp -r /vault/backup/* /vault/secrets/ 2>/dev/null || true
    fi
fi

echo "Starting BigFiles server..."
export HOME=/home/nocalhost-dev
cd /home/nocalhost-dev
./main --enable_debug --config-file=/vault/secrets/config.yaml
