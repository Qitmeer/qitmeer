data_dir=/data/multichain/private

if [ ! -e $data_dir/mytest/params.dat ]; then
  ./multichain-util --datadir=$data_dir create mytest
fi

./multichaind mytest --datadir=$data_dir --txindex --reindex --server --printtoconsole $@
