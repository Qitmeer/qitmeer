# Changelog

## [0.9.1] - 2020-07-14

- Optimized BlockDag network synchronization.
  [#334](https://github.com/Qitmeer/qitmeer/pull/334)
  [#332](https://github.com/Qitmeer/qitmeer/pull/332)
  [#325](https://github.com/Qitmeer/qitmeer/pull/325)
  [#324](https://github.com/Qitmeer/qitmeer/pull/324)
  [#323](https://github.com/Qitmeer/qitmeer/pull/323)
- Fixed CI test script to support the latest qitmeer build.
  [#336](https://github.com/Qitmeer/qitmeer/pull/336)
- Added New RPC: SetLogLevel.
  [#335](https://github.com/Qitmeer/qitmeer/pull/335)
- Added New RPC: GetBlockByNum RPC to obsolete GetBlockByID RPC.
  [#331](https://github.com/Qitmeer/qitmeer/pull/331)
- Added New RPC: GetBlockV2.
  [#329](https://github.com/Qitmeer/qitmeer/pull/329)
- Improved API Doc.
  [#333](https://github.com/Qitmeer/qitmeer/pull/333)
- Improved RPC: GetTransaction json result transactionfee.
  [#330](https://github.com/Qitmeer/qitmeer/pull/330)
- Replaced mixnet seeder.
  [#326](https://github.com/Qitmeer/qitmeer/pull/326)
- Fixed Bug: The RPC API results in wrong hash for some coinbase transactions.
  [#327](https://github.com/Qitmeer/qitmeer/pull/327)

## [0.9.0] - 2020-06-24

- Qitmeer official public testnet : the Medina network 2.0 release.

## [0.8.5] - 2019-12-30

- Qitmeer official public testnet : the Medina network release.

## [0.8.4] - 2019-12-17

- Fixed for 0.8.4 testnet consensus of 3 days adjustment for cuckaroo 29 edge-bits.
- Fixed the RPC out-of-service by deadlock issue.
- Improved node sync performance dramatically.
- Fixed the 0.8.3 DAG consensus issue.
- POW difficulty control for cuckaroo 24 and 29.
- Added new difficulty command in qx.
- Added 0.8.3 testnet reward.

## [0.8.3] - 2019-12-13

- Added Windows node start script (for qitmeer-0.8.3-windows-amd64.cn.zip only).
- Added "qitmeer-0.8.3_checksum.txt" checksum file.
- Optimized v0.8.3 performance.
- Optimized DAG sync.
- Fixed the sync issue in v0.8.2.

## [0.8.2] - 2019-11-29

  - [#195] Remove node stop() rpc from the public API
  - [#191] Rewrite the build/relase script for qitmeer.
  - [#189] Optimize DAG anticone size for automatic calculation (Anticone size change to 6)
  - [#187] Block time reduce to 15s,  coinbase reward adjusts to 65 PMEER.
  - [#179] Geneises block includs 165 block rewards & the token destroyed amount.
  - [#186] Optimize orphans block handing
  - [#186] Optimize meg handing for mempool
  - [3b80] Optimize Block Synchronization
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


[0.9.1]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.9.1-release
[0.9.0]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.9.0-release
[0.8.5]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.5
[0.8.4]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.4.1
[0.8.3]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.3.1
[0.8.2]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.2
[0.8.1]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.1
