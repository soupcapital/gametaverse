#!/usr/bin/env bash
./stop.sh
rm *.log *.out
nohup ../twitterspy spider -c config.toml > spider.out &
nohup ../twitterspy api -c config_api.toml  > api.out &