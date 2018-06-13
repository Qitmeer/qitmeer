#!/bin/bash

# set -e

# ---------------------------
# settings
# ---------------------------
DEBUG=0
DEBUG_FILE=/tmp/btcd_curl_debug
ERR_FILE=/tmp/btcd_curl_error

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
  datadir="/data/btcd/private/" 
fi

case "$network" in
  mainnet)
    server="";;
  testnet)
    server="-testnet";;
  regtest)
    server="-s localhost:18334";;
  *)
    server="-s localhost:18111";;
esac

cli="./btcctl -C $datadir/btcd.conf --rpccert=$datadir/rpc.cert --simnet" 

if [ "$1" == "tx" ]; then
  shift
  $cli getrawtransaction $1 1
elif [ "$1" == "rawtx" ]; then
  shift
  $cli getrawtransaction $1 0
elif [ "$1" == "block" ]; then
  shift
  if [ "$1" == "latest" ]; then
    shift
    $cli getblock $($cli getblockhash $($cli getblockcount))
  elif [ "$2" == "txcount" ]; then
    $cli getblock $($cli getblockhash $1) |jq -r '[.tx[]]|length'
  else
    $cli getblock $($cli getblockhash $1)
  fi
elif [ "$1" == "decode" ]; then
  shift
  $cli decoderawtransaction $1
elif [ "$1" == "api" ]; then
  shift
  if [ $1 == "block" ]; then
    shift
    get_api block "$@" 
  elif [ "$1" == "tx" ]; then
    shift
    get_api tx "$@"
  else
    get_api "$@" 
  fi
else
  $cli "$@"
fi
check_debug
