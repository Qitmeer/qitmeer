function unknown_arg() {
  local at=$1
  echo "unknown arg ($at) : $@"
}

function ssl_dgst_ripemd160() {
  local input=$1
  cmd="printf %s $input|openssl dgst -ripemd160"
  if [ "$_debug" == "true" ]; then
    echo "-------------------------------------------" 
    echo "exec_command : $cmd" 
    echo "  input      : $input"
    echo "  hash argo  : ripemd160"
    echo "exec_result  : "$(eval $cmd 2>/dev/null)
  else
    eval $cmd
  fi
}
function ssl_rand(){
  local num=32
  local output=""
  local encode="-hex"
  local rand=""
  while [[ ! -z "$1" ]];do
    case "$1" in
      -n|--num)
        shift
        num=$1	
        shift
        ;;
      -hex|-base64)
        encode=$1
        shift
        ;;
      -output|-out)
        shift
        output="-out $1"
        shift
        ;;
      -rand)
        shift
        if [ -z "$1" ]; then
          rand="-rand ~/.rnd"
        else
          local rand_file=$1
          if [ "${rand_file:0:1}" == "-" ]; then
            rand="-rand ~/.rnd"
          else
            rand="-rand $1"
          fi
        fi
        shift
        ;;
      -h|--help)
        shift
        openssl rand --help 
        exit
        ;;
      **)
        unknown_arg "ssl_rand" $@
        shift
        exit
        ;;
    esac
  done  
  cmd="openssl rand $num $encode $rand $output"
  if [ "$_debug" == "true" ]; then
    echo "-------------------------------------------" 
    echo "exec_command : $cmd" 
    echo "  rand size  : $num"
    echo "  encode     : ${encode:1}"  #remove -
    echo "  PRNG file  : ${rand:6}"    #remove -rand
    echo "exec_result  : "$(eval $cmd 2>/dev/null)
  else
    eval $cmd
  fi
}

function usage() {
  echo "myssl - my wrapper for the OpenSSL command line tool"
  echo "USAGES:"
  echo "  myssl command [ command_opts ] [ command_args ]"
  echo "EXAMPLES:"
  echo "  myssl rand"
  echo "  myssl rand --help"
  echo "  myssl -h|--help"
}

function main() {
  while [[ ! -z "$1" ]];do
    case "$1" in
      -h|--help)
        shift
        usage
        ;;
      -D|--debug)
        shift
        _debug="true"
        ;;
      rand)
        shift
        ssl_rand $@
        exit
        ;;
      ripemd160)
        shift
        ssl_dgst_ripemd160 $@
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
