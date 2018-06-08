data_dir=/data/bitcoin/mainnet

#./bitcoind --datadir=$data_dir --txindex --server #--reindex  --printtoconsole
./bitcoind --datadir=$data_dir --txindex --server -connect=0  --printtoconsole --debug=net
