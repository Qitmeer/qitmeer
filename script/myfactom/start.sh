data_dir=/data/factom/private
network=LOCAL
log=info

while [ $# -gt 0 ] ;do
  case "$1" in
    -network)
			network=$2
      shift;;
    -log)
      log=$2
      shift;;
    *)
      #echo "cmd is $cmd"
      break;;
    esac
  shift
done

./factomd -config $data_dir/factomd.conf -factomhome $data_dir -network $network -loglvl $log "$@"
