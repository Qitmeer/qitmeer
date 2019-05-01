data_dir="/data/multichain/private/" 

cli="./multichain-cli --datadir=$data_dir mytest" 

if [ "$1" == "tx" ]; then
  $cli getrawtransaction $2 1
elif [ "$1" == "block" ]; then
  $cli getblock $($cli getblockhash $2)
else
  $cli "$@"
fi
