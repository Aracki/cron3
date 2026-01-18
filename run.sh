#!/bin/bash

cd "$(dirname "$0")"

go build -o cron3 . || exit 1

./cron3 2>> log &

echo "cron3 started with PID $!"
