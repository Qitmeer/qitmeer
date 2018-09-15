#!/usr/bin/env bash

network="--privnet"
mining="--miningaddr RmFUBXUpN3W6bSgaBHpPQzQ7pXNb3fG59Kn"
debug="-d trace --printorigin"
rpc="--listen 127.0.0.1:1234 --rpcuser test --rpcpass test"
path="-b ."
index="--txindex"

../../nox $debug $rpc $path $index $network $mining "$@"


