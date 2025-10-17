#!/bin/sh
set -e

REMOTE_HOST="ngamefi"
REMOTE_DIR="/data/chaos/api-server"
PROJECT_DIR="/data/chaos/source/api"

echo "start building......"
ssh "${REMOTE_HOST}" "cd ${REMOTE_DIR} && sudo ./build.sh"

echo "restart application......"
ssh "${REMOTE_HOST}" "cd ${REMOTE_DIR} && sudo ./start.sh"

echo "finished"