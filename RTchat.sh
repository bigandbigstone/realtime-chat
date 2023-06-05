#!/bin/bash


DIR=/home/rtos/measure/realtime-chat
cd $DIR

export DEVICE=Dida
export PYTHONPATH=$(dirname ${DIR})

sudo kill -9 `ps -aux | grep realtime | awk '{print $2}'`

nohup python3 ${DIR}/realtime-chat-client.py &
sleep 3s
sudo ./realtime-chat