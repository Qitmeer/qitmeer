#!/bin/bash

# set -e

# ---------------------------
# settings
# ---------------------------
DEBUG=0
DEBUG_FILE=/tmp/myfactom_curl_debug
ERR_FILE=/tmp/myfactom_curl_error


data_dir="/data/factom/private/" 

cli="./factom-cli" 

function get_result(){
  set +x
  if [ -z "$host" ]; then
     host=127.0.0.1
  fi
  if [ -z "$port" ]; then
     port=8088
  fi
  local method=$1
  local params=$2
  local data='{"jsonrpc": "2.0", "id": 0, "method": "'$method'", "params": {'$params'}}'
  local curl_result=$(curl -s -X POST -H 'Content-Type: application/json' --data "$data" "http://$host:$port/v2")

  local result=$(echo $curl_result|jq -r -c -M '.')
  if [ $DEBUG -gt 0 ]; then
    local curl_cmd="curl -s -X POST -H 'Content-Type: application/json' --data '$data' http://$host:$port/v2"
    echo "$curl_cmd" > $DEBUG_FILE                                                                                        
    echo "$curl_result" >> $DEBUG_FILE
  fi
  if [ "$result" == "null" ];then
    echo "$curl_result" > $ERR_FILE
    result=""
  fi
  echo $result
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
    *)
      cmd="$@"
      #echo "cmd is $cmd"
      break;;
    esac
  shift
done

if [ "$1" == "tx" ]; then
  get_result tx
elif [ "$1" == "ablock" ]; then #amdin block
  shift
  get_result ablock-by-height '"height":'$1''|jq .
elif [ "$1" == "ecblock" ]; then #entry credit block
  shift
  get_result ecblock-by-height '"height":'$1''|jq .
elif [ "$1" == "dblock" ]; then #directory block
  shift
  get_result dblock-by-height '"height":'$1''|jq .
elif [ "$1" == "fblock" ]; then #factoid block
  shift
  get_result fblock-by-height '"height":'$1''|jq .
elif [ "$1" == "api" ]; then
  shift
  get_result "$@" |jq .
else
  $cli "$@"
fi
check_debug
