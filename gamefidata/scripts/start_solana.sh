#!/usr/bin/env bash
ps aux |grep gamefidata |grep solana |awk '{print $2}' | xargs kill -9
nohup ../gamefidata solana -c ./config_solana.toml > solana.log &
