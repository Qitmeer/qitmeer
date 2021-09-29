#!/bin/bash

set -e

# ---------------------------
# settings
# ---------------------------
DEBUG=0
notls=0

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

# qitmeer
function get_block(){
  local order=$1
  local verbose=$2
  if [ "$verbose" == "" ]; then
    verbose="true"
  fi
  local inclTx=$3
  if [ "$inclTx" == "" ]; then
    inclTx="true"
  fi
  local fullTx=$4
  if [ "$fullTx" == "" ]; then
    fullTx="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getBlockByOrder","params":['$order','$verbose','$inclTx','$fullTx'],"id":1}'
  get_result "$data"
}

function get_block_v2(){
  local block_hash=$1
  local verbose=$2
  if [ "$verbose" == "" ]; then
    verbose="true"
  fi
  local inclTx=$3
  if [ "$inclTx" == "" ]; then
    inclTx="true"
  fi
  local fullTx=$4
  if [ "$fullTx" == "" ]; then
    fullTx="true"
  fi
  local addfees=$5
  if [ "$addfees" == "" ]; then
    addfees="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getBlockV2","params":["'$block_hash'",'$verbose','$inclTx','$fullTx'],"id":1}'
  get_result "$data"
}

function get_block_by_id(){
  local id=$1
  local verbose=$2
  if [ "$verbose" == "" ]; then
    verbose="true"
  fi
  local inclTx=$3
  if [ "$inclTx" == "" ]; then
    inclTx="true"
  fi
  local fullTx=$4
  if [ "$fullTx" == "" ]; then
    fullTx="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getBlockByID","params":['$id','$verbose','$inclTx','$fullTx'],"id":1}'
  get_result "$data"
}

function get_block_number(){
  local data='{"jsonrpc":"2.0","method":"getBlockCount","params":[],"id":1}'
  get_result "$data"
}

function get_block_local(){
  local data='{"jsonrpc":"2.0","method":"getBlockTotal","params":[],"id":1}'
  get_result "$data"
}

# qitmeer mempool

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

function get_mempool_count(){
  local data='{"jsonrpc":"2.0","method":"getMempoolCount","params":[],"id":1}'
  get_result "$data"
}

function save_mempool(){
  local data='{"jsonrpc":"2.0","method":"saveMempool","params":[],"id":1}'
  get_result "$data"
}

function miner_info(){
  local data='{"jsonrpc":"2.0","method":"getMinerInfo","params":[],"id":1}'
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
  local inclTx=$3
  if [ "$inclTx" == "" ]; then
    inclTx="true"
  fi
  local fullTx=$4
  if [ "$fullTx" == "" ]; then
    fullTx="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getBlock","params":["'$block_hash'",'$verbose','$inclTx','$fullTx'],"id":1}'
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
function get_tx_by_id(){
  local tx_hash=$1
  local verbose=$2
  if [ "$verbose" == "" ]; then
    verbose="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getRawTransaction","params":["'$tx_hash'",'$verbose'],"id":1}'
  get_result "$data"
}

function get_tx_by_hash(){
  local tx_hash=$1
  local verbose=$2
  if [ "$verbose" == "" ]; then
    verbose="true"
  fi
  local data='{"jsonrpc":"2.0","method":"getRawTransactionByHash","params":["'$tx_hash'",'$verbose'],"id":1}'
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
   local tokenPrivateKey=$3
   local data='{"jsonrpc":"2.0","method":"test_txSign","params":["'$private_key'","'$raw_tx'","'$tokenPrivateKey'"],"id":1}'
   get_result "$data"
}

#
function create_raw_tx() {
  local input=$1
  local data='{"jsonrpc":"2.0","method":"createRawTransaction","params":['$input'],"id":1}'
  get_result "$data"
}

function create_raw_txv2() {
  local input=$1
  local data='{"jsonrpc":"2.0","method":"createRawTransactionV2","params":['$input'],"id":1}'
  get_result "$data"
}

