#!/usr/bin/env bash
./stop.sh
rm *.log *.out
nohup ./twitterspy -c config.toml &
