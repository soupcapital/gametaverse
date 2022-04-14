#!/usr/bin/env bash

count=1
while [ $count -lt 2 ]
do
	echo $count
 	d=`date +%Y-%m-%d -d "-$count day"`
	echo $d
	./cmd/gamefidata daily -m "mongodb+srv://cz:Solong2020@cluster1.g9w77.mongodb.net/myFirstDatabase?retryWrites=true&w=majority" -r "172.31.6.11:8081" -d $d
	let count++
	sleep 1
done
