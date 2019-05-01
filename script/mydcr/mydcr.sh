#!/bin/bash

# set -e

# ---------------------------
# settings
# ---------------------------
DEBUG=0
DEBUG_FILE=/tmp/mydcr_curl_debug
ERR_FILE=/tmp/mydcr_curl_error

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
  if [ ! -z "$3" ]; then
    local verbose="/$3"
  fi
  
  if [ "$method" == "tx" ]; then
    if [ -z "$decode" ] || [ "$decode" == "json" ]; then
      local method="tx/decoded"
    else
      local method="tx/$decode"
    fi
  fi

  # curl -s https://explorer.dcrdata.org/api/tx/decoded/5442e9dd4961ffefad37cb64b8c906574937b1c74850167c1af9d8587ed504ce|jq .

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

function do_jsonrpc() {
  local username=faMtZ/p6y/Ima/a9CssdLg4zXJg=
  local password=eCzUMNhvyYTJYI8Pjv5a+VDAEUw=
  #/usr/local/opt/curl/bin/curl -s -X POST --http1.1 -H "Content-Type:application/json"  -u $username:$password --cacert ./rpc.cert --data '{"jsonrpc":"1.0","method":"getbalance","params":[],"id":1}' https://localhost:19557|jq .
  local account=$1
  /usr/local/opt/curl/bin/curl -s -k -X POST --http1.1 -H "Content-Type:application/json" -u $username:$password --data '{"jsonrpc":"1.0","method":"getbalance","params":["'$account'"],"id":1}' https://localhost:19557
}

function do_grpc() {
  local username=faMtZ/p6y/Ima/a9CssdLg4zXJg=
  local password=eCzUMNhvyYTJYI8Pjv5a+VDAEUw=
  local input="account_number:$1,required_confirmations:$2"
  local msg=$(echo $input|protoc --encode=walletrpc.BalanceRequest api.proto |xxd -p)
  python -c "print '00' + hex(len('$msg')/2).lstrip('0x').zfill(8) + '$msg'"|xxd -r -p - >input.bin
  /usr/local/opt/curl/bin/curl -s -k -X POST --http2 -H "Content-Type:application/grpc" --data-binary @input.bin "$@" https://localhost:19558/walletrpc.WalletService/Balance |xxd -p|cut -c11-|xxd -r -p|protoc --decode=walletrpc.BalanceResponse api.proto
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
    -decode)
      decode=$2
      shift;;
    --network)
      network=$2
      shift;;
    *)
      cmd="$@"
      #echo "cmd is $cmd"
      break;;
    esac
  shift
done

case "$network" in
  main*|mainnet*)
    cli="./dcrctl -C ./dcrd.conf -c ./rpc.cert"
    ;;
  private*|simnet*|'')
    cli="./dcrctl -C ./dcrd.conf -c ./rpc.cert --simnet"
    ;;
  test*|testnet*)
    cli="./dcrctl -C ./dcrd.conf -c ./rpc.cert --testnet"
    ;;
  *)
    cli="./dcrctl -C ./dcrd.conf -c ./rpc.cert"
    ;;
esac

#echo $cli


if [ "$1" == "tx" ]; then
  shift
  $cli getrawtransaction $@
elif [ "$1" == "block" ]; then
  shift
  $cli getblock $($cli getblockhash $1)
elif [ "$1" == "api" ]; then
  shift
  get_result "$@" 
elif [ "$1" == "jsonrpc" ]; then
  shift
  do_jsonrpc "$@"
elif [ "$1" == "grpc" ]; then
  shift
  do_grpc "$@"
else
  $cli "$@"
fi
      

check_debug
