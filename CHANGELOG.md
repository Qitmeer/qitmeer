# Changelog

## [0.10.5] - 2021-09-06
- Optimizition for the mempool sync.
- Bug fixes

## [0.10.4] - 2021-08-25
- Consensus upgrading
  - BlockDAG consensus improvement of the ability to process the massive concurrent blocks.
  - Accordingly, The miner protocol changes to limit the empty block generation.

## [0.10.3] - 2021-08-17
- A lot of optimization & bug fixes for the P2P network and the DAG synchronization.
- upgrade the network protocol version to 33. 


## [0.10.2] - 2021-07-28
- update network protocol to support multiple-network.

## [0.10.1] - 2021-07-25
- Change 0.10.x testnet addr prefix to distinguish 0.9.x testnet
- Optimize network friendly information

## [0.10.0] - 2021-07-24

- Replaced entire new P2P networking layer.
  - Adapt to the `libp2p` code base.
  - Entirely rewritten the block synchronization machanism.
  - Obsoleted the seeder machanism with the bootstrap node.
  - Add new node type (relay node) to improve the local network connectivity.
- Changed to a new Pow algorithm.
  - The `meer_xkeccak`, a new ASCI friendly PoW algorithm.
- Non-std transaction support.
  - Uxto based layer 2 token issue and management mechanism.
  - The customized tx fee support.
  - The customized tx time lock mechanism.
- New consensus deployment mechanism.
- Many improvements on the block DAG consensus.
- A lot of bug fixes overall.

## [0.9.2] - 2020-09-01

- Pmeer soft fork for pow proportion upgrade.
  [#361](https://github.com/Qitmeer/qitmeer/pull/361)

- Optimized Dag, Optmize maturity and blue for block DAG
  [#360](https://github.com/Qitmeer/qitmeer/pull/360)
  [#359](https://github.com/Qitmeer/qitmeer/pull/359)
  [#348](https://github.com/Qitmeer/qitmeer/pull/348)

- Added support of new option '--cacheinvalidtx', improve database compatibility with older versions
  [#357](https://github.com/Qitmeer/qitmeer/pull/357)
  [#356](https://github.com/Qitmeer/qitmeer/pull/356)
  [#354](https://github.com/Qitmeer/qitmeer/pull/354)
  [#352](https://github.com/Qitmeer/qitmeer/pull/352)
  [#350](https://github.com/Qitmeer/qitmeer/pull/350)
  [#349](https://github.com/Qitmeer/qitmeer/pull/349)
  [#343](https://github.com/Qitmeer/qitmeer/pull/343)
  [#339](https://github.com/Qitmeer/qitmeer/pull/339)

- Improved GBT performance tremendously
  [#345](https://github.com/Qitmeer/qitmeer/pull/345)
  [#351](https://github.com/Qitmeer/qitmeer/pull/351)

- Improved 'fastibd' tool
  [#359](https://github.com/Qitmeer/qitmeer/pull/359)
  [#358](https://github.com/Qitmeer/qitmeer/pull/350)

- Fixed Bug: STXO calculation error
  [#338](https://github.com/Qitmeer/qitmeer/pull/338)

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


[0.9.2]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.9.2-release
[0.9.1]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.9.1-release
[0.9.0]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.9.0-release
[0.8.5]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.5
[0.8.4]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.4.1
[0.8.3]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.3.1
[0.8.2]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.2
[0.8.1]: https://github.com/Qitmeer/qitmeer/releases/tag/v0.8.1
