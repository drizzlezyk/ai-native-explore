#!/bin/bash
set -e

export HOME=/home/nocalhost-dev
export GOCACHE=/home/nocalhost-dev/.cache/go-build
mkdir -p /home/nocalhost-dev/.cache
cd /home/nocalhost-dev
go build --buildvcs=false -mod=vendor -o xihe-server ./main.go
