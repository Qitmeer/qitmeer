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


# newAccount
# Generates a new private key and stores it in the key store directory. The key file is encrypted with the given passphrase. 
# Returns the address of the new account.
#    func (s *PrivateAccountAPI) NewAccount(password string) (common.Address, error) 
#
function new_account(){
  local passphrase=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -p|-pass)     shift; passphrase=$1; shift;;
      *)            shift;;
    esac
  done
  if [ "$passphrase" == "" ]; then
    # passphrase="test"
    passphrase=""
  fi
  local payload='{"jsonrpc":"2.0","method":"newAccount","params":["'$passphrase'"],"id":1}'
  get_result "$payload"
}

# returns the requested block by blockNr
# When fullTx is true all transactions in the block are # returned in full detail, otherwise only the transaction hash is returned.
#   func (s *PublicBlockChainAPI) GetBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) (map[string]interface{}, error)
function get_block(){
  local block_number=$(to_hex $1)
  local fullTx=$2
  if [ "$fullTx" == "" ]; then
    fullTx="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getBlockByNumber","params":["'$block_number'",'$fullTx'],"id":1}'
  get_result "$data"
}
# return block by hash
#   func (s *PublicBlockChainAPI) GetBlockByHash(ctx context.Context, blockHash common.Hash, fullTx bool) (map[string]interface{}, error)
function get_block_by_hash(){
  local block_hash=$1
  local data='{"jsonrpc":"2.0","method":"_getBlockByHash","params":["'$block_hash'",true],"id":1}'
  get_result "$data"
}

# the amount of wei for the given address in the state of the given block number.
#   func (s *PublicBlockChainAPI) GetBalance(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (*big.Int, error)
function get_balance(){
  local addr=$1
  local block_num=$2
  if [ "$block_num" == "" ]; then
    block_num="latest"
  fi
  local data='{"jsonrpc":"2.0","method":"getBalance","params":["'$addr'","'$block_num'"],"id":null}'
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
  local result=$(echo $curl_result|jq -r -c -M '.result')
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

  echo "chain    :"
  echo "block    :"
  echo "  block <num|hash>"
  echo "  block -n <numb> | -h <hash>"
  echo "  block <num> -show rlp|<num> -show rlp=dump"
  echo "  block <num|hash> -show tx=[num]"
  echo "  block <num|hash> -show txcount"
  echo "  block <num|hash> -show blocktime"
  echo "  block <num|hash> -show stroot"
  echo "  block <num|hash> -show txroot"
  echo "  block <num|hash> -show rcroot"
  echo "  block <num|hash> -show roots"
  echo "tx       :"
  echo "  tx <hash>"
  echo "  get_tx_by_block_and_index <num_hex> <index_hex>"
  echo "account  :"
  echo "  newaccount"
  echo "  accounts"
  echo "  balance      <account> [blknum]"
  echo "  get_tx_count <account>"
  echo "  get_code     <account> [blknum]"
  echo "  get_storage  <account> <at> [blknum]"
  echo "mining   :"
  echo "  get_coinbase"
  echo "  set_coinbase"
  echo "  start_mining"
  echo "  stop_mining"
  echo "  mining <second>"
  echo "status   :"
  echo "  status|get_status|info|get_info [-mining|-module|-hashrate|-work|-txpool|-all]"
  echo "contract :"
  echo "  compile        <sol_file> [-bin|-abi|-fun|-all] [-q]"
  echo "  send_tx        -from <addr> -to <addr> -v <value> -d <data>"
  echo "  receipt        <tx_hash>"
  echo "  contractaddr   <tx_hash>"
  echo "  call           -from <addr> -to <addr> -v <value> -d <data>"
  echo "debug   :"
  echo "  dump_state    <blknum>"
  echo "  rlp_block     <num>"
  echo "  trace_block   [-n <num>|-h <hash>|-r <rlp>]"
  echo "  trace_tx      <tx_hash>"
  echo "txpool  :"
  echo "  txpool -status|-inspect|-content"

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

  while [[ $# -gt 1 ]]; do
    case "$1" in
      -n|-num)  shift; blknum=$1; shift ;;
      -h|-hash) shift; blkhash=$1; shift ;;
      -show)    shift; show=${1%%=*}; show_opt="${1#*=}"; shift ;;
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
       block_result=$(get_block "$blknum")
    elif ! [ "$blkhash" == "" ]; then
       block_result=$(get_block_by_hash $blkhash)
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


