# Qitmeer Testnet

## Qitmeer Internal Testnet

### v0.8.2 2019/11/25 (Planed)

| Internal Testnet   |     Info             |
| -------------------|----------------------|
| **Testnet Type**   | Internal             |
| **Version**        | *0.8.2*              |
| **Source**         | |
| **Start Date**     | 2019/11/25 (Planned) |
| **End Date**       | |
| **blockOne Hash**  | UNKNOWN |
| **Data CleanUp**   | YES |
| **CleanUP Reason** | community mining reward |
| **Ledger recovery**| YES|
| **Remarks**        | txsvaild test finished & the community mining reward|

#### Reward Plan

##### The 100 block number from 15800 to 30000

https://www.random.org/integer-sets/?sets=1&num=100&min=15800&max=30000&seqnos=on&commas=on&sort=on&order=index&format=plain&rnd=id.2b9a36bb6fd1fa47a19cf822d93eca7df0a9f5000937e2ab87eb30b2b803c4d7

```
15978, 16375, 16445, 16988, 17167, 17264, 17273, 17506, 17524, 17553,
17567, 17570, 17575, 17612, 17774, 18017, 18120, 18299, 18434, 18460,
18539, 18584, 18760, 18952, 19071, 19103, 19396, 19703, 19708, 19881,
20098, 20116, 20252, 20637, 20663, 21014, 21073, 21109, 21286, 21380,
21558, 21623, 21923, 22004, 22041, 22511, 22748, 23296, 23568, 23847,
23959, 24120, 24209, 24229, 24256, 24279, 24290, 24464, 24593, 24700,
24804, 24808, 24855, 24897, 25060, 25157, 25197, 25662, 25747, 25815,
25822, 25912, 25945, 26264, 26273, 26506, 26559, 26635, 26813, 26912,
27492, 27564, 27743, 27762, 27767, 27880, 28138, 28520, 28618, 28810,
28985, 29172, 29214, 29504, 29528, 29696, 29726, 29784, 29821, 29991
```

##### The Ledger recovery

Genesis block will send 130 PMEER to the owner of coinbase address of the 100 Blocks of the previous internal testnet (0.8.0&0.8.1)


### v0.8.1 2019/11/20 (Planed)

| Internal Testnet   |     Info             |
| -------------------|----------------------|
| **Testnet Type**   | Internal             |
| **Version**        | *0.8.1*              |
| **Source**         | |
| **Start Date**     | 2019/11/20 (Planned) |
| **End Date**       | |
| **blockOne Hash**  | 2b9a36bb6fd1fa47a19cf822d93eca7df0a9f5000937e2ab87eb30b2b803c4d7|
| **Data CleanUp**   | NO |
| **CleanUP Reason** | NO |
| **Ledger recovery**| NO NEED|
| **Remarks**        | the destroy finished |

#### HLC Token destroyed result

| Token Destroyed Result  |     Info                     |
| ------------------------|------------------------------|
| Sum of Destroyed        | **200287911**                |
| Sum of 0x00...00 Holder | [200287911][etherscan2]      |
| Tx                      | 1266                         |
| Destroyed Tx            | 633                          |
| Token Holder            | 580                          |
| ALL HLC Hodlers         | [1745][hlc]                  |
| *Result mirror 1*       | [qitmeer.io](https://activity.qitmeer.io/) |
| *Result mirror 2*       | [etherscan.io][etherscan1]|
| *Result mirror 3*       | [hlc-token-destroyed.csv](./hlc-token-destroyed.csv)|

```bash
cat hlc-token-destroyed.csv |grep 0x000000000|cut -d, -f 7|awk '{print substr($0,2,(length($0)-3))}'|python -c "import sys; print(sum(int(l) for l in sys.stdin))"
200287911
cat hlc-token-destroyed.csv |wc -l
    1266
cat hlc-token-destroyed.csv |grep 0x000000000|wc -l
     633
cat hlc-token-destroyed.csv |grep -v 0x000000000|cut -d, -f 5|uniq|wc -l
     580
```

[hlc]:https://etherscan.io/token/0x58c69ed6cd6887c0225d1fccecc055127843c69b
[etherscan1]:https://etherscan.io/token/0x58c69ed6cd6887c0225d1fccecc055127843c69b?a=0x126720ec10f5afbf2184146621f183cae317f573
[etherscan2]:https://etherscan.io/token/0x58c69ed6cd6887c0225d1fccecc055127843c69b?a=0x0000000000000000000000000000000000000000

### v0.8.0 2019/11/14

| Internal Testnet   |     Info             |
| -------------------|----------------------|
| **Testnet Type**   | Internal             |
| **Version**        | *0.8.0*              |
| **Source**         | https://github.com/Qitmeer/qitmeer/tree/v0.8.0 |
| **Start Date**     | 2019-11-14T08:07:32Z |
| **End Date**       | |
| **blockOne Hash**  | 2b9a36bb6fd1fa47a19cf822d93eca7df0a9f5000937e2ab87eb30b2b803c4d7|
| **Data CleanUp**   | YES |
| **CleanUP Reason** | fix [#173](https://github.com/Qitmeer/qitmeer/pull/173)|
| **Ledger recovery**| NO  |

#### TxsVaild Status

| TxsVaild | BlockOrder | Hash |
| --------| -----------| ---- |
| false   | 756  | [329faede8548ef08ac8b22920a2d9ff7d6a4e4c11f1d04c34ff9f6e5d9b6a867][756]|
| false   | 1015 | [0285c30c547317b934481a8dcd43e376c24a7e859abb211462ab32e78aa81ec7][1015] |
| false   | 1026 | [0eb5fd92286feef371cf66a6b1bc51f29d8d02e005cd04f49f291227b87c0dc9][1026] |
| false   | 1035 | [104060793e7c4242f7f478e310b353f5d1d9d1636f2f2a07ffb01ffe5b1f2dfb][1035] |

[756]:https://explorer.qitmeer.io/block/329faede8548ef08ac8b22920a2d9ff7d6a4e4c11f1d04c34ff9f6e5d9b6a867
[1015]:https://explorer.qitmeer.io/block/0285c30c547317b934481a8dcd43e376c24a7e859abb211462ab32e78aa81ec7
[1026]:https://explorer.qitmeer.io/block/0eb5fd92286feef371cf66a6b1bc51f29d8d02e005cd04f49f291227b87c0dc9
[1035]:https://explorer.qitmeer.io/block/104060793e7c4242f7f478e310b353f5d1d9d1636f2f2a07ffb01ffe5b1f2dfb


### ~v0.7.9~

### v0.7.8
| Internal Testnet   |     Info             |
| -------------------|----------------------|
| **Testnet Type**   | Internal             |
| **Version**        | *0.7.8*              |
| **Source**         | https://github.com/Qitmeer/qitmeer/tree/v0.7.8 |
| **Start Date**     | 2019-11-02T17:55:58+08:00 |
| **End Date**       | 2019-11-14 |
| **blockOne Hash**  | 30f056ed3c180735ba77f70a8f7134a1a913c9f4a0cf2d266a2f41f6d33e47e9|
| **Data CleanUp**   | YES |
| **CleanUP Reason** | fix [#163](https://github.com/Qitmeer/qitmeer/pull/163)|
| **Ledger recovery**| NO  |

