#!/bin/bash

set -e

# ---------------------------
# settings
# ---------------------------
DEBUG=0
DEBUG_FILE=/tmp/nox_curl_debug
ERR_FILE=/tmp/nox_curl_error

# ---------------------------
# solc call, need solc command line
# ---------------------------

function solc_compile(){
  local code="$1"
  local suppress_error=0
  local opts=""
  shift
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -q) suppress_error=1; shift ;;
      -all) opts="--bin --abi --hashes"; shift;;
      *) opts+=" $1"; shift;;
    esac
  done
  if [ "$opts" == "" ]; then
    solc --optimize --bin "$code" 2>/dev/null | awk /^[0-9]/
    return
  elif [ "$opts" == " -bin" ] ;then  # by-default, out cleanup binary
    solc --optimize --bin "$code" 2>/dev/null | awk /^[0-9]/ |jq -R {"binary":.}
    return
  elif [ "$opts" == " -abi" ];then
    solc --abi "$code" 2>/dev/null | tail -n 1 |jq .
    return
  elif [ "$opts" == " -func" ];then
    solc --hashes "$code" 2>/dev/null |awk /^[0-9a-f]/|jq -R '.|split(": ")|{name:.[1],sign: .[0]}'
    return
  fi;
  #echo "debug opts=$opts s_err=$suppress_error"
  if [[ $suppress_error -eq 0 ]]; then
    solc $opts "$code"
  else
    solc $opts "$code" 2>/dev/null
  fi
}


# ---------------------------
# All jsonrpc calls
# ---------------------------

# Nox
function get_block(){
  local order=$1
  local verbose=$2
  if [ "$verbose" == "" ]; then
    verbose="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getBlockByOrder","params":['$order','$verbose'],"id":1}'
  get_result "$data"
}

function get_block_number(){
  local data='{"jsonrpc":"2.0","method":"getBlockCount","params":[],"id":1}'
  get_result "$data"
}

# Nox mempool

function get_mempool(){
  local type=$1
  local verbose=$2
  if [ "$type" == "" ]; then
    type="regular"
  fi
  if [ "$verbose" == "" ]; then
    verbose="false"
  fi
  local data='{"jsonrpc":"2.0","method":"getMempool","params":["'$type'",'$verbose'],"id":1}'
  get_result "$data"
}

# return block by hash
#   func (s *PublicBlockChainAPI) GetBlockByHash(ctx context.Context, blockHash common.Hash, fullTx bool) (map[string]interface{}, error)
function get_block_by_hash(){
  local block_hash=$1
  local verbose=$2
  if [ "$verbose" == "" ]; then
    verbose="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getBlock","params":["'$block_hash'",'$verbose'],"id":1}'
  get_result "$data"
}

function get_blockheader_by_hash(){
  local block_hash=$1
  local verbose=$2
  if [ "$verbose" == "" ]; then
    verbose="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getBlockHeader","params":["'$block_hash'",'$verbose'],"id":1}'
  get_result "$data"
}


# return tx by hash
function get_tx_by_hash(){
  local tx_hash=$1
  local verbose=$2
  if [ "$verbose" == "" ]; then
    verbose="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getRawTransaction","params":["'$tx_hash'",'$verbose'],"id":1}'
  get_result "$data"
}

# return info about UTXO
function get_utxo() {
  local tx_hash=$1
  local vout=$2
  local include_mempool=$3
  if [ "$include_mempool" == "" ]; then
    include_mempool="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getUtxo","params":["'$tx_hash'",'$vout','$include_mempool'],"id":1}'
  get_result "$data"
}

function tx_sign(){
   local private_key=$1
   local raw_tx=$2
   local data='{"jsonrpc":"2.0","method":"txSign","params":["'$private_key'","'$raw_tx'"],"id":1}'
   get_result "$data"
}

#
function create_raw_tx(){
  local input=$1
  local data='{"jsonrpc":"2.0","method":"createRawTransaction","params":['$input'],"id":1}'
  get_result "$data"
}

function decode_raw_tx(){
  local input=$1
  local data='{"jsonrpc":"2.0","method":"decodeRawTransaction","params":["'$input'"],"id":1}'
  get_result "$data"
}

function send_raw_tx(){
  local input=$1
  local allow_high_fee=$2
  if [ "$allow_high_fee" == "" ]; then
    allow_high_fee="false"
  fi

  local data='{"jsonrpc":"2.0","method":"sendRawTransaction","params":["'$input'",'$allow_high_fee'],"id":1}'
  get_result "$data"
}

