#!/usr/bin/env bash
ps aux |grep gamefidata |grep "config_api" |awk '{print $2}' | xargs kill -9
nohup ../gamefidata api -c ./config_api.toml > api.out &
