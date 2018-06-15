app_dir=/data/ethereum/private
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

./geth --verbosity "6" --syncmode full --identity myethtest --datadir $data_dir --ethash.dagdir $ethash_dir --keystore $keystore --rpcapi eth,net,web3,miner,debug,personal,txpool --rpc --rpcport 8549  --port 60606 --nodiscover 