function create_token_raw_tx(){
  local txtype=$1
  local coinId=$2

  local coinName=$3
  local owners=$4
  local uplimit=$5
  local inputs=$6
  local amounts=$7
  local feeType=$8
  local feeValue=$9

  if [ "$coinName" == "" ]; then
    coinName=""
  fi

  if [ "$owners" == "" ]; then
    owners=""
  fi

  if [ "$uplimit" == "" ]; then
    uplimit=0
  fi

  if [ "$inputs" == "" ]; then
    inputs='[{"txid":"","vout":0}]'
  fi

  if [ "$amounts" == "" ]; then
    amounts='{"":0}'
  fi

  if [ "$feeType" == "" ]; then
    feeType=0
  fi
  if [ "$feeValue" == "" ]; then
    feeValue=0
  fi

  local data='{"jsonrpc":"2.0","method":"createTokenRawTransaction","params":["'$txtype'",'$coinId',"'$coinName'","'$owners'",'$uplimit','$inputs','$amounts','$feeType','$feeValue'],"id":1}'
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
  local powtype=$2
  if [ "$powtype" == "" ]; then
    powtype=6
  fi
  local data='{"jsonrpc":"2.0","method":"miner_generate","params":['$count','$powtype'],"id":null}'
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
  get_result "$data"
}

function get_block_template(){
  local capabilities=$1
  local powtype=$2
  if [ "$powtype" == "" ]; then
    powtype=6
  fi
  local data='{"jsonrpc":"2.0","method":"getBlockTemplate","params":[["'$capabilities'"],'$powtype'],"id":1}'
  get_result "$data"
}

function get_mainchain_height(){
  local data='{"jsonrpc":"2.0","method":"getMainChainHeight","params":[],"id":1}'
  get_result "$data"
}

function get_block_weight(){
  local block_hash=$1
  local data='{"jsonrpc":"2.0","method":"getBlockWeight","params":["'$block_hash'"],"id":1}'
  get_result "$data"
}

function set_log_level(){
  local level=$1
  local data='{"jsonrpc":"2.0","method":"log_setLogLevel","params":["'$level'"],"id":1}'
  get_result "$data"
}

function get_blockhash_range(){
  local blk_num0=$1
  local blk_num1=$2
  if [ "$block_num1" == "" ]; then
    block_num1=blk_num0
  fi
  local data='{"jsonrpc":"2.0","method":"getBlockhashByRange","params":['$blk_num0','$blk_num1'],"id":null}'
  get_result "$data"
}

function get_node_info(){
  local data='{"jsonrpc":"2.0","method":"getNodeInfo","params":[],"id":null}'
  get_result "$data"
}

function get_peer_info(){
  local verbose=$1
  local network=$2
  if [ "$verbose" == "" ]; then
    verbose="false"
  fi

  local data='{"jsonrpc":"2.0","method":"getPeerInfo","params":['$verbose',"'$network'"],"id":null}'
  get_result "$data"
}

function get_network_info(){
  local data='{"jsonrpc":"2.0","method":"getNetworkInfo","params":[],"id":null}'
  get_result "$data"
}

function get_rpc_info(){
  local data='{"jsonrpc":"2.0","method":"getRpcInfo","params":[],"id":null}'
  get_result "$data"
}

function get_orphans_total(){
  local data='{"jsonrpc":"2.0","method":"getOrphansTotal","params":[],"id":null}'
  get_result "$data"
}

function stop_node(){
  local data='{"jsonrpc":"2.0","method":"test_stop","params":[],"id":null}'
  get_result "$data"
}

function is_current(){
  local data='{"jsonrpc":"2.0","method":"isCurrent","params":[],"id":null}'
  get_result "$data"
}

function tips(){
  local data='{"jsonrpc":"2.0","method":"tips","params":[],"id":null}'
  get_result "$data"
}

