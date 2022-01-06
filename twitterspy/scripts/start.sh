#!/usr/bin/env bash
nohup ../twitterspy spider -c config.toml &
nohup ../twitterspy api -c config_api.toml   &
nohup ../twitterspy digger -c config_digger.toml  &