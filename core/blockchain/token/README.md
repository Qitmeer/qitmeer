# Qitmeer Token

## Prerequisites
For the convenience of our explanation, we take the private net as an example.
```
    ./qitmeer --privnet
    cd ./script
```
## How to create a new token ?

```
    ./cli.sh createTokenRawTx new [CoinId] [CoinName] [TokenOwnersAddress] [UpLimit]
    ./cli.sh txSign fff2cefe258ca60ae5f5abec99b5d63e2a561c40d784ee50b04eddf8efc84b0d [RawTxHex]  
```

* `RawTxHex` is results of the previous step.
* `fff2...` is private key from [testwallet](https://github.com/Qitmeer/qitmeer/blob/927cd48a0b6336efee34fb3679cae9bc4e2a8567/testutils/testwallet.go#L30)
* Please be careful not to use `txSign` in a formal scenario because it is not safe. Be sure to use [qitmeer-wallet](https://github.com/Qitmeer/qitmeer-wallet) in formal cases.

```
    ./cli.sh sendRawTx [SignedRawTxHex]
    ./cli.sh generate 1
```
* Finally, you can query the token status of the Qitmeer node.
```
    ./cli.sh tokeninfo
```

## How to renew a token ?
```
    ./cli.sh createTokenRawTx renew [CoinId] [CoinName] [TokenOwnersAddress] [UpLimit]
    
    See above...
    ...
```

## How to validate a token ?
```
    ./cli.sh createTokenRawTx validate [CoinId] 
    
    See above...
    ...
```

## How to invalidate a token ?
```
    ./cli.sh createTokenRawTx invalidate [CoinId]
    
    See above...
    ...
```