#!/bin/bash

# set -e

# ---------------------------
# settings
# ---------------------------
DEBUG=0
DEBUG_FILE=/tmp/myneo_curl_debug
ERR_FILE=/tmp/myneo_curl_error


data_dir="/data/neo/private/" 

cli=""


function get_result(){
  set +x
  if [ -z "$host" ]; then
     host="neoscan.io"
  fi
  if [ -z "$port" ]; then
     port=""
  fi
  if [ -z "$network" ]; then
     network="main_net"
  fi
  local method=$1
  local params=$2

  # curl -s https://neoscan.io/api/main_net/v1/get_highest_block|jq .

  local curl_result=$(curl -s "https://$host$port/api/$network/v1/$method/$params$verbose")

  local result=$(echo $curl_result)
  if [ $DEBUG -gt 0 ]; then
    local curl_cmd="curl -s https://$host$port/api/$network/v1/$method/$params$verbose"
    echo "$curl_cmd" > $DEBUG_FILE                                                                                        
    echo "$curl_result" >> $DEBUG_FILE
  fi
  if [ "$result" == "null" ];then
    echo "$curl_result" > $ERR_FILE
    result=""
  fi
  if [ "$decode" == "hex" ]; then
    echo $result
  else
    echo $result |jq .
  fi
}

function check_error() {
  if [ -e $ERR_FILE ]; then
    echo "Error:"
    cat $ERR_FILE|jq .error
    rm $ERR_FILE
    if [ "$1" == "-e" ]; then
      check_debug
      exit -1
    fi
  fi
}
function check_debug() {
  if [ $DEBUG -gt 0 ]; then
    if [ -e $DEBUG_FILE ]; then
      cat $DEBUG_FILE
      rm $DEBUG_FILE
    fi
  fi
}

# main logic
if [ $? != 0 ]; then
  echo "Usage: -h [host] -p [port] "
  exit;
fi
#echo "$@"
while [ $# -gt 0 ] ;do
  case "$1" in
    -h)
      host=$2
      #echo "host is $host"
      shift;;
    -p)
      port=$2
      #echo "port is $port"
      shift;;
    -D)
      DEBUG=1
      ;;
    -n|--network)
      network=$2
      shift;;
    *)
      cmd="$@"
      #echo "cmd is $cmd"
      break;;
    esac
  shift
done

if [ "$1" == "tx" ]; then
  shift
  echo $cli $@
elif [ "$1" == "block" ]; then
  shift
  echo $cli $@
elif [ "$1" == "api" ]; then
  shift
  get_result "$@" 
else
  echo $cli "$@"
fi
      

check_debug
