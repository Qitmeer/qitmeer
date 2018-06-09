data_dir=/data/bitcoin/testnet

# ./bitcoind -testnet --datadir=$data_dir --txindex --server  --printtoconsole --debug=net
# ./bitcoind -testnet --datadir=$data_dir --txindex --printtoconsole
./bitcoind -testnet --datadir=$data_dir --txindex --printtoconsole --connect=0 --debug=net

