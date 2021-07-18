## The Qitmeer Burn Address Generateor
- The tool which generate a valid `qitmeer-base58check` encoded 
  address for the specified network (the default is testnet).
- **Security Note**: 
  - The template need to be long enough to remain the strong security.
  recommend at least 16 words, if you are not sure please keep using 
  the default values
  - See https://en.bitcoin.it/wiki/Vanitygen for the details.

### usage
```
$ ./burn --help
Usage of ./burn:
-n string
network [mainnet|testnet|mixnet|privnet] (default "testnet")
-t string
template (default "TmQitmeerTestNetBurnAddress")
```

### generate a burn address for testnet (default)

```
$ ./burn
template = TmQitmeerTestNetBurnAddress
    addr = TmQitmeerTestNetBurnAddressLn4jhCih
```
### for a specified network

using `-n` option for a network. 
```
$ ./burn -n mixnet
template = XmQitmeerMixnetBurnAddress
    addr = XmQitmeerMixnetBurnAddressgqNYbsmqv
```
