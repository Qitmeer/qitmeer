# nx user guide
nx 是一个命令行工具，是 bx 命令的超集，提供了各种用于密钥管理和交易构建的命令。

## Prerequisites

- Update Go to version at least 1.12 (required >= **1.12**)

Check your golang version

```bash
~ go version
go version go1.12 darwin/amd64
```

## How to build

```bash
~ mkdir -p /tmp/work
~ cd /tmp/work
~ git clone https://github.com/HalalChain/qitmeer.git
~ cd qitmeer/tools
~ go build
~ ./nx --version
Nx Version : "0.0.1"
```

## nx Commands

```bash
~ nx

Usage: nx [--version] [--help] <command> [<args>]

encode and decode :
    base58-encode         encode a base16 string to a base58 string
    base58-decode         decode a base58 string to a base16 string
    base58check-encode    encode a base58check string
    base58check-decode    decode a base58check string
    base64-encode         encode a base16 string to a base64 string
    base64-encode         encode a base64 string to a base16 string
    rlp-encode            encode a string to a rlp encoded base16 string
    rlp-decode            decode a rlp base16 string to a human-readble representation

hash :
    blake2b256            calculate Blake2b 256 hash of a base16 data.
    blake2b512            calculate Blake2b 512 hash of a base16 data.
    sha256                calculate SHA256 hash of a base16 data.
    sha3-256              calculate SHA3 256 hash of a base16 data.
    keccak-256            calculate legacy keccak 256 hash of a bash16 data.
    blake256              calculate blake256 hash of a base16 data.
    ripemd160             calculate ripemd160 hash of a base16 data.
    bitcion160            calculate ripemd160(sha256(data))
    hash160               calculate ripemd160(blake2b256(data))

entropy (seed) & mnemoic & hd & ec
    entropy               generate a cryptographically secure pseudorandom entropy (seed)
    hd-new                create a new HD(BIP32) private key from an entropy (seed)
    hd-to-ec              convert the HD (BIP32) format private/public key to a EC private/public key
    hd-to-public          derive the HD (BIP32) public key from a HD private key
    hd-decode             decode a HD (BIP32) private/public key serialization format
    hd-derive             Derive a child HD (BIP32) key from another HD public or private key.
    mnemonic-new          create a mnemonic world-list (BIP39) from an entropy
    mnemonic-to-entropy   return back to the entropy (the random seed) from a mnemonic world list (BIP39)
    mnemonic-to-seed      convert a mnemonic world-list (BIP39) to its 512 bits seed 
    ec-new                create a new EC private key from an entropy (seed).
    ec-to-public          derive the EC public key from an EC private key (the compressed format by default )
    ec-to-wif             convert an EC private key to a WIF, associates with the compressed public key by default.
    wif-to-ec             convert a WIF private key to an EC private key.
    wif-to-public         derive the EC public key from a WIF private key.

addr & tx & sign
    ec-to-addr            convert an EC public key to a paymant address. default is nox address
    tx-encode             encode a unsigned transaction.
    tx-decode             decode a transaction in base16 to json format.
    tx-sign               sign a transactions using a private key.
    msg-sign              create a message signature
    msg-verify            validate a message signature
    signature-decode      decode a ECDSA signature

```

## Encoding Commands

解码和编码 hlc 地址；

#### base58-encode

encode a base16 string to a base58 string

##### Example

```bash
~ nx base58-decode RmCYoUMqKZopUkai2YhUFHR9UeqjeyjTAgW
```

```bash
# base16 string
0df144d959afb6db4ad730a6e2c0daf46ceeb98c53a059cd6527
```

---

#### base58-decode

decode a base58 string to a base16 string

##### Example

```bash
~ nx base58-decode 1234567890abcdef
```

```bash
# base58 string
43c9JGZmRvE
```

---

#### base58check-encode

encode a base58check string,

```bash
~ nx base58check-encode
Usage: nx base58check-encode [-v <ver>] [hexstring]
  -a string
    base58check hasher
  -c int
    base58check checksum size (default 4)
  -v version
    base58check version [mainnet|testnet|privnet|btcmainnet|btctestnet|btcregressionnet] (default privnet)
```

##### Example

```bash
~ nx base58check-encode 1234567890abcdef
```

```bash
# base58 string
43c9JGZmRvE
```

---

base58check-encode    encode a base58check string
base58check-decode    decode a base58check string
