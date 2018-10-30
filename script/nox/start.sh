#!/usr/bin/env bash

# 
# echo 7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b|nx ec-new |nx ec-to-public |nx ec-to-addr
# RmFa5hnPd3uQRpzr3xWTfr8EFZdX7dS1qzV
# echo 7025927350b0f968c4a012df2b30cc494786cfff55b177d199069d9bc5aa4035|nx ec-new |nx ec-to-public |nx ec-to-addr
# RmG6xQsV7gnS4JZmoq5FgmyEbmUQRenrTCo
network="--privnet"
mining="--miningaddr RmM4oveyHptJMkHRf396f6QawErf7Min6yU"
debug="-d trace --printorigin"
rpc="--listen 127.0.0.1:1234 --rpcuser test --rpcpass test"
path="-b "$(pwd)
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


