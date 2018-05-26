data_dir=/data/bitcoin/private

./bitcoind --datadir=$data_dir --txindex --reindex --server --regtest --printtoconsole
