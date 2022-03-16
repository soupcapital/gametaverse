#!/usr/bin/env bash
protoc --go_out=.  --go-grpc_out=.  --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative  proxy.proto 
