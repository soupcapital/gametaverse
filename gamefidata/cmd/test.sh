#!/usr/bin/env bash

#curl "https://www.4byte.directory/api/v1/signatures/?hex_signature=0x0c53c51c"
#./gamefidata eth -c ../scripts/config_polygon.toml 
#./gamefidata eth -c ../scripts/config_eth.toml 
#./gamefidata eth -c ../scripts/config_bsc.toml 
#./gamefidata wax -c ../scripts/config_wax.toml 
#./gamefidata solana -c ../scripts/config_solana.toml 
#./gamefidata api -c ../scripts/config_api.toml 
#./gamefidata eth -c ../scripts/config_avax.toml 
./gamefidata daily -m "mongodb+srv://cz:Solong2020@cluster1.g9w77.mongodb.net/myFirstDatabase?retryWrites=true&w=majority" -r "172.31.6.11:8081" -d "2022-04-07" 
