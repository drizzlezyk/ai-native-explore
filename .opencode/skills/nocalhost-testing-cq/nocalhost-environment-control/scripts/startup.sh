#!/bin/bash
set -e

# 1. Setup templates (Required for xihe-server to start)
echo "Setting up templates..."
mkdir -p /opt/app/points/task-docs-templates
cp ./points/infrastructure/taskdocimpl/doc_chinese.tmpl /opt/app/points/task-docs-templates/
cp ./points/infrastructure/taskdocimpl/doc_english.tmpl /opt/app/points/task-docs-templates/

# 2. Setup secrets backup/restore
# Nocalhost duplicate mode can sometimes miss mounted secrets if the sidecar is not correctly duplicated.
if [ -d "/vault/secrets" ] && [ "$(ls -A /vault/secrets 2>/dev/null)" ]; then
    echo "Backing up vault secrets..."
    mkdir -p /vault/backup
    cp -r /vault/secrets/* /vault/backup/ 2>/dev/null || true
else
    if [ -d "/vault/backup" ] && [ "$(ls -A /vault/backup 2>/dev/null)" ]; then
        echo "Restoring vault secrets from backup..."
        mkdir -p /vault/secrets
        cp -r /vault/backup/* /vault/secrets/ 2>/dev/null || true
    fi
fi

# Ensure /vault/secrets/application.yml exists (even as a touch) if we are in a tight spot
if [ ! -f "/vault/secrets/application.yml" ]; then
    echo "Warning: /vault/secrets/application.yml not found. Attempting to create empty file."
    mkdir -p /vault/secrets
    touch /vault/secrets/application.yml
fi

# 3. Start server
echo "Starting xihe-server with XIHE_USERNAME=${XIHE_USERNAME}..."
export HOME=/home/nocalhost-dev
./xihe-server --port 8000 --config-file /vault/secrets/application.yml --enable_debug
