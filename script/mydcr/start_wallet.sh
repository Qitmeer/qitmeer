basepath=$(pwd)

./dcrwallet -A $basepath --cafile=$basepath/rpc.cert --simnet $@ 
