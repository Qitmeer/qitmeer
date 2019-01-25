app_dir=/data/btcd/private
data_dir=$app_dir/data
log_dir=$app_dir/log
debug_level=trace

mining_addr=SiMVDjmSfbR9RAvvaxoBgMA4xi36yrwer4


#  --rpclisten=    Add an interface/port to listen for RPC connections (default port: 8334, testnet: 18334)
#  --rpccert=      File containing the certificate file (rpc.cert)
#  --rpckey=       File containing the certificate key (rpc.key)

#./btcd -C=$app_dir/btcd.conf -b=$data_dir --rpclisten=127.0.0.1:18111 --rpccert=$app_dir/rpc.cert --rpckey=$app_dir/rpc.key --logdir=$log_dir --txindex --regtest


./btcd -C=$app_dir/btcd.conf -b=$data_dir --rpccert=$app_dir/rpc.cert --rpckey=$app_dir/rpc.key --logdir=$log_dir --txindex --simnet --miningaddr=$mining_addr --debuglevel=$debug_level --addrindex
