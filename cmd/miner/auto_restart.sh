#!/bin/bash

APP_NAME=qitmeer-miner
while true
do
    procnum=`ps -ec|grep "$APP_NAME"|grep -v grep|wc -l`
   if [ $procnum -eq 0 ]; then
        echo "./build/bin/$APP_NAME -C pool.conf >> $APP_NAME.log 2>&1 &"
       ./build/bin/$APP_NAME -C pool.conf >> $APP_NAME.log 2>&1 &
       #./build/bin/$APP_NAME -C solo.conf >> $APP_NAME.log 2>&1 &
   fi
   sleep 1
done