function get_coinbase(){
  local block_hash=$1
  local verbose=$2
  if [ "$verbose" == "" ]; then
    verbose="false"
  fi
  local data='{"jsonrpc":"2.0","method":"getCoinbase","params":["'$block_hash'",'$verbose'],"id":1}'
  get_result "$data"
}

function ban_list(){
  local data='{"jsonrpc":"2.0","method":"test_banlist","params":[],"id":null}'
  get_result "$data"
}

function remove_ban(){
  local bhost=$1
  local data='{"jsonrpc":"2.0","method":"test_removeBan","params":["'$bhost'"],"id":1}'
  get_result "$data"
}

function set_rpc_maxclients(){
  local max=$1
  local data='{"jsonrpc":"2.0","method":"test_setRpcMaxClients","params":['$max'],"id":null}'
  get_result "$data"
}

function get_rawtxs(){
  local address=$1
  local param2=$2
  local param3=$3
  local param4=$4
  local param5=$5
  local param6=$6

  if [ "$param2" == "" ]; then
      param2=false
  fi
  if [ "$param3" == "" ]; then
      param3=100
  fi
  if [ "$param4" == "" ]; then
      param4=0
  fi
  if [ "$param5" == "" ]; then
      param5=false
  fi
  if [ "$param6" == "" ]; then
      param6=true
  fi
  local data='{"jsonrpc":"2.0","method":"getRawTransactions","params":["'$address'",'$param2','$param3','$param4','$param5','$param6'],"id":null}'
  get_result "$data"
}

function is_blue(){
  local block_hash=$1
  local data='{"jsonrpc":"2.0","method":"isBlue","params":["'$block_hash'"],"id":1}'
  get_result "$data"
}

function get_fees(){
  local block_hash=$1
  local data='{"jsonrpc":"2.0","method":"getFees","params":["'$block_hash'"],"id":1}'
  get_result "$data"
}

function time_info(){
  local block_hash=$1
  local data='{"jsonrpc":"2.0","method":"getTimeInfo","id":1}'
  get_result "$data"
}

function get_tokeninfo(){
  local data='{"jsonrpc":"2.0","method":"getTokenInfo","params":[],"id":null}'
  get_result "$data"
}

function submit_block() {
  local input=$1
  local data='{"jsonrpc":"2.0","method":"submitBlock","params":["'$input'"],"id":1}'
  get_result "$data"
}

function get_remote_gbt() {
  local powtype=$1
    if [ "$powtype" == "" ]; then
    powtype=8
  fi
  local data='{"jsonrpc":"2.0","method":"getRemoteGBT","params":['$powtype'],"id":1}'
  get_result "$data"
}

function submit_block_header() {
  local input=$1
  local data='{"jsonrpc":"2.0","method":"submitBlockHeader","params":["'$input'"],"id":1}'
  get_result "$data"
}

function get_subsidy(){
  local data='{"jsonrpc":"2.0","method":"getSubsidy","params":[],"id":null}'
  get_result "$data"
}

