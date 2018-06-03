#!/bin/bash

set -e

# ---------------------------
# settings
# ---------------------------
DEBUG=0
DEBUG_FILE=/tmp/mynem_curl_debug
ERR_FILE=/tmp/mynem_curl_error

function get_tx_by_hash(){
  local tx_hash=$1
  local data='{"jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["'$tx_hash'"],"id":1}'
  get_result "$data"
}

function get_block_by_number(){
  local block_number=$1
  local method="block/at/public"
  local data='{"height":'$block_number'}'
  
  get_result "POST" "$method" "$data"
}

function get_block_by_hash(){
  local block_hash=$1
  local data='{"jsonrpc":"2.0","method":"eth_getBlockByHash","params":["'$block_hash'",true],"id":1}'
  get_result "$data"
}

function get_result(){
  set +x
  if [ -z "$host" ]; then
     host=127.0.0.1
  fi
  if [ -z "$port" ]; then
     port=7890
  fi
  local type=$1
  local method=$2
  local data=$3
  if [ "$type" == "POST" ]; then
    local curl_result=$(curl -s -X POST -H 'Content-Type: application/json' --data $data http://$host:$port/$method)
  elif [ "$type" == "GET" ]; then
    local curl_result=$(curl -s -X GET http://$host:$port/$method)
  else
    check_err -e  
  fi

  local result=$(echo $curl_result|jq -r -c -M '.')
  if [ $DEBUG -gt 0 ]; then
    local curl_cmd="curl -s -X POST -H 'Content-Type: application/json' --data '"$data"' http://$host:$port/$method"
    echo "$curl_cmd" > $DEBUG_FILE
    echo "$curl_result" >> $DEBUG_FILE
  fi
  if [ "$result" == "null" ];then
    echo "$curl_result" > $ERR_FILE
    result=""
  fi
  echo $result
}

# -------------------------
# util functions
# -------------------------
function pad_hex_prefix(){
  local input=$1
  if [ "${input:0:2}" == "0x" ];then
    echo "$input"
  else
    echo "0x$input"
  fi
}

# convert int to hex, also add 0x prefix if missing
function to_hex() {
  printf "0x%x\n" $1
}
function to_dec() {
  printf "%d\n" $1
}

function to_base64() {
  echo -n $1|xxd -r -p|base64
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


function usage(){

  echo "block    :"
  echo "  block <num|hash>"
  echo "tx       :"
  echo "  tx <hash>"
  echo "account  :"
  echo "node  :"
  echo "  info"
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

## Block
if [ $1 == "block" ]; then
  shift
  get_block_by_number $@ |jq .
## Tx
elif [ $1 == "tx" ]; then
  shift
  #get_tx $@
## Accounts
elif [ $1 == "account" ]; then
  shift
  #get_acct $@
## INFO & STATUS
elif [ $1 == "node" ]; then
  shift
  #get_status $@
elif [ $1 == "get" ]; then
  shift
  get_result GET $@ |jq .
elif [ $1 == "post" ]; then
  shift
  get_result POST $@|jq .
else
  echo "Unkown cmd : $1"
  usage
  exit -1
fi
check_debug
