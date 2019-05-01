# See 
#   https://github.com/iotaledger/iri/blob/dev/src/main/java/com/iota/iri/service/API.java
#   https://github.com/iotaledger/iota.lib.js/blob/dev/lib/api/apiCommands.js
#   https://github.com/iotaledger/iota.lib.go/blob/master/api.go
# for API details
#   addNeighbors
#   attachToTangle
#   broadcastTransactions
#   findTransactions
#   getBalances
#   getInclusionStates
#   getNeighbors
#   getNodeInfo
#   getTips
#   getTransactionsToApprove
#   getTrytes
#   interruptAttachingToTangle
#   removeNeighbors
#   storeTransactions
#   getMissingTransactions

function unknown_arg() {
  local at=$1 
  echo "unknown arg ($at) : $@"
}

function nodeinfo() {
  local data='{"command":"getNodeInfo"}' 
  get_result "$data"
}
function nbinfo() {
  local data='{"command":"getNeighbors"}' 
  get_result "$data"
}
function addnb() {
  local data='{"command":"addNeighbors"}'
  get_result "$data"
}
function tips() {
  local data='{"command":"getTips"}'
  get_result "$data"
}

function get_result(){
  local host=${host:-localhost}
  local port=${port:-14600}
  local data=$1
  cmd="curl -s -X POST -H 'Content-Type: application/json' -H 'X-IOTA-API-Version: 1.4' --data $data http://$host:$port"
  if [ "$_debug" == "true" ]; then
    echo "{" 
    echo "  command : $cmd" 
    echo "  payload : $data"
    echo "}"
  fi
  result=$(eval $cmd)
  if [ "$_debug" == "true" ]; then
    echo $result
  else
    echo $result|jq .
  fi
}


function usage() {
  echo "myiota - my bash helper for the iota json api" 
  echo "USAGES:"
  echo "  myiota command [ command_opts ] [ command_args ]"
  echo "EXAMPLES:"
  echo "  myiota nodeinfo"
}

function main() {
  while [[ ! -z "$1" ]];do
    case "$1" in
      -h|--help)
        shift
        usage
        ;;
      -p|--port)
        shift
        port=$1
        shift
        ;;
      -H|--host)
        shift
        host=$1
        shift
        ;;
      -D|--debug)
        shift
        _debug="true"
        ;;
      nodeinfo|getNodeInfo)
        shift
        nodeinfo $@
        #echo $result|jq .
        exit
        ;;
      nbinfo|getNeighbors)
        shift
        nbinfo $@
        exit
        ;;
      addnb|addNeighbors)
        shift
        addnb $@
        exit
        ;;
      tips|getTips)
        shift
        tips $@
        exit
        ;;
      **)
        unknown_arg "main" $1
        shift
        usage
        exit
        ;;
    esac
  done  
}
if [ -z "$1" ];then
  usage
else
  main $@
fi
