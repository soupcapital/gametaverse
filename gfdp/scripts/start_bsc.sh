#!/usr/bin/env bash
ps aux |grep gfdp|grep "config_bsc" |awk '{print $2}' | xargs kill -9
nohup ../gfdp spider -c ./config_bsc.toml > /dev/null 2>&1 &
