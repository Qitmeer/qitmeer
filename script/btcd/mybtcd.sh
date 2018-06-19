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
    -pass|--passphrase)
      passphrase=$2
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

#case "$network" in
#  mainnet)
#    server="";;
#  testnet)
#    server="-testnet";;
#  regtest)
#    server="-s localhost:18334";;
#  *)
#    server="-s localhost:18111";;
#esac

cli="./btcctl -C $datadir/btcd.conf --rpccert=$datadir/rpc.cert --simnet" 

function unlockwallet() {
  if [ ! -z "$passphrase" ]; then
    $cli --wallet walletpassphrase "$passphrase" 10
  else
    echo "unlock wallet with -pass <passphrase>"
    exit -1
  fi 
}

function call_json_rpc() {
  local user=3d9d8d1f97e9e352bcc578eaf4cc2a0f67214ab50b854e70bdec61a8b221dfaa
  local pass=b67903f2240c3acd453e04ba0fa0780cfb1462d88456c6f9699b2f1d60392789
  /usr/local/opt/curl/bin/curl -s -k --http1.1 -H "Content-Type:application/json" -u $user:$pass --data '{"jsonrpc":"1.0","method":"getbalance","params":[],"id":1}' "$@" https://localhost:18554  
}

function call_grpc() {
 /usr/local/opt/curl/bin/curl -s -k  -H "Content-Type: application/grpc" --data-binary @$1 https://localhost:18558/walletrpc.WalletService/Balance |xxd -p|cut -c11-|xxd -r -p|protoc --decode_raw
}

### tx & block
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

### Wallet commands
elif [ "$1" == "addr" ]; then
  shift
  $cli --wallet getaccountaddress $1
elif [ "$1" == "newaddr" ]; then
  shift
  $cli --wallet getnewaddress
elif [ "$1" == "newaccount" ]; then
  shift
  unlockwallet
  $cli --wallet createnewaccount $1 
elif [ "$1" == "accounts" ]; then
  shift
  $cli --wallet listaccounts 
elif [ "$1" == "sendtoaddr" ]; then
  shift
  unlockwallet
  $cli --wallet sendtoaddress "$@"
elif [ "$1" == "balance" ]; then
  shift
  $cli --wallet getbalance "$@" 
elif [ "$1" == "received" ]; then
  shift
  $cli --wallet listreceivedbyaddress "$@"
  #$cli --wallet listreceivedbyaccount 

### curl json
elif [ "$1" == "jsonrpc" ]; then
  shift
  if [ -z "$1" ]; then 
    call_json_rpc |jq . 
  else
    call_json_rpc "$@" 
  fi
### curl grpc
elif [ "$1" == "grpc" ]; then
  shift
  if [ "$1" == "-in" ]; then 
    shift
    call_grpc $1
  fi
### Web API
elif [ "$1" == "api" ]; then
  shift
  if [ "$1" == "block" ]; then
    shift
    get_api block "$@" 
  elif [ "$1" == "tx" ]; then
    shift
    get_api tx "$@"
  else
    get_api "$@" 
  fi

### native cli 
else
  $cli "$@"
fi
check_debug
