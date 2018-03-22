#!/bin/bash

set -e

# settings

ERR_FILE=/tmp/myeth_curl_result

# All jsonrpc calls 

function get_tx_by_blocknum_and_index_hex(){
  local block_num=$1 #"0x467a65"
  local tx_index=$2  #"0x0"
  local data='{"jsonrpc":"2.0","method":"eth_getTransactionByBlockNumberAndIndex","params":["'$block_num'","'$tx_index'"],"id":1}'
  get_result "$data"
}
function get_tx_by_hash(){
  local tx_hash=$1
  local data='{"jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["'$tx_hash'"],"id":1}'
  get_result "$data"
}
function get_block_number(){
  local data='{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":83}'
  get_result "$data"
}
function get_syncing(){
  local data='{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}'
  get_result "$data"
}
function get_block(){
  local block_number=$1
  local hex_prefix=${block_number:0:2}
  local num_hex=${block_number:2}
  if [ ! "$hex_prefix" == "0x" ] ;then
    # $block_number not using hex_prefix, try to convert hex"
    num_hex=$(echo "obase=16;$block_number"|bc)
  fi
  local data='{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x'$num_hex'",true],"id":1}'
  get_result "$data" 
}
function get_block_by_hash(){
  local block_hash=$1
  local data='{"jsonrpc":"2.0","method":"eth_getBlockByHash","params":["'$block_hash'",true],"id":1}'
  get_result "$data"
}
function get_balance(){
  local addr=$1
  local block_num=$2
  if [ "$block_num" == "" ]; then
    #echo '{"result":"block_tag is latest and addr is '$addr'"}'
    block_num="latest"
  fi
  local data='{"jsonrpc":"2.0","method":"eth_getBalance","params":["'$addr'","'$block_num'"],"id":1}'
  get_result "$data"
}
function get_code(){
  #local contract_addr="0xbdeb5b87843062116b118e574a68a58f511a30e6"
  local contract_addr=$1
  local data='{"jsonrpc":"2.0","method":"eth_getCode","params":["'$contract_addr'","latest"],"id":1}'
  get_result "$data"
}
function get_storage(){
  #local contract_addr="0xbdeb5b87843062116b118e574a68a58f511a30e6"
  #local at="0x0"
  local contract_addr=$1 
  local at=$2 
  local data='{"jsonrpc":"2.0","method":"eth_getStorageAt","params":["'$contract_addr'","'$at'","latest"],"id":1}'
  get_result "$data"
}
function get_tx_count_by_addr(){
  local addr=$1
  local data='{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["'$addr'","latest"],"id":1}'
  get_result "$data"
}
function get_result(){
  set +x
  #local site=10.0.0.6
  #local site=10.0.0.8
  if [ -z "$host" ]; then
     host=127.0.0.1
  fi
  if [ -z "$port" ]; then
     port=8545
  fi
  local data=$1
  local curl_result=$(curl -s -X POST -H 'Content-Type: application/json' --data $data http://$host:$port)
  local result=$(echo $curl_result|jq -r -c -M '.result')
  if [ "$result" == "null" ];then
    echo "$curl_result" > $ERR_FILE
    result="" 
  fi
  echo $result
}

# util functions 
function pad_hex_prefix(){
  local input=$1
  if [ "${input:0:2}" == "0x" ];then
    echo "$input"
  else
    echo "0x$input"
  fi
}
function to_hex() {
  printf "%x\n" $1 
}
function to_hex_with_0x_prefix(){
  printf "0x%x\n" $1 
}
function check_error() {
  if [ -e $ERR_FILE ]; then
    echo "Error:"
    cat $ERR_FILE|jq .error
    rm $ERR_FILE 
    exit -1
  fi
}

function usage(){

  echo "chain :"
  echo "  get_block_number"
  echo "  get_syncing_info"
  echo "  get_current_block"
  echo "  get_current_block2 <num|hash> [-tx |-txcount|-blocktime|...]"
  echo "  get_highest_block"
  echo "block :"
  echo "  get_block <num|hash>"
  echo "  get_block <num|hash> -tx [num]"
  echo "  get_block <num|hash> -txcount"
  echo "  get_block <num|hash> -blocktime"
  echo "  get_block <num|hash> -stroot"
  echo "  get_block <num|hash> -txroot"
  echo "  get_block <num|hash> -rcroot"
  echo "  get_block <num|hash> -roots"
  echo "tx    :"
  echo "  get_tx <hash>"
  echo "  get_tx_by_block_and_index <num_hex> <index_hex>"
  echo "addr  :"
  echo "  get_balance <addr> [blocknum]"
  echo "  get_tx_count <addr>"
  echo "  get_code <addr> <at>"
  echo "  get_storage <addr> <at>"

}
# level 2 functions


function get_current_block_num(){
  get_syncing $@|jq .currentBlock -r|xargs printf "%d\n"
} 

# control function
function cmd_get_block(){
  # echo "debug cmd_get_block $@"
  if [ "$1" == "" ] ;then
      echo "get lastet block"
      blocknum=$(get_block_number|xargs printf "%d")
      echo "the lastet block is $blocknum"
      exit 
  fi
  if [ "${1:0:2}" == "0x" ];then
    block_hash=$1
    block_result=$(get_block_by_hash $(pad_hex_prefix $block_hash) $@)
  else 
    blocknum=$1
    block_result=$(get_block "$blocknum")
  fi

  #echo debug $block_result
  if [ "$block_result" == "null" ];then
    echo "block $1 not found"
    exit -1
  fi
  shift
  if [ -z "$1" ]; then
    echo $block_result|jq '.'
  elif [ "$1" == "-tx" ]; then
    shift
    tx=$1
    if [ "${tx:0:2}" == "0x" ];then
      echo $block_result|jq '.transactions|.[]|select(.transactionIndex == "'$tx'")'
    else
      echo $block_result|jq '.transactions['$tx']'
    fi
  elif [ "$1" == "-txcount" ];then
    shift
    echo $block_result|jq '.transactions|length'
  elif [ "$1" == "-blocktime" ];then
    shift
    echo $block_result|jq '.timestamp'| hex2dec|timestamp
  elif [ "$1" == "-stroot" ];then
    echo $block_result|jq '.stateRoot'
    shift
  elif [ "$1" == "-txroot" ];then
    echo $block_result|jq '.transactionsRoot'
    shift
  elif [ "$1" == "-rcroot" ];then
    echo $block_result|jq '.receiptsRoot'
    shift
  elif [ "$1" == "-roots" ]; then
    echo $block_result|jq '{"stroot":.stateRoot, "txroot":.transactionsRoot, "rcroot":.receiptsRoot}'
  fi
}

# main logic 
args=$(getopt h:p: "$@")
if [ $? != 0 ]; then
  echo "Usage: -h [host] -p [port] "
  exit;
fi
set -- $args
#echo $@
while [ -n "$1" ] ;do
  case "$1" in 
    -h) 
      host=$2
      #echo "host is $host"
      shift;;
    -p)
      port=$2
      #echo "port is $port"
      shift;;
    --)
      shift
      cmd=$@
      #echo "cmd is $cmd"
      break;;
    *)
      echo "$1 not a option"
  esac
  shift
