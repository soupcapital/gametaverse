#!/usr/bin/env bash

d=`date +"%Y-%m-%d"`
echo $d
/home/ubuntu/twitterspy/dist_twitterspy_v0.1.7/linux/twitterspy digger -c /home/ubuntu/twitterspy/dist_twitterspy_v0.1.7/linux/scripts/config_digger.toml -s -d $d >/tmp/score.log
