#!/bin/bash
count=$3
if [ -z $3 ]; then
  count=1
fi
for ((block=$1;block<=$2;block++)); do
    txcount=$(./btc.sh block $block txcount); 
    if [ $txcount -gt $count ]; then 
      echo $block $txcount
    fi 
done
