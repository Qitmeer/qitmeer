#!/bin/bash
 
while true
do 
    procnum=`ps -ef|grep "miner"|grep -v grep|wc -l`
   if [ $procnum -eq 0 ]; then
        echo "./linux-miner -C solo.conf >>/tmp/miner.log 2>&1 &"
       ./qitmeer-miner -C pool.conf >> miner.log 2>&1 &
   fi
   sleep 1
done