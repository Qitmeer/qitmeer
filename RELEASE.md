# Changelog

## [0.8.2] - 2019-11-29

### 0.8.2 Download

- MacOS   [qitmeer-0.8.2-darwin-amd64.tar.gz][0.8.2-m]
- Linux   [qitmeer-0.8.2-linux-amd64.tar.gz][0.8.2-l]
- Windows [qitmeer-0.8.2-windows-amd64.zip][0.8.2-w]

[0.8.2-m]:https://github.com/Qitmeer/qitmeer/releases/download/v0.8.2/qitmeer-0.8.2-darwin-amd64.tar.gz
[0.8.2-l]:https://github.com/Qitmeer/qitmeer/releases/download/v0.8.2/qitmeer-0.8.2-linux-amd64.tar.gz
[0.8.2-w]:https://github.com/Qitmeer/qitmeer/releases/download/v0.8.2/qitmeer-0.8.2-windows-amd64.zip

#### 0.8.2 SHA512 Checksum

```
5ab4b3fd94eb1252a31083e229b31cbf60070648740c74e9dd2d77e2800d0873febd3e8d1e6dab8466e0fc2aace3c60e8973b1ad6946a03ab5ccba3a803fb854  build/release/darwin/amd64/bin/qitmeer
71af9b5b1d2e959c6b345aecb430395f298e677527870604ea502a48ac978fd37255fe7de0936dbc28e16622a53d8251348b4d9c4ef79ea9ee72345eef899a89  build/release/linux/amd64/bin/qitmeer
653defb7eff780772e88e000772815f7e61e9167ea67e7afae91cffb49a52758916aab050ce56f59bbf06304f5bf0db21636d80a964a6971fd53334d621a3751  build/release/windows/amd64/bin/qitmeer.exe
d64c7aa5e64874a5bca7cd8e2b1fd65b75bae6cedaec72448e02ceba06992917ec74227a8398d89f51298689af39298c98c844115ac042262a19edc95d1e4dec  qitmeer-0.8.2-darwin-amd64.tar.gz
586e627d1ddb1db23a1d3914152d4be8b1d85296a040e6faab474a1fe0a5c6aa202067fe9e7710a4d55ca5cce9d070d81aa3adb90e59daa6841507e11be83aef  qitmeer-0.8.2-linux-amd64.tar.gz
019392b5a9deb97ac566140352ad317e2b4d5f4562d47f5ce465dce9ef68743a7581ddb22b867e85958b67534156290fb7ee9fc1f412bc104b7c25464903a920  qitmeer-0.8.2-windows-amd64.zip
```

### Changes & Features
  - [#195] Remove node stop() rpc from the public API
  - [#191] Rewrite the build/relase script for qitmeer.
  - [#189] Optimize DAG anticone size for automatic calculation (Anticone size change to 6)
  - [#187] Block time reduce to 15s,  coinbase reward adjusts to 65 PMEER. 
  - [#179] Geneises block includs 165 block rewards & the token destroyed amount.
  - [#186] Optimize orphans block handing
  - [#186] Optimize meg handing for mempool
  - [3b80] Optimize Block Synchronization
  
### Bug Fixes
  - [#196] qitmeer crashes on launching while receiving miner connection
  - [703d] no minning address config file raise error for the CPU miner
  - [#185] fix wrong getNodeInfo getDifficultyRatio method display

[703d]:https://github.com/Qitmeer/qitmeer/commit/703d4f5bed11379c737acff2a1d1431fb2d5a989
[3b80]:https://github.com/Qitmeer/qitmeer/commit/3b804be0026f4ee9bffdd47f445c2afee49772ec
[#185]:https://github.com/Qitmeer/qitmeer/pull/185  
[#186]:https://github.com/Qitmeer/qitmeer/pull/186
[#187]:https://github.com/Qitmeer/qitmeer/pull/187
[#189]:https://github.com/Qitmeer/qitmeer/pull/189  
[#179]:https://github.com/Qitmeer/qitmeer/pull/179
[#180]:https://github.com/Qitmeer/qitmeer/pull/180
[#191]:https://github.com/Qitmeer/qitmeer/pull/191
[#195]:https://github.com/Qitmeer/qitmeer/pull/195
[#196]:https://github.com/Qitmeer/qitmeer/pull/196


## [0.8.1] - 2019-11-20

[0.8.2]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.2
[0.8.1]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.1
