#!/usr/bin/env bash
ps aux |grep gamefidata |grep "config_wax" |awk '{print $2}' | xargs kill -9
nohup ../gfdp spider -c ./config_wax.toml > /dev/null 2>&1 &
