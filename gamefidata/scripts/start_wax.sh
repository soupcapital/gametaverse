#!/usr/bin/env bash
ps aux |grep gamefidata |grep wax |awk '{print $2}' | xargs kill -9
nohup ../gamefidata wax -c ./config_wax.toml > wax.log &
