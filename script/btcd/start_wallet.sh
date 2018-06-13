app_dir=/data/btcd/private

./btcwallet -A $app_dir -C $app_dir/btcwallet.conf --cafile=$app_dir/rpc.cert --simnet "$@"

