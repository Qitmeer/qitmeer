#!/usr/bin/env bash

network="--privnet"
mining="--miningaddr RmFUBXUpN3W6bSgaBHpPQzQ7pXNb3fG59Kn"
debug="-d trace --printorigin"
rpc="--listen 127.0.0.1:1234 --rpcuser test --rpcpass test"
path="-b ."
index="--txindex"

#1.) The start script used only for dev test purpose
#2.) the relative path of nox executable link sould be the same path of the start.sh link 
#3.) the "-b ." set the base data directory as the same place where the start script executed
#
# EX:
# $ WORK=/tmp/my_test_workspace
# $ mkdir -p $WORK
# $ cd $WORK
# $ ls -s /where/is/my/nox/executable
# $ ln -s /where/is/my/nox/start.sh
# $ ./start.sh
./nox $debug $rpc $path $index $network $mining "$@"