done
#echo "get opt done!"

set -- $cmd
#echo $@
if [ $1 == "get_block" ]; then
  shift
  cmd_get_block $@
elif [ $1 == "get_block_number" ]; then
  shift
  if [ "$1" == "-hex" ]; then # result ishex by default
    get_block_number
  else                        # human can read (hex->decimal)
    get_block_number |xargs printf "%d\n"
  fi
elif [ $1 == "get_syncing_info" ]; then
  shift
  get_syncing $@
elif [ $1 == "get_current_block" ]; then
  shift
  get_current_block_num
elif [ $1 == "get_current_block2" ]; then
  shift
  blocknum=$(get_current_block_num)
  cmd_get_block $blocknum $@ 
elif [ $1 == "get_highest_block" ]; then
  shift
  get_syncing $@|jq .highestBlock -r|xargs printf "%d\n"
elif [ $1 == "get_tx" ]; then
  shift
  get_tx_by_hash $@
elif [ $1 == "get_tx_by_block_and_index" ]; then
  shift
  # note: the input is block number & tx index in hex
  get_tx_by_blocknum_and_index_hex $@
elif [ $1 == "get_balance" ]; then
  shift
  addr=$(pad_hex_prefix $1)
  shift
  if [ "${1:0:2}" == "0x" ];then
    num=$1
  else
    num=$(to_hex_with_0x_prefix $1)
  fi
  shift
  balance=$(get_balance $addr $num $@)
  check_error
  #echo "debug get_balance $addr $num --> $balance"
  echo $balance|xargs -I {} python -c 'print "%.4f ether" % ('{}/1000000000000000000.0')'
elif [ $1 == "get_code" ]; then
  shift
  addr=$1
  shift
  get_code $(pad_hex_prefix $addr) $@
elif [ $1 == "get_storage" ]; then
  shift
  addr=$1
  shift
  at=$1
  shift
  get_storage $(pad_hex_prefix $addr) $at $@
elif [ $1 == "get_tx_count" ]; then
  shift
  addr=$1
  shift
  if [ "$1" == "-h" ]; then
    get_tx_count_by_addr $(pad_hex_prefix $addr) $@|xargs printf "%d\n"
  else
    get_tx_count_by_addr $(pad_hex_prefix $addr) $@
  fi
elif [ $1 == "list_command" ]; then
  usage
else
  echo "Unkown cmd : $1"
  usage
  exit -1
fi
