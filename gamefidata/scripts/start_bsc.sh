#!/usr/bin/env bash
ps aux |grep gamefidata |grep "config_bsc" |awk '{print $2}' | xargs kill -9
nohup ../gamefidata eth -c ./config_bsc.toml > bsc.log &
