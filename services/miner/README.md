# Solo Mining

```shell script
Get BlockTemplate
curl -X POST \
  http://127.0.0.1:1234/ \
  -H 'authorization: Basic dGVzdDp0ZXN0' \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -H 'postman-token: a6702d2a-9deb-4d42-6fbd-51dfb3173001' \
  -d '{
  "method":"getBlockTemplate",
  "version":"2.0",
  "params":[[],8],
  "id":1
}'

{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "stateroot": "0000000000000000000000000000000000000000000000000000000000000000",
        "curtime": 1608033773,
        "height": 1,
        "blues": 1,
        "previousblockhash": "45512392e69843f98182582f4279c1745074633fe7a00fb8eb43ac143d23a9a7",
        "sigoplimit": 80000,
        "sizelimit": 1048576,
        "weightlimit": 4000000,
        "parents": [
            {
                "data": "a7a9233d14ac43ebb80fa0e73f63745074c179422f588281f94398e692235145",
                "hash": "45512392e69843f98182582f4279c1745074633fe7a00fb8eb43ac143d23a9a7"
            }
        ],
        "transactions": [],
        "version": 18,
        "coinbaseaux": {
            "flags": "092f7169746d6565722f"
        },
        "coinbasevalue": 12000000000,
        "longpollid": "45512392e69843f98182582f4279c1745074633fe7a00fb8eb43ac143d23a9a7-1608033773",
        "pow_diff_reference": {
            "nbits": "2003ffff",
            "target": "03ffff0000000000000000000000000000000000000000000000000000000000"
        },
        "maxtime": 1608034133,
        "mintime": 1547735582,
        "mutable": [
            "time",
            "transactions/add",
            "prevblock",
            "coinbase/append"
        ],
        "noncerange": "00000000ffffffff",
        "capabilities": [
            "proposal"
        ],
        "template_header": "12000000a7a9233d14ac43ebb80fa0e73f63745074c179422f588281f94398e6922351456951fd47a7985e3b78923b6f537e3891371f3b41bbf5447e142e7a7b23015e300000000000000000000000000000000000000000000000000000000000000000ffff0320eda5d85f080000000000000000"
    }
}
```

#### Use template_header and mining nonce , replace the 8 bytes end of header
#### pow_diff_reference.target is the target hash
#### submit work
- header hex is 117 header + 169 zero bytes 

| header hex (286 bytes)	| parents length hex	| parents data (union hex)	| tx length hex | tx data (union hex)	|
| ---| --- | --- | --- | --- | 
| 286 bytes	|4 bytes	|	| 4 bytes	| 	|
```shell script
curl -X POST \
  http://127.0.0.1:1234/ \
  -H 'authorization: Basic dGVzdDp0ZXN0' \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -H 'postman-token: 6b997b0f-a45f-9c4c-9d2d-224f71d79adc' \
  -d '{
  "method":"submitBlock",
  "version":"2.0",
  "params":["0c000000cf7aa78c76e17fb4d79e94c9f687fb9aa57c6dd02a226d544814f28b9e9245f1e7ad56b73c0a366b153c4d5e3223b22e584eaca910de85c4f20629ebf414c13b0000000000000000000000000000000000000000000000000000000000000000ffff001b3892595fcf8048e406000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000246ac03a4420e84c3cdec9ea7b169013449a2e806c5bbc247510b5c78d6e7718a01010000000154127065e760fb6ae0d3ec173780bad2f3006a31c2259a9f6e224e19b6556086ffffffffffffffff01007841cb020000001976a914a0d4cbddb28afa0fea20abc8a0b1516cdebf00e488ac00000000000000003892595f011903513a020800011d1d2c3f4dc70b2f7575706f6f6c2e636e2f"],
  "id":1
}
'
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": "0c034550cf7aa78c76e17fb4d79e94c9f687fb9aa57c6dd00c000000cf7ad290",
}
```