# MeerXKeccakV1

# Meer Header

| version	| parent_root	| tx_root	| state_root | difficulty 	| time | pow_type | nonce 
| ---| --- | --- | --- | --- | --- |  --- |  --- |
| 4 bytes	|32 bytes	|32 bytes	| 32 bytes	| 4 bytes	| 4 bytes | 1 byte |8 bytes 

### QitmeerKeccak256(NewLegacyKeccak512(NewLegacyKeccak512(header))^1)
#### [MeerXKeccakV1 See Golang Code](https://github.com/Qitmeer/qitmeer/blob/meer_pow/common/hash/hashfuncs_meer_crypto.go)
#### [QitmeerKeccak256 Use PaddingFix See C Code](https://github.com/jamesvan2019/keccakhash_c/commit/68cd0af8e573eafd2adeab1747e1760cbec99cf3)
#### [QitmeerKeccak256 Use PaddingFix See Golang Code](https://github.com/Qitmeer/crypto/blob/master/sha3/hashes.go#L76)

### NewLegacyKeccak512 Use Standard Keccak512

### Example
```golang
input : 117 bytes
HashMeerXKeccakV1(helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworldhel)
result: 046bb4ee0e487afb53c428f5d18f2875951f80330fe39870011ddac6a9c06b1c

```
### Header assemble example
```golang
# example 1
header(117 bytes):12000000c24900a189710157b8d020f2d09068ec4ee1eacae6b38753a6304894d0e2104b80eb912fbeb3b6735a4acfb89fd199e0f0abc50a48dccd52e2245e0619755468000000000000000000000000000000000000000000000000000000000000000068db061fda20d75f088f10000000000000
version : 12000000
parent_root :c24900a189710157b8d020f2d09068ec4ee1eacae6b38753a6304894d0e2104b
tx_root :80eb912fbeb3b6735a4acfb89fd199e0f0abc50a48dccd52e2245e0619755468
state_root :0000000000000000000000000000000000000000000000000000000000000000
difficulty : 68db061f
time :da20d75f
pow_type :08
nonce :8f10000000000000(4239)

target: 
0006db6800000000000000000000000000000000000000000000000000000000

meerHash:84499238afb6b544e01032f2ea73d5692ef5d29775daa3a778bfbd2efbb90000

meerHash(reverse):0000b9fb2ebdbf78a7a3da7597d2f52e69d573eaf23210e044b5b6af38924984

Compliance target difficulty

# example 2
header(117 bytes):120000005b42f3a292059337e2097c4077d6578adf4253c63ab74a51d57a3cf330003267b012b719ebfb86d8ea7864307bd0b743980ead38d1f2c06919fc45d6eb604ed9000000000000000000000000000000000000000000000000000000000000000068db061fa421d75f086605000000000000
version : 12000000
parent_root :5b42f3a292059337e2097c4077d6578adf4253c63ab74a51d57a3cf330003267
tx_root :b012b719ebfb86d8ea7864307bd0b743980ead38d1f2c06919fc45d6eb604ed9
state_root :0000000000000000000000000000000000000000000000000000000000000000
difficulty : 68db061f
time :a421d75f
pow_type :08
nonce :6605000000000000(1382)

target: 
0006db6800000000000000000000000000000000000000000000000000000000

meerHash:3b94fab4d25980aca4512b86b32bc96fd7cb8f8fdf3f265ffde45f72688e0100

meerHash(reverse):00018e68725fe4fd5f263fdf8f8fcbd76fc92bb3862b51a4ac8059d2b4fa943b

Compliance target difficulty
```