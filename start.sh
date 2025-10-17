#!/bin/bash

# 配置环境变量
export DALINK_GO_CONFIG_PATH=/data/chaos/api-server/prod.yml

# 进程名称
PROCESS_NAME="./chaos-api"
LOG_FILE="output.log"

# 检查进程是否存在
PID=$(pgrep -f $PROCESS_NAME)

if [ -n "$PID" ]; then
  echo "Killing existing process $PROCESS_NAME with PID $PID"
  kill -9 $PID
  sleep 2
else
  echo "No existing process found for $PROCESS_NAME"
fi

# 启动新的进程
echo "Starting new process $PROCESS_NAME"
nohup ./$PROCESS_NAME > $LOG_FILE 2>&1 &

# 获取新的进程ID
NEW_PID=$(pgrep -f $PROCESS_NAME)
if [ -n "$NEW_PID" ]; then
  echo "New process $PROCESS_NAME started with PID $NEW_PID"
else
  echo "Failed to start new process $PROCESS_NAME"
fi
