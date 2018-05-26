data_dir="/data/bitcoin/private/" 

cli="./bitcoin-cli -regtest --datadir=$data_dir" 

if [ "$1" == "tx" ]; then
  $cli getrawtransaction $2 1
elif [ "$1" == "block" ]; then
  $cli getblock $($cli getblockhash $2)
else
  $cli "$@"
fi