function get_result(){
  local proto="https"
  if [ $notls -eq 1 ]; then
     proto="http"
  fi
  if [ -z "$host" ]; then
     host=127.0.0.1
  fi
  if [ -z "$port" ]; then
     port=1234
  fi
  if [ -z "$user" ]; then
     user="test"
  fi
  if [ -z "$pass" ]; then
     pass="test"
  fi

  local data=$1
  local current_result=$(curl -s -k -u "$user:$pass" -X POST -H 'Content-Type: application/json' --data $data $proto://$host:$port)
  local result=$(echo $current_result|jq -r -M '.result')
  if [ $DEBUG -gt 0 ]; then
    local current_cmd="curl -s -k -u "$user:$pass" -X POST -H 'Content-Type: application/json' --data '"$data"' $proto://$host:$port"
    echo "$current_cmd" >> "./cli.debug"
    echo "$current_result" >> "./cli.debug"
  fi

  local hashjson=$(echo $result |grep "{")
  if [ "$hashjson" == "" ]; then
      echo $result
  else
      echo $result |jq .
  fi
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

function usage(){
  echo "chain  :"
  echo "  nodeinfo"
  echo "  peerinfo"
  echo "  networkinfo"
  echo "  rpcinfo"
  echo "  rpcmax <max>"
  echo "  main  <hash>"
  echo "  stop"
  echo "  banlist"
  echo "  removeban"
  echo "  loglevel [trace, debug, info, warn, error, critical]"
  echo "  timeinfo"
  echo "  subsidy"
  echo "block  :"
  echo "  block <order|hash>"
  echo "  blockid <id>"
  echo "  blockhash <order>"
  echo "  block_count"
  echo "  block_local"
  echo "  blockrange <start,end>"
  echo "  mainHeight"
  echo "  weight <hash>"
  echo "  orphanstotal"
  echo "  isblue <hash>   ;return [0:not blue;  1：blue  2：Cannot confirm]"
  echo "  iscurrent"
  echo "  tips"
  echo "  coinbase <hash>"
  echo "  fees <hash>"
  echo "  tokeninfo"
  echo "tx     :"
  echo "  tx <id>"
  echo "  txv2 <id>"
  echo "  txbyhash <hash>"
  echo "  createRawTx"
  echo "  createRawTxV2"
  echo "  createTokenRawTx"
  echo "  txSign <rawTx>"
  echo "  sendRawTx <signedRawTx>"
  echo "  getrawtxs <address>"
  echo "utxo   :"
  echo "  getutxo <tx_id> <index> <include_mempool,default=true>"
  echo "miner  :"
  echo "  template"
  echo "  generate <num>"
  echo "  mempool"
  echo "  mempool_count"
  echo "  savemempool"
  echo "  minerinfo"
  echo "  submitblock"
  echo "  submitblockheader"
  echo "  remotegbt"
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

  elif [ "$1" == "-txpool" ]; then
    txpool -status|jq .

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
  local verbose="true"

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
    -notls)
      notls=1
      ;;
    -h)
      host=$2
      #echo "host is $host"
      shift;;
    -p)
      port=$2
      #echo "port is $port"
      shift;;
    --user)
      user=$2
      #echo "user is $user"
      shift;;
    --password)
      pass=$2
      #echo "pass is $pass"
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
elif [ "$1" == "blockid" ]; then
  shift
  get_block_by_id $@
elif [ "$1" == "loglevel" ]; then
  shift
  set_log_level $@
elif [ "$1" == "blockv2" ]; then
  shift
  get_block_v2 $@
elif [ "$1" == "block_count" ]; then
   shift
   get_block_number
elif [ "$1" == "block_local" ]; then
   shift
   get_block_local

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
elif [ "$1" == "blockbyhash" ]; then
  shift
  get_block_by_hash $@

elif [ "$1" == "header" ]; then
  shift
  get_blockheader_by_hash $@

elif [ "$1" == "main" ]; then
    shift
    is_on_mainchain $1

elif [ "$1" == "template" ]; then
    shift
    get_block_template $@ | jq .

elif [ "$1" == "mainHeight" ]; then
    shift
    get_mainchain_height

elif [ "$1" == "weight" ]; then
    shift
    get_block_weight $1

elif [ "$1" == "blockrange" ]; then
  shift
  get_blockhash_range $@

elif [ "$1" == "isblue" ]; then
  shift
  is_blue $@

elif [ "$1" == "fees" ]; then
  shift
  get_fees $@

elif [ "$1" == "timeinfo" ]; then
  shift
  time_info $@

elif [ "$1" == "subsidy" ]; then
  shift
  get_subsidy $@

elif [ "$1" == "nodeinfo" ]; then
  shift
  get_node_info

elif [ "$1" == "peerinfo" ]; then
  shift
  get_peer_info $@

elif [ "$1" == "networkinfo" ]; then
  shift
  get_network_info $@

elif [ "$1" == "rpcinfo" ]; then
  shift
  get_rpc_info

