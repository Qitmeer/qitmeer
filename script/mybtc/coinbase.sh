./btc.sh block $1|jq -r .tx[0]|xargs ./btc.sh tx
