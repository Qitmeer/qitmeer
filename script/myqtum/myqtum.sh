data_dir="/data/qtum/private/" 

cli="./qtum-cli -regtest --datadir=$data_dir" 

if [ "$1" == "tx" ]; then
  $cli getrawtransaction $2 1
elif [ "$1" == "block" ]; then
  $cli getblock $($cli getblockhash $2)
elif [ "$1" == "receipt" ]; then
  $cli gettransactionreceipt $2
elif [ "$1" == "storage" ]; then
  $cli getstorage $2
else
  $cli "$@"
fi