elif [ "$1" == "rpcmax" ]; then
  shift
  set_rpc_maxclients $@

elif [ "$1" == "orphanstotal" ]; then
  shift
  get_orphans_total

elif [ "$1" == "stop" ]; then
  shift
  stop_node

elif [ "$1" == "iscurrent" ]; then
  shift
  is_current

elif [ "$1" == "tips" ]; then
  shift
  tips | jq .
elif [ "$1" == "tokeninfo" ]; then
  shift
  get_tokeninfo | jq .
elif [ "$1" == "coinbase" ]; then
  shift
  get_coinbase $@

elif [ "$1" == "banlist" ]; then
  shift
  ban_list | jq .

elif [ "$1" == "removeban" ]; then
  shift
  remove_ban $@

## Tx
elif [ "$1" == "tx" ]; then
  shift
  get_tx_by_id $@
elif [ "$1" == "txbyhash" ]; then
  shift
  get_tx_by_hash $@

elif [ "$1" == "createRawTx" ]; then
  shift
  create_raw_tx $@

elif [ "$1" == "createRawTxV2" ]; then
  shift
  create_raw_txv2 $@

elif [ "$1" == "createTokenRawTx" ]; then
  shift
  create_token_raw_tx $@

elif [ "$1" == "decodeRawTx" ]; then
  shift
  decode_raw_tx $@

elif [ "$1" == "sendRawTx" ]; then
  shift
  send_raw_tx $@

elif [ "$1" == "getrawtxs" ]; then
  shift
  get_rawtxs $@

elif [ "$1" == "get_tx_by_block_and_index" ]; then
  shift
  # note: the input is block number & tx index in hex
  get_tx_by_blocknum_and_index_hex $@

## MemPool
elif [ "$1" == "mempool" ]; then
  shift
  get_mempool $@

elif [ "$1" == "mempool_count" ]; then
  shift
  get_mempool_count $@

elif [ "$1" == "savemempool" ]; then
  shift
  save_mempool $@

elif [ "$1" == "minerinfo" ]; then
  shift
  miner_info $@

elif [ "$1" == "txSign" ]; then
  shift
  tx_sign $@

## UTXO
elif [ "$1" == "getutxo" ]; then
  shift
  get_utxo $@

## Accounts
elif [ "$1" == "newaccount" ]; then
  shift
  new_account "$@"

elif [ "$1" == "accounts" ]; then
  shift
  accounts=$(get_accounts)

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
  generate $@|jq .


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
  qitmeer_call $@

elif [ "$1" == "send_tx" ]; then
  shift
  send_tx $@

elif [ "$1" == "receipt" ]; then
  shift
  get_receipt $@

elif [ "$1" == "contractaddr" ]; then
  shift
  get_receipt $@ |jq -r .contractAddress

## TXPOOL
elif [ "$1" == "txpool" ]; then
  shift
  txpool $@


## DEBUG Moduls
elif [ "$1" == "dump_state" ]; then
  shift
  dump_block $@

elif [ "$1" == "rlp_block" ]; then
  shift
  block_rlp $@

elif [ "$1" == "trace_block" ]; then
  shift
  trace_block $@

elif [ "$1" == "trace_tx" ]; then
  shift
  trace_tx $@


## UTILS
elif [ "$1" == "to_hex" ]; then
  shift
  to_hex $1
elif [ "$1" == "to_base64" ]; then
  shift
  to_base64 $1

elif [ "$1" == "submitblock" ]; then
  shift
  submit_block $@

elif [ "$1" == "submitblockheader" ]; then
  shift
  submit_block_header $@

elif [ "$1" == "remotegbt" ]; then
  shift
  get_remote_gbt $@

elif [ "$1" == "list_command" ]; then
  usage
else
  echo "Unkown cmd : $1"
  usage
  exit -1
fi


# debug info
if [ $DEBUG -gt 0 ]; then
    echo -e "\nDebug info:"
    cat ./cli.debug
    rm ./cli.debug
fi
