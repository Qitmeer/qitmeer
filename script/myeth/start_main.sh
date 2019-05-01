app_dir=/data/ethereum/mainnet
data_dir=$app_dir/data
ethash_dir=$app_dir/ethash
keystore=$app_dir/keystore
log_dir=$app_dir/log
 

#  --port value          Network listening port (default: 30303)
#  
#  --rpc                  Enable the HTTP-RPC server
#  --rpcaddr value        HTTP-RPC server listening interface (default: "localhost")
#  --rpcport value        HTTP-RPC server listening port (default: 8545)
#  --rpcapi value         API's offered over the HTTP-RPC interface
#  --ws                   Enable the WS-RPC server
#  --wsaddr value         WS-RPC server listening interface (default: "localhost")
#  --wsport value         WS-RPC server listening port (default: 8546)
#  --wsapi value          API's offered over the WS-RPC interface

verbose=3
./geth --verbosity "$verbose" --syncmode full --identity myethtest --datadir $data_dir --ethash.dagdir $ethash_dir --keystore $keystore --rpcapi eth,net,web3,txpool --rpc 