## Block
if [ "$1" == "block" ]; then
  shift
  if [ "$1" == "latest" ]; then
    shift
    call_get_block $(get_block_number)
  else
    call_get_block $@
  fi
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

## Tx
elif [ $1 == "tx" ]; then
  shift
  get_tx_by_hash $@ |jq .
elif [ $1 == "get_tx_by_block_and_index" ]; then
  shift
  # note: the input is block number & tx index in hex
  get_tx_by_blocknum_and_index_hex $@

## Accounts
elif [ $1 == "newaccount" ]; then
  shift
  new_account "$@"
  check_error
elif [ $1 == "accounts" ]; then
  shift
  accounts=$(get_accounts)
  check_error
  if [ -z "$1" ]; then
    echo $accounts|jq '.'
  else
    echo $accounts|jq '.['$1']' -r
  fi
elif [ $1 == "balance" ]; then
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
elif [ $1 == "get_tx_count" ]; then
  shift
  addr=$1
  shift
  if [ "$1" == "-h" ]; then
    get_tx_count_by_addr $(pad_hex_prefix $addr) $@|xargs printf "%d\n"
  else
    get_tx_count_by_addr $(pad_hex_prefix $addr) $@
  fi
elif [ $1 == "get_code" ]; then
  shift
  addr=$(pad_hex_prefix $1)
  shift
  if [ ! -z "$1" ]; then
    num=$(to_hex $1)
    shift
  fi
  get_code $addr $num
  check_error
elif [ $1 == "get_storage" ]; then
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
elif [ $1 == "get_coinbase" ]; then
  shift
  get_coinbase
elif [ $1 == "start_mining" ]; then
  shift
  start_mining
elif [ $1 == "stop_mining" ]; then
  shift
  stop_mining
elif [ $1 == "mining" ]; then
  shift
  start_mining
  sleep $1
  stop_mining

## INFO & STATUS
elif [ $1 == "status" ] || [ $1 == "info" ] || [ $1 == "get_status" ] || [ $1 == "get_info" ]; then
  shift
  get_status $@

## Execute
elif [ $1 == "compile" ]; then
  shift
  solc_compile "$@"
elif [ $1 == "call" ]; then
  shift
  nox_call $@ 
  check_error
elif [ $1 == "send_tx" ]; then
  shift
  send_tx $@
  check_error
elif [ $1 == "receipt" ]; then
  shift
  get_receipt $@ |jq .
  check_error
elif [ $1 == "contractaddr" ]; then
  shift
  get_receipt $@ |jq -r .contractAddress
  check_error
## TXPOOL
elif [ $1 == "txpool" ]; then
  shift
  txpool $@ |jq .
  check_error

## DEBUG Moduls
elif [ $1 == "dump_state" ]; then
  shift
  dump_block $@|jq .
  check_error
elif [ $1 == "rlp_block" ]; then
  shift
  block_rlp $@
  check_error
elif [ $1 == "trace_block" ]; then
  shift
  trace_block $@|jq .
  check_error
elif [ $1 == "trace_tx" ]; then
  shift
  trace_tx $@|jq .
  check_error

## UTILS
elif [ $1 == "to_hex" ]; then
  shift
  to_hex $1
elif [ $1 == "to_base64" ]; then
  shift
  to_base64 $1
elif [ $1 == "list_command" ]; then
  usage
else
  echo "Unkown cmd : $1"
  usage
  exit -1
fi
check_debug
