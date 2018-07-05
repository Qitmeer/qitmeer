app_dir=/data/btcd-dag/private
debug_level=trace
jsonrpc=0.0.0.0:18554
grpc=0.0.0.0:18558

./btcwallet -A $app_dir -C $app_dir/btcwallet.conf --cafile=$app_dir/rpc.cert --simnet  --rpclisten=$jsonrpc --experimentalrpclisten=$grpc -d $debug_level "$@"


