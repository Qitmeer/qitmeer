data_dir=/data/decred/private
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


./dcrd -A $data_dir --txindex --simnet --miningaddr SsmmJgfYDKHTh5JHXxEUHTC7Ddeos8dbdPT $@
