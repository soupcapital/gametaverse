#!/usr/bin/env bash
ps aux |grep gamefidata |grep eth |awk '{print $2}' | xargs kill -9
nohup ../gamefidata eth -c ./config_eth.toml > eth.log &
