#!/bin/bash

# set -e

# ---------------------------
# settings
# ---------------------------
DEBUG=0
DEBUG_FILE=/tmp/mybtc_curl_debug
ERR_FILE=/tmp/mybtc_curl_error

function get_result(){
  set +x
  if [ -z "$host" ]; then
     host="blockexplorer.com"
  fi
  if [ -z "$port" ]; then
     port=""
  fi
  local method=$1
  local params=$2

  # curl -s https://blockexplorer.com/api/rawtx/5756ff16e2b9f881cd15b8a7e478b4899965f87f553b6210d0f8e5bf5be7df1d|jq -r .rawtx

  local curl_result=$(curl -s "https://$host$port/api/$method/$params$verbose")

  local result=$(echo $curl_result)
  if [ $DEBUG -gt 0 ]; then
    local curl_cmd="curl -s https://$host$port/api/$method/$params$verbose"
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
    -p|-port|--port)
      port=$2
      #echo "port is $port"
      shift;;
    -D)
      DEBUG=1
      ;;
    -n|-network|--network)
      network=$2
      shift;;
    -datadir|--datadir)
      datadir=$2
      shift;;
    *)
      cmd="$@"
      #echo "cmd is $cmd"
      break;;
    esac
  shift
done

if [ -z "$datadir" ]; then
  datadir="/data/bitcoin/private/" 
fi

case "$network" in
  mainnet)
    net="";;
  testnet)
    net="-testnet";;
  regtest)
    net="-regtest";;
  *)
    net="-regtest";;
esac

if [ ! -z "$port" ]; then
  rpcport="--rpcport=$port"
fi

cli="./bitcoin-cli $net --datadir=$datadir $rpcport" 

if [ "$1" == "tx" ]; then
  shift
  $cli getrawtransaction $1 1
elif [ "$1" == "rawtx" ]; then
  shift
  $cli getrawtransaction $1 0
elif [ "$1" == "block" ]; then
  shift
  $cli getblock $($cli getblockhash $1)
elif [ "$1" == "decode" ]; then
  shift
  $cli decoderawtransaction $1
elif [ "$1" == "api" ]; then
  shift
  if [ $1 == "block" ]; then
    shift
    get_result block $(get_result block-index $1|jq -r .blockHash)
  elif [ "$1" == "tx" ]; then
    shift
    get_result tx "$@"
  else
    get_result "$@" 
  fi
else
  $cli "$@"
fi
check_debug
