function get_tx_block_and_index(){
  local block_num=$1 #"0x467a65"
  local tx_index=$2  #"0x0"
  local data='{"jsonrpc":"2.0","method":"eth_getTransactionByBlockNumberAndIndex","params":["'$block_num'","'$tx_index'"],"id":1}'
  get_result "$data"
}
function get_block_number(){
  local data='{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":83}'
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
function get_block_byhash(){
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
function get_tx_count(){
  local addr=$1
  local data='{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["'$addr'","latest"],"id":1}'
  get_result "$data"
}
function get_result(){
  local site=10.0.0.6
  #local site=10.0.0.8
  #site=127.0.0.1
  #local port=8545
  local port=8546
  local data=$1
  curl -s -X POST -H 'Content-Type: application/json' --data $data http://$site:$port
}
function pad_hex_prefix(){
  local input=$1
  if [ "${input:0:2}" == "0x" ];then
    echo "$input"
  else
    echo "0x$input"
  fi
}
if [ $1 == "get_block" ]; then
  shift
  blocknum=$1
  if [ "$blocknum" == "" ] || [ "$1" == "-tx" ] ;then
      echo "get lastet block"
      blocknum=$(get_block_number|jq '.result'|xargs printf "%d")
      echo "the lastet block is $blocknum"
  else
      shift
  fi
  if [ "$1" == "-tx" ]; then
    shift
    tx=$1
    if [ "${tx:0:2}" == "0x" ];then
      get_block $blocknum |jq '.result.transactions|.[]|select(.transactionIndex == "'$tx'")'
    else
      get_block $blocknum |jq '.result.transactions['$tx']'
    fi
  else
    get_block "$blocknum" |jq .
  fi
elif [ $1 == "get_block_byhash" ]; then
  shift
  block_hash=$1
  shift
  get_block_byhash $(pad_hex_prefix $block_hash) $@ |jq '.result'
elif [ $1 == "get_block_number" ]; then
  shift
  if [ "$1" == "-h" ]; then # human (hex->decimal)
    get_block_number |jq '.result'|xargs printf "%d\n"
  else
    get_block_number |jq '.result'
  fi
elif [ $1 == "get_tx" ]; then
  shift
  get_tx_block_and_index $@|jq '.result'
elif [ $1 == "get_balance" ]; then
  shift
  addr=$1
  shift
  num=$1
  shift
  get_balance $(pad_hex_prefix $addr) $num $@ |jq '.result'|xargs -I {} python -c 'print "%.4f ether" % ('{}/1000000000000000000.0')'
elif [ $1 == "get_code" ]; then
  shift
  addr=$1
  shift
  get_code $(pad_hex_prefix $addr) $@|jq '.result'
elif [ $1 == "get_storage" ]; then
  shift
  addr=$1
  shift
  at=$1
  shift
  get_storage $(pad_hex_prefix $addr) $at $@|jq '.result'
elif [ $1 == "get_tx_count" ]; then
  shift
  addr=$1
  shift
  if [ "$1" == "-h" ]; then
    get_tx_count $(pad_hex_prefix $addr) $@|jq '.result'|xargs printf "%d\n"
  else
    get_tx_count $(pad_hex_prefix $addr) $@|jq '.result'
  fi
fi