function generate() {
  local count=$1
  local block_num=$2
  if [ "$block_num" == "" ]; then
    block_num="latest"
  fi
  local data='{"jsonrpc":"2.0","method":"generate","params":['$count'],"id":null}'
  get_result "$data"
}

function get_blockhash(){
  local blk_num=$1
  local data='{"jsonrpc":"2.0","method":"getBlockhash","params":['$blk_num'],"id":null}'
  get_result "$data"
}

function is_on_mainchain(){
  local block_hash=$1
  local data='{"jsonrpc":"2.0","method":"isOnMainChain","params":["'$block_hash'"],"id":1}'
  echo $data
  get_result "$data"
}

function get_result(){
  set +x
  if [ -z "$host" ]; then
     host=127.0.0.1
  fi
  if [ -z "$port" ]; then
     port=1234
  fi
  local user="test"
  local pass="test"
  local data=$1
  local curl_result=$(curl -s -k -u "$user:$pass" -X POST -H 'Content-Type: application/json' --data $data https://$host:$port)
  local result=$(echo $curl_result|jq -r -M '.result')
  if [ $DEBUG -gt 0 ]; then
    local curl_cmd="curl -s -k -u "$user:$pass" -X POST -H 'Content-Type: application/json' --data '"$data"' https://$host:$port"
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
  echo "  txSign <rawTx>"
  echo "  sendRawTx <signedRawTx>"
  echo "utxo  :"
  echo "  getutxo <tx_id> <index> <include_mempool,default=true>"
  
}

# -------------------
# level 2 functions
# -------------------

function start_mining(){
  check_api_error_stop miner
  local mining=$(get_mining)
  if [ "$mining" == "false" ]; then
    echo start mining ...
    $(miner_start)
    get_mining
  else
    echo already stated
    get_mining_status
  fi
}

function stop_mining(){
  check_api_error_stop miner
  local mining=$(get_mining)
  if [ "$mining" == "true" ]; then
    echo stop mining ...
    miner_stop
  else
    echo already stopped
    get_mining_status
  fi
}

function check_api_error_stop(){
  if [ "$(check_api $1)" == "" ]; then
    echo "$1 api not find, need to enable the management APIs"
    echo "For example: geth --rpcapi eth,web3,miner --rpc"
    exit
  fi
}

function check_api() {
  local api=$1
  if ! [ "$api" == "" ]; then
    local api_ver=$(get_rpc_modules|jq .$api -r)
     if ! [ "$api_ver" == "null" ]; then
         echo $api_ver
     fi
  fi
}
function get_mining_status(){
  local mining=$(get_mining)
  local hashrate=$(get_hashrate)
  local gasprice=$(get_gasprice)
  local gasprice_dec=$(printf "%d" $gasprice)
  echo "mining   : $mining"
  echo "hashRate : $hashrate"
  echo "gasPrice : $gasprice -> ($gasprice_dec wei/$(wei_to_ether $gasprice) ether)"
}
function get_modules(){
  get_rpc_modules|jq . -r
}

function get_status() {
  #echo "debug get_status $@"
  if [ "$1" == "" ]; then
    get_modules
  elif [ "$1" == "-module" ]; then
    get_modules
  elif [ "$1" == "-mining" ]; then
    get_mining_status
  elif [ "$1" == "-hashrate" ]; then
    get_hashrate
  elif [ "$1" == "-work" ]; then
    get_work|jq .
    check_error
  elif [ "$1" == "-txpool" ]; then
    txpool -status|jq .
    check_error
  elif [ "$1" == "-all" ]; then
    echo "modules  : $(get_modules|jq -c -M .)"
    get_mining_status
  else
    echo "unknown opt $@"
  fi
}

function get_current_block_num(){
  get_syncing $@|jq .currentBlock -r|xargs printf "%d\n"
}

function call_get_block() {
  # echo "debug call_get_block $@"
  local blknum=""
  local blkhash=""
  local show=""
  local show_opt=""
  local verbose="false"

  if [[ $# -eq 0 ]]; then
    # echo "get lastet block"
    local latest_num=$(get_block_number|xargs printf "%d")
    echo "the lastet block is $latest_num"
    exit
  fi

  if ! [ "${1:0:1}" == "-" ]; then
    if [ "${1:0:2}" == "0x" ] && [[ ${#1} -eq 66 ]] ; then # 64
      blkhash=$1
    else
      blknum=$1
    fi
    shift
  fi

  while [[ $# -gt 0 ]]; do
    case "$1" in
      -n|-num)  shift; blknum=$1; shift ;;
      -h|-hash) shift; blkhash=$1; shift ;;
      -show)    shift; show=${1%%=*}; show_opt="${1#*=}"; shift ;;
      -v|-verbose|--verbose) shift; verbose="true"; ;;
      *) echo '{ "error" : "unkown option: '$1'"}'|jq .; exit -1;;
    esac
  done

  if [ "$show_opt" == "$show" ]; then show_opt=""; fi;
  #echo "debug: blknum=$blknum,blkhash=$blkhash,show=$show;show_opt=$show_opt"

  if [ "$show" == "rlp" ]; then
    if ! [ "$blknum" == "" ]; then
      if [ "$show_opt" == "dump" ];then
        block_rlp $blknum|rlpdump
      else
        block_rlp $blknum
      fi
      return
    else
       echo '{ "error" : "show rlp only support by using blknum."}'|jq .; exit -1;
    fi
  else #default show all
    if ! [ "$blknum" == "" ]; then
       block_result=$(get_block "$blknum" "$verbose")
    elif ! [ "$blkhash" == "" ]; then
       block_result=$(get_block_by_hash "$blkhash" "$verbose")
       if [ "$verbose" != "true" ]; then
         block_result='{ "hex" : "'$block_result'"}'
       fi
    else
       echo '{ "error" : "need to provide blknum or blkhash"}'; exit -1;
    fi
  fi
  check_error -e

  if [ "$show" == "" ]; then
    echo $block_result|jq '.'
  elif [ "$show" == "tx" ]; then
    tx=$show_opt
    if [ "${tx:0:2}" == "0x" ];then
      echo $block_result|jq '.transactions|.[]|select(.transactionIndex == "'$tx'")'
    else
      echo $block_result|jq '.transactions['$tx']'
    fi
  elif [ "$show" == "txcount" ]; then
    echo $block_result|jq '.transactions|length'
  elif [ "$show" == "blocktime" ];then
    echo $block_result|jq '.timestamp'|hex2dec|timestamp
  elif [ "$show" == "stroot" ];then
    echo $block_result|jq '.stateRoot'
  elif [ "$show" == "txroot" ];then
    echo $block_result|jq '.transactionsRoot'
  elif [ "$show" == "rcroot" ];then
    echo $block_result|jq '.receiptsRoot'
  elif [ "$show" == "roots" ]; then
    echo $block_result|jq '{"stroot":.stateRoot, "txroot":.transactionsRoot, "rcroot":.receiptsRoot}'
  else
    echo '{ "error" : "unkown option: '$show'"}'; exit -1;
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

## chain
if [ "$1" == "main" ]; then
    shift
    is_on_mainchain $1
    check_error
## Block
elif [ "$1" == "block" ]; then
  shift
  if [ "$1" == "latest" ]; then
    shift
    call_get_block $(get_block_number)
  else
    call_get_block $@
  fi
elif [ "$1" == "get_block_number" ]; then
  shift
  if [ "$1" == "-hex" ]; then # result ishex by default
    get_block_number
  else                        # human can read (hex->decimal)
    get_block_number |xargs printf "%d\n"
  fi
elif [ "$1" == "get_syncing_info" ]; then
  shift
  get_syncing $@
elif [ "$1" == "get_current_block" ]; then
  shift
  get_current_block_num
elif [ "$1" == "get_current_block2" ]; then
  shift
  blocknum=$(get_current_block_num)
  cmd_get_block $blocknum $@
elif [ "$1" == "get_highest_block" ]; then
  shift
  get_syncing $@|jq .highestBlock -r|xargs printf "%d\n"
elif [ "$1" == "blockhash" ]; then
  shift
  get_blockhash $1
  check_error
elif [ "$1" == "header" ]; then
  shift
  get_blockheader_by_hash $@
  check_error

## Tx
elif [ "$1" == "tx" ]; then
  shift
  if [ "$2" == "false" ]; then
    get_tx_by_hash $@
  else
    get_tx_by_hash $@|jq .
  fi
  check_error
elif [ "$1" == "createRawTx" ]; then
  shift
  create_raw_tx $@
  check_error
elif [ "$1" == "decodeRawTx" ]; then
  shift
  decode_raw_tx $@|jq .
  check_error
elif [ "$1" == "sendRawTx" ]; then
  shift
  send_raw_tx $@
  check_error
elif [ "$1" == "get_tx_by_block_and_index" ]; then
  shift
  # note: the input is block number & tx index in hex
  get_tx_by_blocknum_and_index_hex $@

## MemPool
elif [ "$1" == "mempool" ]; then
  shift
  get_mempool $@|jq .
  check_error

elif [ "$1" == "txSign" ]; then
  shift
  tx_sign $@
  echo $@
  check_error

## UTXO
elif [ "$1" == "getutxo" ]; then
  shift
  get_utxo $@|jq .

## Accounts
elif [ "$1" == "newaccount" ]; then
  shift
  new_account "$@"
  check_error
elif [ "$1" == "accounts" ]; then
  shift
  accounts=$(get_accounts)
  check_error
  if [ -z "$1" ]; then
    echo $accounts|jq '.'
  else
    echo $accounts|jq '.['$1']' -r
  fi
elif [ "$1" == "balance" ]; then
  shift
  addr=$(pad_hex_prefix $1)
  shift
  if [ ! -z "$1" ]; then
    num=$(to_hex $1)
    shift
  fi
  # echo "debug get_balance $addr $num"
  balance=$(get_balance $addr $num $@)
  check_error
  echo $balance
  # echo "debug get_balance $addr $num --> $balance"
elif [ "$1" == "get_tx_count" ]; then
  shift
  addr=$1
  shift
  if [ "$1" == "-h" ]; then
    get_tx_count_by_addr $(pad_hex_prefix $addr) $@|xargs printf "%d\n"
  else
    get_tx_count_by_addr $(pad_hex_prefix $addr) $@
  fi
elif [ "$1" == "get_code" ]; then
  shift
  addr=$(pad_hex_prefix $1)
  shift
  if [ ! -z "$1" ]; then
    num=$(to_hex $1)
    shift
  fi
  get_code $addr $num
  check_error
elif [ "$1" == "get_storage" ]; then
  shift
  addr=$(pad_hex_prefix $1)
  shift
  if [ ! -z "$1" ]; then
    at=$(to_hex $1)
    shift
  fi
  if [ ! -z "$1" ]; then
    num=$(to_hex $1)
    shift
  fi
  get_storage $addr $at $num
  check_error

## Mining
elif [ "$1" == "get_coinbase" ]; then
  shift
  get_coinbase
elif [ "$1" == "start_mining" ]; then
  shift
  start_mining
elif [ "$1" == "stop_mining" ]; then
  shift
  stop_mining
elif [ "$1" == "mining" ]; then
  shift
  start_mining
  sleep $1
  stop_mining
elif [ "$1" == "generate" ]; then
  shift
  generate $1|jq .
  check_error

## INFO & STATUS
elif [ "$1" == "status" ] || [ "$1" == "info" ] || [ "$1" == "get_status" ] || [ "$1" == "get_info" ]; then
  shift
  get_status $@

## Execute
elif [ "$1" == "compile" ]; then
  shift
  solc_compile "$@"
elif [ "$1" == "call" ]; then
  shift
  nox_call $@
  check_error
elif [ "$1" == "send_tx" ]; then
  shift
  send_tx $@
  check_error
elif [ "$1" == "receipt" ]; then
  shift
  get_receipt $@ |jq .
  check_error
elif [ "$1" == "contractaddr" ]; then
  shift
  get_receipt $@ |jq -r .contractAddress
  check_error
## TXPOOL
elif [ "$1" == "txpool" ]; then
  shift
  txpool $@ |jq .
  check_error

## DEBUG Moduls
elif [ "$1" == "dump_state" ]; then
  shift
  dump_block $@|jq .
  check_error
elif [ "$1" == "rlp_block" ]; then
  shift
  block_rlp $@
  check_error
elif [ "$1" == "trace_block" ]; then
  shift
  trace_block $@|jq .
  check_error
elif [ "$1" == "trace_tx" ]; then
  shift
  trace_tx $@|jq .
  check_error

## UTILS
elif [ "$1" == "to_hex" ]; then
  shift
  to_hex $1
elif [ "$1" == "to_base64" ]; then
  shift
  to_base64 $1
elif [ "$1" == "list_command" ]; then
  usage
else
  echo "Unkown cmd : $1"
  usage
  exit -1
fi
check_debug
