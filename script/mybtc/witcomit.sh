./btc.sh block $1|jq .tx[0]|xargs ./btc.sh tx|jq -r .vout[1].scriptPubKey.hex|sed s/6a24aa21a9ed//
