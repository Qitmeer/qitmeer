#!/bin/bash

# set -e

# ---------------------------
# settings
# ---------------------------
DEBUG=0
DEBUG_FILE=/tmp/mydcr_curl_debug
ERR_FILE=/tmp/mydcr_curl_error


data_dir="/data/dcr/private/" 

function get_result(){
  set +x
  if [ -z "$host" ]; then
     host="explorer.dcrdata.org"
  fi
  if [ -z "$port" ]; then
     port=""
  fi
  local method=$1
  local params=$2
  local verbose=$3
  # curl -s https://explorer.dcrdata.org/api/tx/decoded/5442e9dd4961ffefad37cb64b8c906574937b1c74850167c1af9d8587ed504ce|jq .

  local curl_result=$(curl -s "https://$host$port/api/$method/$params/$verbose")

  local result=$(echo $curl_result|jq -r -c -M '.')
  if [ $DEBUG -gt 0 ]; then
    local curl_cmd="curl -s https://$host:$port/api/$method/$params/$verbose"
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
  shift
  get_result tx $@ |jq .
elif [ "$1" == "block" ]; then #amdin block
  shift
  get_result block $@|jq .
elif [ "$1" == "api" ]; then
  shift
  get_result "$@" |jq .
else
  $cli "$@"
fi
check_debug
