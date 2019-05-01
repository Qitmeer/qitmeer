#!/bin/bash

app_dir=/work/zenprotocol/zenprotocol/src/Node/bin/Debug/
cd $app_dir

port=5555
ip=127.0.0.1
api=$ip:$port
data=/data/zen/local

mono zen-node.exe --chain local --api $api --data-path $data "$@"

cd -
