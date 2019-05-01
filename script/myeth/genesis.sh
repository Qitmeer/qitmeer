app_dir=/data/ethereum/private
data_dir=$app_dir/data
ethash_dir=$app_dir/ethash
keystore=$app_dir/keystore


while [ $# -gt 0 ] ;do
  case "$1" in
    -datadir|--datadir)
      shift;
      data_dir=$1
      shift;;
    -f|--force)
      force="true"
      shift;;
    *)
      cmd="$@"
      # echo "cmd is $cmd"
      break;;
  esac
done

if [ "X$force" != "X" ]; then
	rm -rf $data_dir
fi

./geth -datadir $data_dir --keystore $keystore init "$cmd" 
