#!/usr/bin/env bash
./stop.sh
rm *.log *.out
nohup ./cti -c config/config.toml &
