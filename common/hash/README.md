# MeerCrypto

# Meer Header

| version	| parent_root	| tx_root	| state_root | difficulty 	| time | pow_type | nonce 
| ---| --- | --- | --- | --- | --- |  --- |  --- |
| 4 bytes	|32 bytes	|32 bytes	| 32 bytes	| 4 bytes	| 4 bytes | 1 byte |8 bytes 

### QitmeerKeccak256(NewLegacyKeccak512(NewLegacyKeccak512(header))^1)
#### [MeerCrypto See Golang Code](https://github.com/jamesvan2019/meer/blob/meer_pow/common/hash/hashfuncs_meer_crypto.go)
#### [QitmeerKeccak256 Use PaddingFix See C Code](https://github.com/jamesvan2019/keccakhash_c/commit/68cd0af8e573eafd2adeab1747e1760cbec99cf3)
#### [QitmeerKeccak256 Use PaddingFix See Golang Code](https://github.com/Qitmeer/crypto/blob/master/sha3/hashes.go#L76)

### NewLegacyKeccak512 Use Standard Keccak512

### Example
```golang
input : 117 bytes
HashMeerCrypto(helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhel)
result: 250bcbab5c6959fd0325c513ff1e95e05ef1bbbd39214a050b44e52046573b5a

```
### Header assemble example
```golang
# example 1
header(117 bytes):120000009cbbf9987538e02437af518ee0e262e1381e6cd075b579ea08a2ba95242d5dfe740a01155aa4ae1fec77c127d19d0c93ccf745d879af54bcb0c4724d04d4570e0000000000000000000000000000000000000000000000000000000000000000ffff03209638d65f08e600000000000000
version : 12000000
parent_root :9cbbf9987538e02437af518ee0e262e1381e6cd075b579ea08a2ba95242d5dfe
tx_root :740a01155aa4ae1fec77c127d19d0c93ccf745d879af54bcb0c4724d04d4570e
state_root :0000000000000000000000000000000000000000000000000000000000000000
difficulty : ffff0320
time :9638d65f
pow_type :08
nonce :e600000000000000(230)

target: 
03ffff0000000000000000000000000000000000000000000000000000000000

meerHash:fbde247d9418ad7886572e2e90c76ed05ac58b0674088ca2b9596e9479af1003

meerHash(reverse):0310af79946e59b9a28c0874068bc55ad06ec7902e2e578678ad18947d24defb

Compliance target difficulty

# example 2
header(117 bytes):12000000352e2843a0536e80a2af770b87be244ea0d007f368e57c05dcbf58c1d98f3a95b47b480124466fe5088a3490a8729e3f0c57a85204ad2ab5ab502a6a6d46a69a0000000000000000000000000000000000000000000000000000000000000000d0b3061f283ad65f08f12b000000000000
version : 12000000
parent_root :352e2843a0536e80a2af770b87be244ea0d007f368e57c05dcbf58c1d98f3a95
tx_root :b47b480124466fe5088a3490a8729e3f0c57a85204ad2ab5ab502a6a6d46a69a
state_root :0000000000000000000000000000000000000000000000000000000000000000
difficulty : d0b3061f
time :283ad65f
pow_type :08
nonce :f12b000000000000(11249)

target: 
0006b3d000000000000000000000000000000000000000000000000000000000

meerHash:5274d9733b48a7513a57950c59ee897106bdcc8313ea60b777b1c3ad739d0500

meerHash(reverse):00059d73adc3b177b760ea1383ccbd067189ee590c95573a51a7483b73d97452

Compliance target difficulty
```