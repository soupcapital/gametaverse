#!/usr/bin/env bash
ps aux |grep gfdp|grep "config_bsc" |awk '{print $2}' | xargs kill -9
nohup ../gfdp eth -c ./config_bsc.toml > bsc.log &
