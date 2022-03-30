#!/usr/bin/env bash
ps aux |grep gfdp |grep "config_polygon" |awk '{print $2}' | xargs kill -9
nohup ../gfdp spider -c ./config_polygon.toml > /dev/null 2>&1 &
