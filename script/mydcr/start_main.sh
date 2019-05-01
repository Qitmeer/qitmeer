data_dir=/data/decred/main
network=

while [ $# -gt 0 ] ;do
  case "$1" in
    -network)
      network=$2
      shift;;
    *)
      #echo "cmd is $cmd"
      break;;
    esac
  shift
done


./dcrd -A $data_dir --txindex $@
