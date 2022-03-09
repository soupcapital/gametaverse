#!/usr/bin/env bash
ps aux |grep gfdp |grep "config_polygon" |awk '{print $2}' | xargs kill -9
nohup ../gfdp eth -c ./config_polygon.toml > polygon.log &
