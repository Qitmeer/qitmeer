# Qitmeer Testnet Mining Reward

## Table of Contents
   * [0.8.3 Testnet Reward](#083-testnet-reward)
      * [0.8.3 rewards](#083-rewards)
   * [0.8.0 &amp; 0.8.1 Testnet Reward](#080--081-testnet-reward)
      * [The 162 Mined Blocks List](#the-162-mined-blocks-list)
      * [Reward Ranking](#reward-ranking)

## 0.8.3 Testnet Reward

### 0.8.3 rewards

for the block mined from order `38000` to `110000`, if block is valid (`txvalid=true`),  and valid minded block `count` > 30 , the reward is `(count%30) * 65` PMEER.
see (https://github.com/Qitmeer/083testnet-data) for details

```bash
$ ./check-block-reward.sh -l
TmgWMUZtDcBe36bxXG8FBCu8C3KFuy3F9jC 10403 22490 PMEER = (10403/30) * 65
TmPK4WJRLqPjhjGEpVmofZ19apjD2rsnSoF 8405 18200 PMEER = (8405/30) * 65
TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws 8206 17745 PMEER = (8206/30) * 65
TmUBfuZVNB4DemaPaqVsLofYiJJjbbf6226 6252 13520 PMEER = (6252/30) * 65
Tmc87uCQXqUaa7UPA6ESwe7P9L6ZV921UAh 5022 10855 PMEER = (5022/30) * 65
TmUquGBdL1PgCuq7t29yJQ6DLC2teb1U8AS 2855 6175 PMEER = (2855/30) * 65
TmcwnXt3d3bMxp7p2KY4jWJnZ3RALX8cABH 2433 5265 PMEER = (2433/30) * 65
TmjM2Vrf8tV3hmpKLkqyd5cNyL5JLisDSm9 2263 4875 PMEER = (2263/30) * 65
TmRZH8a4nQhBkfYnq28x1HdTNQ5wMJLWPh4 2102 4550 PMEER = (2102/30) * 65
TmYL7UUoFQNkZedhYWP4SbhtHtYVzgLjmis 2034 4355 PMEER = (2034/30) * 65
TmTr6FyJeBWq8LHdj1ZbH84nKxDiz3tQGJG 1535 3315 PMEER = (1535/30) * 65
TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj 1467 3120 PMEER = (1467/30) * 65
TmhBGnH35LSwPecZ9tCqL88UTfo28Q3PUi6 1450 3120 PMEER = (1450/30) * 65
TmYphgQo9AtCAoyeZqFMY7qRdkJTtSHh2TD 1366 2925 PMEER = (1366/30) * 65
TmdynELSnV85MivyKsHWGjh3d6gSyDxjooD 1297 2795 PMEER = (1297/30) * 65
TmRXN84v3XB2E1HZmCuTSHgCeYRs3MV7q9A 1246 2665 PMEER = (1246/30) * 65
TmZkkmz8t1KR8uvvnuHcXr7LMXs1AArTbHX 1132 2405 PMEER = (1132/30) * 65
Tmcubtx2XRRQMoBa4Cb2LHJhWnR1TLXEEjB 910 1950 PMEER = (910/30) * 65
TmUWxxq66VCdAJtf5kxPVGwHN8LzXyqRk8L 908 1950 PMEER = (908/30) * 65
TmUvv3cYB4TDtqSo47kKBvkvGn7JjgWwuGe 816 1755 PMEER = (816/30) * 65
TmQ67Hr2hfHpYEU4ThKPRnVE62h8aBMY3DT 769 1625 PMEER = (769/30) * 65
TmfEem5qkmqggTFBshbwPyTTqNERgbD4EPn 767 1625 PMEER = (767/30) * 65
TmkuVRGjxor5iNJKThmkoRMKe4cBX8DRc6e 737 1560 PMEER = (737/30) * 65
Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7 701 1495 PMEER = (701/30) * 65
TmRxANuoKN9xPvtAaLe1SA77B279pQC9Wbo 692 1495 PMEER = (692/30) * 65
TmhSc9QZsruv8PCeDgBsY3ew5EPbrgUrAhY 690 1495 PMEER = (690/30) * 65
TmjTooPeHr27TLkJzvXM9NabbyTEqXY2Bay 670 1430 PMEER = (670/30) * 65
TmcWTAY3mM7pQEzsWudtRnZNmHuWNgdhfeC 651 1365 PMEER = (651/30) * 65
Tmj8PCEms1fAT9SYEWuxfHjbppataAM7rP2 444 910 PMEER = (444/30) * 65
Tmgw6ttQsrWRsWfXPw6BgWRcviLarZUca48 422 910 PMEER = (422/30) * 65
Tmgi1VwHrAG9bb9hXSeEUv76uchPQGgt5pj 396 845 PMEER = (396/30) * 65
TmjpmECxxekaLjZA61f6T32cmmPmvjUr7iD 379 780 PMEER = (379/30) * 65
TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5 369 780 PMEER = (369/30) * 65
TmRzxNU7QxxSq9ErvWPZ9bjqABchVehSYUL 338 715 PMEER = (338/30) * 65
TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW 304 650 PMEER = (304/30) * 65
TmUtMdS8gd6QCCVeQvJ3kk1Ah28c59vw2SE 284 585 PMEER = (284/30) * 65
TmZjUjatt9yixCxWX7bwyB2DbM8PpPWSLPb 257 520 PMEER = (257/30) * 65
TmkVvov4MEgCLKjty4D5zFevy4cftaiRwCW 205 390 PMEER = (205/30) * 65
Tmin7cp5Bn2bTfpFq9my6bsecQGJdTHcFCg 189 390 PMEER = (189/30) * 65
TmcCoVxM457qo1dtchcjsm9z8rr7YCedAKo 131 260 PMEER = (131/30) * 65
TmgpqUCHq2ibCTKsn2hDxqQ5i7RyyaqrA66 106 195 PMEER = (106/30) * 65
TmdbCb3MMdK8AxW1P74oPTZfEa4YCq14wLV 92 195 PMEER = (92/30) * 65
TmeGMabPcGPZLBXWXAuD9o1oEyfXU3ucMKh 68 130 PMEER = (68/30) * 65
TmVPSevet6ejiFVfYiuhi5CDhLhqLPPVudZ 39 65 PMEER = (39/30) * 65
Tmbz9mNaiuWKJdvpid9XQLHbFcUXPHSFY3j 32 65 PMEER = (32/30) * 65
The totol reward is :  154505 PMEER
```


## 0.8.0 & 0.8.1 Testnet Reward

### The 162 Mined Blocks List

|Num   | Block | Info        |                                                                  |
|------|-------|-------------|------------------------------------------------------------------|
|1     |18978 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 03db8eb3efcc0aa0963af175add691243d23f6a99a5a4a179f43d8e8a73cf868 |
|      |       | *txsvalid*  | true                                                             |
|2     |19375 | *address*   | TmkuVRGjxor5iNJKThmkoRMKe4cBX8DRc6e                           |
|      |       | *hash*      | 087846f451039254bcc755fdf9bbb239a2d2260aaf2f1c462b967c9f3e1023f1 |
|      |       | *txsvalid*  | true                                                             |
|3     |19445 | *address*   | TmZs2BjFxze1oPG28tQ8NfSrw8mHUrti9rc                           |
|      |       | *hash*      | 04bf10c829f1097a90d0b70cf0d4735a66e059fe63e488b3c9f7b987e577a7d5 |
|      |       | *txsvalid*  | true                                                             |
|4     |19706 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 03ca89de2fbd65aae1eef559b9d072719466d05bc24dfa67850621d2ab952b00 |
|      |       | *txsvalid*  | true                                                             |
|5     |19880 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 053198986cd610263f255a0a01b17c955ddb47cb58fa86edc209d37348d56fee |
|      |       | *txsvalid*  | true                                                             |
|6     |19988 | *address*   | TmUQjNKPA3dLBB6ZfcKd4YSDThQ9Cqzmk5S                           |
|      |       | *hash*      | 032d821d4aa7ead3e94f2c37ece7d323494e927f94e86788f2de3a3aaba66a75 |
|      |       | *txsvalid*  | true                                                             |
|7     |20023 | *address*   | TmdyFy6HNfQAWooYGhyS4wQgNZQmRMZsQnN                           |
|      |       | *hash*      | 053ca7fa961d27c73d8b5cfbed81b1c374bafa47542dde107f5b4597120ce656 |
|      |       | *txsvalid*  | true                                                             |
|8     |20167 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 0343b4f7d49467522eb81b35cb4bbf47254af87aa6e28f8637c745e94c1165e3 |
|      |       | *txsvalid*  | true                                                             |
|9     |20264 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | 95809c7889eab902dbe0dc34cede037b58a4e039e704d26f9a47478aa9504cd3 |
|      |       | *txsvalid*  | true                                                             |
|10     |20273 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 0312d5ccb7b7a4aa908d9cf565a4e333eafecaf7ab79a9546a68596a3aeb56f2 |
|      |       | *txsvalid*  | true                                                             |
|11     |20506 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 0387c8f6629f04c43a98ed509ac99fc60c9ee775a49089f15b8c7f4d57d2bb1d |
|      |       | *txsvalid*  | true                                                             |
|12     |20524 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 044d462eabd55d1552c1aa8bbf4d91352a26aa9d16b96f5339f29a1b788d7247 |
|      |       | *txsvalid*  | true                                                             |
|13     |20553 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | f9e804b4e6fdb5c1664dc8c09dbc0579e5b136fab7a54521e4c8bf8ca696aea5 |
|      |       | *txsvalid*  | true                                                             |
|14     |20567 | *address*   | TmRZH8a4nQhBkfYnq28x1HdTNQ5wMJLWPh4                           |
|      |       | *hash*      | 0080cf467203344a90be0e48ec2db1125231ca85ccba7a5ce3e6716e70fd068a |
|      |       | *txsvalid*  | true                                                             |
|15     |20570 | *address*   | TmTr6FyJeBWq8LHdj1ZbH84nKxDiz3tQGJG                           |
|      |       | *hash*      | 028b0a53660e4d8326cfe5218d24ff28f5616a4d6b24cd36b3ed8ade81f4fc8a |
|      |       | *txsvalid*  | true                                                             |
|16     |20575 | *address*   | TmcwnXt3d3bMxp7p2KY4jWJnZ3RALX8cABH                           |
|      |       | *hash*      | 003cd3dceeb0aa555b590ece1252fb5e57f544a237b573e3b7e335d9f0645b52 |
|      |       | *txsvalid*  | true                                                             |
|17     |20612 | *address*   | TmYWs9RgJmFnc5J9tGWsTYQ3PSyFgqVSTVx                           |
|      |       | *hash*      | 029c6580c8db10b5e0586dcc4b41221125af8bf80ee83a672a0a21c4896d4bbc |
|      |       | *txsvalid*  | true                                                             |
|18     |20737 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 041176f5b7c3feb3adda8619e90cc6190ee7a6be0164d8ada023d2a50e598e87 |
|      |       | *txsvalid*  | true                                                             |
|19     |20774 | *address*   | TmTr6FyJeBWq8LHdj1ZbH84nKxDiz3tQGJG                           |
|      |       | *hash*      | 0196d238f10854bd5432e58ab6ab9b463e8bc3e32b0bd210d31127fc7c4796ea |
|      |       | *txsvalid*  | true                                                             |
|20     |20871 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 00a2df1840fc24a87c8a08928f58ec36762075a02ac1249b0d3471983acf22bf |
|      |       | *txsvalid*  | true                                                             |
|21     |20881 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 051edfe4c767970d2ca5d24850d358c3deaff3b89e7c5611d55ca3b15394456e |
|      |       | *txsvalid*  | true                                                             |
|22     |20920 | *address*   | TmZs2BjFxze1oPG28tQ8NfSrw8mHUrti9rc                           |
|      |       | *hash*      | 061cc54988ae00e8000bc8222b4c320b8b82c638d9bd9da19b6f3f35eb0ca810 |
|      |       | *txsvalid*  | true                                                             |
|23     |21017 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 03e24d526f2382e6c0dd83434254e39a2f7d46ac7f86f71e4153df74350ab900 |
|      |       | *txsvalid*  | true                                                             |
|24     |21059 | *address*   | TmUWxxq66VCdAJtf5kxPVGwHN8LzXyqRk8L                           |
|      |       | *hash*      | 03ef5552c1f7d24b8d17607cfa748bf388d510e9bc91faedc5fcfaa15b6928c8 |
|      |       | *txsvalid*  | true                                                             |
|25     |21120 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | 7394c0a0701c71260b084a59374c778993b0a6820b9170e99fd3b7a7d8fa707b |
|      |       | *txsvalid*  | true                                                             |
|26     |21299 | *address*   | TmUWxxq66VCdAJtf5kxPVGwHN8LzXyqRk8L                           |
|      |       | *hash*      | 02ffc49bc051a4291fc88f4b5e2cfbfafe29bf7b18d1f247eecb970d90263ede |
|      |       | *txsvalid*  | true                                                             |
|27     |21434 | *address*   | TmRxANuoKN9xPvtAaLe1SA77B279pQC9Wbo                           |
|      |       | *hash*      | 057ea11888ade839c850f6950f3976bd49a53dde107f97bb18243590338ac349 |
|      |       | *txsvalid*  | true                                                             |
|28     |21460 | *address*   | TmcwnXt3d3bMxp7p2KY4jWJnZ3RALX8cABH                           |
|      |       | *hash*      | 01a9402ac873232d5c7cbbc0cbd7ca51bff1f8a7f6cf21c9d995377c8981cf80 |
|      |       | *txsvalid*  | true                                                             |
|29     |21539 | *address*   | TmkuVRGjxor5iNJKThmkoRMKe4cBX8DRc6e                           |
|      |       | *hash*      | 0076629666134f63bc3f208254193fec5f7f0f472db95aef912a131ddeeef93b |
|      |       | *txsvalid*  | true                                                             |
|30     |21584 | *address*   | TmWRM7fk8SzBWvuUQv2cJ4T7nWPnNmzrbxi                           |
|      |       | *hash*      | 01942b1918fb43857beaa0fbaa47d0ab13db980e82acd836c52d4fa9c598b43d |
|      |       | *txsvalid*  | true                                                             |
|31     |21760 | *address*   | TmUWxxq66VCdAJtf5kxPVGwHN8LzXyqRk8L                           |
|      |       | *hash*      | 035c049bedd9cb39bb07e51c2771e737e3a27e3f4e8fc288cfb9ce5d2f1c2878 |
|      |       | *txsvalid*  | true                                                             |
|32     |21952 | *address*   | Tmgw6ttQsrWRsWfXPw6BgWRcviLarZUca48                           |
|      |       | *hash*      | 02668a059384f14f5b30f73b861765b03c9ad8a31e38f9a0156bcff66c2f2d0e |
|      |       | *txsvalid*  | true                                                             |
|33     |22071 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 0030002832137844ba63a31afff1eb017029f38c23665e4006d0ad4aa1b9fbc4 |
|      |       | *txsvalid*  | true                                                             |
|34     |22103 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 03e0efb036ba46d98db36d7de62f68eb15d69c39326a27d1745097f960fa9028 |
|      |       | *txsvalid*  | true                                                             |
|35     |22396 | *address*   | TmTr6FyJeBWq8LHdj1ZbH84nKxDiz3tQGJG                           |
|      |       | *hash*      | 0308b40d4ba65830f5eb4c5fcb930d26b4f7b924573c7695dcc2463e15bace66 |
|      |       | *txsvalid*  | true                                                             |
|36     |22703 | *address*   | Tmcubtx2XRRQMoBa4Cb2LHJhWnR1TLXEEjB                           |
|      |       | *hash*      | 0047f3d248a08474a40d0cffc087092614484e8ffe6beaab5c4ac96836c3e18c |
|      |       | *txsvalid*  | true                                                             |
|37     |22708 | *address*   | TmTr6FyJeBWq8LHdj1ZbH84nKxDiz3tQGJG                           |
|      |       | *hash*      | 0206b3bcb404aef6de05e5e03dc848cdf69f02ded6d270edddb9575bf3379ed0 |
|      |       | *txsvalid*  | true                                                             |
|38     |22881 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 0050e0d2e769936a7a99ea9be1ee339e5ac1072eb839ceec741e35f168c451d5 |
|      |       | *txsvalid*  | true                                                             |
|39     |23098 | *address*   | TmUvv3cYB4TDtqSo47kKBvkvGn7JjgWwuGe                           |
|      |       | *hash*      | 01432ffdcf875ccff1838fb4c12490a2eca2b64f0b4423c878b5a357d5448454 |
|      |       | *txsvalid*  | true                                                             |
|40     |23116 | *address*   | TmPK4WJRLqPjhjGEpVmofZ19apjD2rsnSoF                           |
|      |       | *hash*      | 01a1e33845b8f89c8f5368702c69f6327f934e7d73da9d74e0905d48aec2952e |
|      |       | *txsvalid*  | true                                                             |
|41     |23227 | *address*   | TmTr6FyJeBWq8LHdj1ZbH84nKxDiz3tQGJG                           |
|      |       | *hash*      | 0236ad7c7f7086c31710ed4149838817e1712dc457be82d217de891e0f97f427 |
|      |       | *txsvalid*  | true                                                             |
|42     |23252 | *address*   | Tmc2PkNwmPznWu4erZTbL1JXWFnVHXKdWxj                           |
|      |       | *hash*      | 045c3f2b35d5825e342e7ff63c4e250af08c9d95c234980224460dfb6be5b328 |
|      |       | *txsvalid*  | true                                                             |
|43     |23345 | *address*   | Tmcubtx2XRRQMoBa4Cb2LHJhWnR1TLXEEjB                           |
|      |       | *hash*      | 0222776df6b42f188d6c34610c55fb052f95b15d906532567ead6d3a905732f9 |
|      |       | *txsvalid*  | true                                                             |
|44     |23550 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 03c9b3e7261cafe69457d86c5ed10168ebac12d9463d3e1f73d0ea3ab2f985db |
|      |       | *txsvalid*  | true                                                             |
|45     |23637 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 023cdea94b29701fe9f5297d7946e1840d721ae25f7305eb53efb9b7b309295d |
|      |       | *txsvalid*  | true                                                             |
|46     |23663 | *address*   | TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5                           |
|      |       | *hash*      | 03e1b77de93d8115cafc17bbfe6fdab1b2e7f01a809eed237e84e8fefe235443 |
|      |       | *txsvalid*  | true                                                             |
|47     |23922 | *address*   | TmjpmECxxekaLjZA61f6T32cmmPmvjUr7iD                           |
|      |       | *hash*      | 05d2a9863966a4a92668378269a761855ec37a05a24d356b31278f3d8ddbe743 |
|      |       | *txsvalid*  | true                                                             |
|48     |24014 | *address*   | TmaAh9M9232fjkfPJaMPRK15i5Nn9YAYQCM                           |
|      |       | *hash*      | 030ae9cad44e3897037cb2554823d516f827f36f849da4e2f1ef2ef0b45dd2e9 |
|      |       | *txsvalid*  | true                                                             |
|49     |24073 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | ac453da369ba6cae06a3a81034103904f4d9c6ec693bbea360db2dce22fc1a48 |
|      |       | *txsvalid*  | true                                                             |
|50     |24109 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | da5265bd0f5f2f2d350051f65ebc67982ca11dbcbf46b03aa557d53a77c05b42 |
|      |       | *txsvalid*  | true                                                             |
|51     |24286 | *address*   | TmkuVRGjxor5iNJKThmkoRMKe4cBX8DRc6e                           |
|      |       | *hash*      | 043483ca9a0a19467121ffc153d1c326d43c776b499cadd3e20e0f523ba0da3a |
|      |       | *txsvalid*  | true                                                             |
|52     |24380 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 05f31f0ff78e17e0eef20538a66dea5c36f54513178ab88608fced47134a4b47 |
|      |       | *txsvalid*  | true                                                             |
|53     |24558 | *address*   | TmUWxxq66VCdAJtf5kxPVGwHN8LzXyqRk8L                           |
|      |       | *hash*      | 01e9aae63c0aa3963510baa50668a8f48d0c996d0fd003e920976ae86dfce8a0 |
|      |       | *txsvalid*  | true                                                             |
|54     |24560 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 0213445022fa90fabdb4561f95913d690cc8fcdfdc0a3533dac6f6773a3e4f27 |
|      |       | *txsvalid*  | true                                                             |
|55     |24584 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 06961e09a0576eaaf444745c36a3930dd0257a71703e2738d15d9086974d2bcf |
|      |       | *txsvalid*  | true                                                             |
|56     |24623 | *address*   | TmborPU4WChv8SGo6rfRuTo8JL9cKFqmCXj                           |
|      |       | *hash*      | 0438c0dc46726adcf9df91e465b33f8c96a9f2a12737a2f15de281f53572f042 |
|      |       | *txsvalid*  | true                                                             |
|57     |24923 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | 2752cba7e61b3b3035d9d0e3c65eac24b0111b24eacf119325c311dd5d7f1a37 |
|      |       | *txsvalid*  | true                                                             |
|58     |25004 | *address*   | TmZjUjatt9yixCxWX7bwyB2DbM8PpPWSLPb                           |
|      |       | *hash*      | 04f5b897c1d66994611fd6c818cc2307734f84a99a1717b9333030bac5b51934 |
|      |       | *txsvalid*  | true                                                             |
|59     |25041 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | 7fa10bd0bb72bcbc77a1ff003c0ef1db968ffabb89b0bcd6b351cd2e5ccd7a19 |
|      |       | *txsvalid*  | true                                                             |
|60     |25254 | *address*   | TmRic3PMcHT8mPVfSYUdhUcjSDYmBXDyUmv                           |
|      |       | *hash*      | 0701b0bd8ee76dba779c9ace99bdf87ceec1bdae6afa4713ec17cef72e15f799 |
|      |       | *txsvalid*  | true                                                             |
|61     |25343 | *address*   | TmRic3PMcHT8mPVfSYUdhUcjSDYmBXDyUmv                           |
|      |       | *hash*      | 0426cef320aec3fc8f3ac6c1606dcaddac3afed7e000f593f8e0cca5f0d873ad |
|      |       | *txsvalid*  | true                                                             |
|62     |25511 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 05fc36d120dd0f8b61879c675548e0c4834c19892352d72b037118355ed43da3 |
|      |       | *txsvalid*  | true                                                             |
|63     |25520 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 0740c40793a9903e9c305e6dd2cede80cfbaf3aa0a79846c64a3b7d7c82c4f5d |
|      |       | *txsvalid*  | true                                                             |
|64     |25748 | *address*   | TmhDBBZgJ3HL4mBBW8cBGNECKDakZmvRFkc                           |
|      |       | *hash*      | 00128cea75e344685701784447a5783b623b19bd350c6885d5bb73a52762e23f |
|      |       | *txsvalid*  | true                                                             |
|65     |25931 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 027a7fdfe4a4aa98b72a31e14912892be6537eec0a09f3d4e8bef2c3e5ba3f3d |
|      |       | *txsvalid*  | true                                                             |
|66     |26214 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 018ccb65d610aeafbecac04c1fe0ecb61834aa4cc1215b762d977892198b042a |
|      |       | *txsvalid*  | true                                                             |
|67     |26296 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 015b4a60a92cec26648aba14f8bd1d76301ef7fecf409074034670b66182bc9d |
|      |       | *txsvalid*  | true                                                             |
|68     |26386 | *address*   | TmhKLnX4knKkr4e7iHfw1X7BfjDFWrXkfGC                           |
|      |       | *hash*      | 00bb0ab425c3103a14bbb2f4217cd97359bf9baed2d3aeba773e31fd0b00db06 |
|      |       | *txsvalid*  | true                                                             |
|69     |26513 | *address*   | TmdynELSnV85MivyKsHWGjh3d6gSyDxjooD                           |
|      |       | *hash*      | 9351af8df286e8b64358491429afd0aefb01a9f1fe484514164b8edb17af4ab4 |
|      |       | *txsvalid*  | true                                                             |
|70     |26568 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 024b82d41ffbcdf549588b39d9cf6e8d7abf8ba14d005ab5eebb08ef3d452e17 |
|      |       | *txsvalid*  | true                                                             |
|71     |26607 | *address*   | TmeGMabPcGPZLBXWXAuD9o1oEyfXU3ucMKh                           |
|      |       | *hash*      | 0316ee94e79d4eabe9a17dce5ef6f54357fd4f664e1405d1388386c7343bbb93 |
|      |       | *txsvalid*  | true                                                             |
|72     |26708 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 01a61354d665a5bd836535da1ac4c5289138c5dec979a76c86ad51e1f34782a2 |
|      |       | *txsvalid*  | true                                                             |
|73     |26847 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | 3901044e2e85aa8951dc9c0956e875cdb185c7aae62ce5973e69c5ff497640f8 |
|      |       | *txsvalid*  | true                                                             |
|74     |26881 | *address*   | TmkuVRGjxor5iNJKThmkoRMKe4cBX8DRc6e                           |
|      |       | *hash*      | 0367c9b753db77d3f8bbe06390de2b6cd90d836a5408c97840e5b6065fe4cbe4 |
|      |       | *txsvalid*  | true                                                             |
|75     |26959 | *address*   | Tmg4w8T53Ptr2Lsrdtn6HK8EBoKDZ4qDRNn                           |
|      |       | *hash*      | 0084c93cbb99b10b96f3eb3b366bcd9c6c0bdbbb5cb64b3274f2b7393d85bbda |
|      |       | *txsvalid*  | true                                                             |
|76     |27065 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | aac482e8bad07efc73296797802e3ae2a863264a71839b289efe8e11b5a9c909 |
|      |       | *txsvalid*  | true                                                             |
|77     |27097 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 0004c359c4685b9dd11b5deb226ca8abf527d233c7ec4c7f8e66ca4b6fe62a11 |
|      |       | *txsvalid*  | true                                                             |
|78     |27120 | *address*   | TmborPU4WChv8SGo6rfRuTo8JL9cKFqmCXj                           |
|      |       | *hash*      | 024a36c0f1af63fe7cc4495dfdc7481b96a852f7095735ed7e1a0fc0e2e73e65 |
|      |       | *txsvalid*  | true                                                             |
|79     |27192 | *address*   | TmdynELSnV85MivyKsHWGjh3d6gSyDxjooD                           |
|      |       | *hash*      | 35b0ba26261db772cec22e8a115fb231ac46a87fc7cfdaba33b14ba390ccdbb8 |
|      |       | *txsvalid*  | true                                                             |
|80     |27209 | *address*   | TmWRM7fk8SzBWvuUQv2cJ4T7nWPnNmzrbxi                           |
|      |       | *hash*      | 019f01e0527612b63cfca456b745f6619601bbfb65a8b607dd6b44a82a9c9a0d |
|      |       | *txsvalid*  | true                                                             |
|81     |27229 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 0066842d4ec0065be4a8120798916d753a026c5f839d8375feaf157724664e9b |
|      |       | *txsvalid*  | true                                                             |
|82     |27252 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | eaffa289641e4484bba41651efa00811c52a5f6a6d39a7088be1dcb4f2b54e2b |
|      |       | *txsvalid*  | true                                                             |
|83     |27256 | *address*   | TmQ67Hr2hfHpYEU4ThKPRnVE62h8aBMY3DT                           |
|      |       | *hash*      | 007ba168186dea77c3d2fd2a4b5951dd85250330f9a69066af956f70be0761e0 |
|      |       | *txsvalid*  | true                                                             |
|84     |27279 | *address*   | Tmg4w8T53Ptr2Lsrdtn6HK8EBoKDZ4qDRNn                           |
|      |       | *hash*      | 0054203505265c6ba0994b3e843f8d8a560e9887c648634a1c17eeb56f98a437 |
|      |       | *txsvalid*  | true                                                             |
|85     |27290 | *address*   | TmYd9W3KKqD76J4SXK8mkh9uXYmWuwgk3U2                           |
|      |       | *hash*      | 00392ba3063917f6729c34ff808d14fabc5988c260e8882b140c0cf3791d7152 |
|      |       | *txsvalid*  | true                                                             |
|86     |27464 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 01ef77e4dec041a3114a9bf34ce0b4465301e4646580fa0a08a83581d9cd59c0 |
|      |       | *txsvalid*  | true                                                             |
|87     |27469 | *address*   | TmdynELSnV85MivyKsHWGjh3d6gSyDxjooD                           |
|      |       | *hash*      | 7c7be2f76811219972f75aae20b2e5c1bb9c63346fac95fa6a2e64a4ca393744 |
|      |       | *txsvalid*  | true                                                             |
|88     |27552 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 01ae7940b50ee14f00eba74f6a0daafcacb8799a5535cf085ca2fd4d6f760de7 |
|      |       | *txsvalid*  | true                                                             |
|89     |27593 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 02cc494cca15c4ac39c67f6c2082861dd1cacd7d5a152984d1f8fd8279f7f01f |
|      |       | *txsvalid*  | true                                                             |
|90     |27603 | *address*   | TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv                           |
|      |       | *hash*      | e7fdd25149c1e5473ae6617f53de05536152e54c0a99e41f0a6c3970496081a0 |
|      |       | *txsvalid*  | true                                                             |
|91     |27700 | *address*   | TmgV2kTjryz7XTivVSfQcPjpUfenBEdTdJ9                           |
|      |       | *hash*      | 00a010267d58aeac49094060519a1688c25f27394d934a338a09715cffd60966 |
|      |       | *txsvalid*  | true                                                             |
|92     |27802 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 04a229b7b6704e226ab4a971d0caa67cba2236f1bad690d79f0eecac891a6056 |
|      |       | *txsvalid*  | true                                                             |
|93     |27804 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 0333ba87b0d788ac127dbcbdbe29f28b939ef5d6a00b825e4537e0d505d258f3 |
|      |       | *txsvalid*  | true                                                             |
|94     |27808 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 0547234ffaf28de924fb0c1096c15cb0e434d6ca28e0562e3c08c1420a392e1e |
|      |       | *txsvalid*  | true                                                             |
|95     |27855 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 047a50af75293e9906f9255ba6d4dd9932338441d22f52e8619bdb0afecf355d |
|      |       | *txsvalid*  | true                                                             |
|96     |27897 | *address*   | TmhKLnX4knKkr4e7iHfw1X7BfjDFWrXkfGC                           |
|      |       | *hash*      | 018fd88277d037a30f5cf557f7964249202a60c12fd2a9adc1935f9c3e3a731d |
|      |       | *txsvalid*  | true                                                             |
|97     |28060 | *address*   | Tmg4w8T53Ptr2Lsrdtn6HK8EBoKDZ4qDRNn                           |
|      |       | *hash*      | 028df3a893a500cfde1a099331c7f80792d264ecdb00da47ce7f3974552990fa |
|      |       | *txsvalid*  | true                                                             |
|98     |28157 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 00e12672dc034ba6aa9c638475e0ef87f6d880a65ebbe7c9244d7adb67c4141e |
|      |       | *txsvalid*  | true                                                             |
|99     |28197 | *address*   | TmhKLnX4knKkr4e7iHfw1X7BfjDFWrXkfGC                           |
|      |       | *hash*      | 01818778ddb40e030911e2bfb3c9f566479b8b1019ac9f06eca112444836643e |
|      |       | *txsvalid*  | true                                                             |
|100     |28604 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 084286b50cd3392f5a3069e9f387392a65f7194961a0928c895306f0c89f395d |
|      |       | *txsvalid*  | true                                                             |
|101     |28662 | *address*   | TmRxANuoKN9xPvtAaLe1SA77B279pQC9Wbo                           |
|      |       | *hash*      | 0164945721647a65a58273b8d54f2dc9dbba50e65b42a4d03fffa4f2ba27c3fd |
|      |       | *txsvalid*  | true                                                             |
|102     |28747 | *address*   | TmdynELSnV85MivyKsHWGjh3d6gSyDxjooD                           |
|      |       | *hash*      | 4854851cb8f392eecc7277b62c9b821d4c97208ec833393f6acc7c54cbbacb3e |
|      |       | *txsvalid*  | true                                                             |
|103     |28815 | *address*   | TmeGMabPcGPZLBXWXAuD9o1oEyfXU3ucMKh                           |
|      |       | *hash*      | 002fe8679c86a6bfc55be44049fcacee38f22232103db1aa97b84c290a193ad3 |
|      |       | *txsvalid*  | true                                                             |
|104     |28822 | *address*   | TmYd9W3KKqD76J4SXK8mkh9uXYmWuwgk3U2                           |
|      |       | *hash*      | 04676e6ae4e95e82fff1e0792345ff0790515b1d358beb8a97dc2cd9ded72d64 |
|      |       | *txsvalid*  | true                                                             |
|105     |28894 | *address*   | TmgV2kTjryz7XTivVSfQcPjpUfenBEdTdJ9                           |
|      |       | *hash*      | 016cf650074c8e740b6ea8da6cfbfdeffa7e8be555b1e09951aa414573aa7043 |
|      |       | *txsvalid*  | true                                                             |
|106     |28912 | *address*   | TmPK4WJRLqPjhjGEpVmofZ19apjD2rsnSoF                           |
|      |       | *hash*      | 0052cbd9df036614df4da2b0204cb9f1ccc7e1748f56a4c03fc08fc372970451 |
|      |       | *txsvalid*  | true                                                             |
|107     |28945 | *address*   | TmeGMabPcGPZLBXWXAuD9o1oEyfXU3ucMKh                           |
|      |       | *hash*      | 00e4b59b58658e0feef8d65eae1e6b233c4646a0b21b627d19a9ad0f7375ab46 |
|      |       | *txsvalid*  | true                                                             |
|108     |29264 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 016fd19a0d7301dde2a414f95843ea52275de5add4a609853d60fce6fac3b097 |
|      |       | *txsvalid*  | true                                                             |
|109     |29273 | *address*   | TmgV2kTjryz7XTivVSfQcPjpUfenBEdTdJ9                           |
|      |       | *hash*      | 024bcda56010c612c12fb6595687cbebb04b86887fd0d4abdf233a1214df47a4 |
|      |       | *txsvalid*  | true                                                             |
|110     |29506 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 05f34c468ddadf8bace6c7bcd8b77f19c2dac9991b25fa16636dc3cafc657f50 |
|      |       | *txsvalid*  | true                                                             |
|111     |29508 | *address*   | TmRxANuoKN9xPvtAaLe1SA77B279pQC9Wbo                           |
|      |       | *hash*      | 04c15fea323a0643c5942f2e1e8d428345677b9e29a847c279253154d812a30d |
|      |       | *txsvalid*  | true                                                             |
|112     |29559 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 030855c00e04f321434309922a7a878e288fe374643ac4e0e9718f682a9510ac |
|      |       | *txsvalid*  | true                                                             |
|113     |29635 | *address*   | TmdynELSnV85MivyKsHWGjh3d6gSyDxjooD                           |
|      |       | *hash*      | 3f10d0c13ed877635e0256d5cd05dfccee1d973c711cab9ca50548b197f0b9a6 |
|      |       | *txsvalid*  | true                                                             |
|114     |29667 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 030bd7f233a207362d4981ef05e9d1405482374303f371f2f28256b37a0c3059 |
|      |       | *txsvalid*  | true                                                             |
|115     |29810 | *address*   | TmdynELSnV85MivyKsHWGjh3d6gSyDxjooD                           |
|      |       | *hash*      | 81765019c2273c7a0b6c3a04fb5d23c8abd2f9846524c8ef6ad2e25cb70d1821 |
|      |       | *txsvalid*  | true                                                             |
|116     |29813 | *address*   | TmdynELSnV85MivyKsHWGjh3d6gSyDxjooD                           |
|      |       | *hash*      | 3f864d2544361acc543c9cd42841a2718c6a5dd41b847f8cfd95825a35f52a1b |
|      |       | *txsvalid*  | true                                                             |
|117     |29912 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 0163f786b2404926cc6a633a2d5057b682ca91592b4045cab67952e78d01d0da |
|      |       | *txsvalid*  | true                                                             |
|118     |30145 | *address*   | TmWRM7fk8SzBWvuUQv2cJ4T7nWPnNmzrbxi                           |
|      |       | *hash*      | 0607e61f68a56ebcbf44315d360a65a8fa2d1c6f6d6f7a2e527eeec9b702e64f |
|      |       | *txsvalid*  | true                                                             |
|119     |30233 | *address*   | Tmg4w8T53Ptr2Lsrdtn6HK8EBoKDZ4qDRNn                           |
|      |       | *hash*      | 016bf15eb10f7d3e1b34be4738498d11ff0eb01afd5a4cd29140181d4a887506 |
|      |       | *txsvalid*  | true                                                             |
|120     |30293 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 038449a51fda184af9a69c37f2249e29083f1c0df260048e1fdc31ba81f30cc4 |
|      |       | *txsvalid*  | true                                                             |
|121     |30454 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 018be1f394f523ccd97fce656f8e63406e776fc23871fe77b171372c6c4c51e1 |
|      |       | *txsvalid*  | true                                                             |
|122     |30492 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 05f95210c9667d03b5f076028a496ff30dea6e8f5a3ce6c3a16e52398de6e2ba |
|      |       | *txsvalid*  | true                                                             |
|123     |30564 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 04cce669f3c0254b6625c4fa91be00831b5c5ca61ba06f365167cafff09e695b |
|      |       | *txsvalid*  | true                                                             |
|124     |30566 | *address*   | TmYd9W3KKqD76J4SXK8mkh9uXYmWuwgk3U2                           |
|      |       | *hash*      | 0288ada2847e0711a545ee41ae814b3fd61308c3f38c9fa1780a17df99127751 |
|      |       | *txsvalid*  | true                                                             |
|125     |30743 | *address*   | TmYd9W3KKqD76J4SXK8mkh9uXYmWuwgk3U2                           |
|      |       | *hash*      | 059cf0226ceac4b4065b4ede73e131c08cf043fb2edd63a1728f098fa3b91058 |
|      |       | *txsvalid*  | true                                                             |
|126     |30762 | *address*   | TmbNTwPm9aTN7TtStnJjrcjsaRJWm4gt3Fs                           |
|      |       | *hash*      | 069216a24a8562533974c2b9857e9e2b0a3d619b52a2f339fe294b32ee8352d6 |
|      |       | *txsvalid*  | true                                                             |
|127     |30767 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 0425ba4b1496e2f8f178b2ff92ef186a5031ee1493c27300072e27d4d2b7576a |
|      |       | *txsvalid*  | true                                                             |
|128     |30880 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 05ecf16e2e9045c3ab470dac7b7d762da592f5fb4a1039b1ea51ecf4af52ef8e |
|      |       | *txsvalid*  | true                                                             |
|129     |31095 | *address*   | TmRZH8a4nQhBkfYnq28x1HdTNQ5wMJLWPh4                           |
|      |       | *hash*      | 032a58760473a65a10e13f9ade0f78bf49c514503c608524fb3f83be785a8a48 |
|      |       | *txsvalid*  | true                                                             |
|130     |31138 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 01f3b29035a4c3400dc89beecfd03491e067f3d3f910021d814f3cd86c38efd2 |
|      |       | *txsvalid*  | true                                                             |
|131     |31520 | *address*   | TmUWxxq66VCdAJtf5kxPVGwHN8LzXyqRk8L                           |
|      |       | *hash*      | 00c0f2824d92611e5f235c66bfbada287ad8480d5a26f05372b88f25816fed52 |
|      |       | *txsvalid*  | true                                                             |
|132     |31618 | *address*   | TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW                           |
|      |       | *hash*      | 06ac375db18229edfd517b3709c3b79e793775df17ae6d8e25aa4e04b8f0f311 |
|      |       | *txsvalid*  | true                                                             |
|133     |31810 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 011112188802046d20f1c81c90b382c7a359df3b70c28c18ec5fc8da215ad50f |
|      |       | *txsvalid*  | true                                                             |
|134     |31985 | *address*   | Tmj8PCEms1fAT9SYEWuxfHjbppataAM7rP2                           |
|      |       | *hash*      | 064233104e76cb7ba8d468c50c2d472819faa1f26cd00bc4ec67c54d485ab06f |
|      |       | *txsvalid*  | true                                                             |
|135     |32015 | *address*   | TmRZH8a4nQhBkfYnq28x1HdTNQ5wMJLWPh4                           |
|      |       | *hash*      | 07d1168671e6fc1ff3d8cd7632fb340cb3f5129ca2d2a4f0c0514dc44868899d |
|      |       | *txsvalid*  | true                                                             |
|136     |32172 | *address*   | TmYL7UUoFQNkZedhYWP4SbhtHtYVzgLjmis                           |
|      |       | *hash*      | 0431e70841879b595cecfc18537d2ce91449c74606a2d6c54a628b1e049656f6 |
|      |       | *txsvalid*  | true                                                             |
|137     |32214 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 00fe70b187609df3a3477136f7f55389e49c44cfa2f1fad6e462b35434a93f88 |
|      |       | *txsvalid*  | true                                                             |
|138     |32358 | *address*   | TmWRM7fk8SzBWvuUQv2cJ4T7nWPnNmzrbxi                           |
|      |       | *hash*      | 02ee4f9479e498a8c6a15f7a06f3ead48436cc88e5b0f4a7cb8f811c0f7d9f3d |
|      |       | *txsvalid*  | true                                                             |
|139     |32504 | *address*   | Tmgi1VwHrAG9bb9hXSeEUv76uchPQGgt5pj                           |
|      |       | *hash*      | 01c2d198c720a30f774d5a663aa210770eb08e2a18bb52d5cacdfea4f88fc92b |
|      |       | *txsvalid*  | true                                                             |
|140     |32528 | *address*   | Tmh3je9zbnHAvPfwwHhQsFSJmKkeRTtKqmV                           |
|      |       | *hash*      | 0525a3ccf19890a5d2c5e7c1b61f222f24ebe3889c39a9677ff8a2a714ed30b1 |
|      |       | *txsvalid*  | true                                                             |
|141     |32696 | *address*   | TmUWxxq66VCdAJtf5kxPVGwHN8LzXyqRk8L                           |
|      |       | *hash*      | 06cbcb22143f148c9ffe7eec3b20eced421f7b1c7b44b7b1cbff5fc182964100 |
|      |       | *txsvalid*  | true                                                             |
|142     |32726 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 060fdea12c3cb26ffa24948156a942b8260ceaee1f65ddb5d21f697782852f7e |
|      |       | *txsvalid*  | true                                                             |
|143     |32784 | *address*   | TmRxANuoKN9xPvtAaLe1SA77B279pQC9Wbo                           |
|      |       | *hash*      | 080ec9c4e96c32d4a4202b2e6dc1297536e1d0f41cb27f1688f4add02ec769b9 |
|      |       | *txsvalid*  | true                                                             |
|144     |32821 | *address*   | Tmg4w8T53Ptr2Lsrdtn6HK8EBoKDZ4qDRNn                           |
|      |       | *hash*      | 030fe8ea960454bbd144bdabee6441ddde070efcfea41a9e715c46139fd2bb57 |
|      |       | *txsvalid*  | true                                                             |
|145     |32839 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 0870278ceefd3cf60fd944c5728e097425225da0532c24a1b6d293a97b32289d |
|      |       | *txsvalid*  | true                                                             |
|146     |32991 | *address*   | TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5                           |
|      |       | *hash*      | 017db07a93f0b0f001204764bc36f8938ff92fae60d539801ed5081fedfa555f |
|      |       | *txsvalid*  | true                                                             |
|147     |33011 | *address*   | Tmh3je9zbnHAvPfwwHhQsFSJmKkeRTtKqmV                           |
|      |       | *hash*      | 03763b475d8864d1cd57afb9b0771b537f7d51d3aae341265ec77398448c681a |
|      |       | *txsvalid*  | true                                                             |
|148     |33161 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 03a2c3143d07f16c6e7637f83a85a15e6fabf1955b7c07325c1ca20f7f95102f |
|      |       | *txsvalid*  | true                                                             |
|149     |33417 | *address*   | TmRxANuoKN9xPvtAaLe1SA77B279pQC9Wbo                           |
|      |       | *hash*      | 0117810cbbbf49495911ee5ef49dab82f9801aec3054764ec2b7395fd6170bdf |
|      |       | *txsvalid*  | true                                                             |
|150     |33581 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 05c45dfb4c8158abf52fb00de97109b0069b6a8b98ed3b0c232ac16ff3f6c53f |
|      |       | *txsvalid*  | true                                                             |
|151     |33724 | *address*   | TmbNTwPm9aTN7TtStnJjrcjsaRJWm4gt3Fs                           |
|      |       | *hash*      | 0480f248c4495f7ed25a8f1e4ab0a7309c88a6e675ee2c49d3587b6e56bb99e3 |
|      |       | *txsvalid*  | true                                                             |
|152     |33773 | *address*   | Tmh3je9zbnHAvPfwwHhQsFSJmKkeRTtKqmV                           |
|      |       | *hash*      | 05013cffad51677ad76a12a0bf4b1880172d3776ac937115db733719d42e7fed |
|      |       | *txsvalid*  | true                                                             |
|153     |33779 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 065be6d97b327beb5a8f249588ed98f76126b81ad2b5c7aaf9c953072a13948e |
|      |       | *txsvalid*  | true                                                             |
|154     |33871 | *address*   | TmRZH8a4nQhBkfYnq28x1HdTNQ5wMJLWPh4                           |
|      |       | *hash*      | 06846ab5afefda93cdf39561144af790e057f5dfa8200c6a44c8d1ceaa89246f |
|      |       | *txsvalid*  | true                                                             |
|155     |34023 | *address*   | Tmh3je9zbnHAvPfwwHhQsFSJmKkeRTtKqmV                           |
|      |       | *hash*      | 01a1e44150ca2be1573068e5b08ea0f5f95ee957130cb624afd11d016771757c |
|      |       | *txsvalid*  | true                                                             |
|156     |34055 | *address*   | Tmcubtx2XRRQMoBa4Cb2LHJhWnR1TLXEEjB                           |
|      |       | *hash*      | 00c0f31f91c17b24cf647c168f882b860cf08a5b4ed199386457b69f7cc1ee6e |
|      |       | *txsvalid*  | true                                                             |
|157     |34333 | *address*   | TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj                           |
|      |       | *hash*      | 05d51458ca0436a282ce893a99663af54f072c0340f3fb004f1138f120f1ea70 |
|      |       | *txsvalid*  | true                                                             |
|158     |34498 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 019fad1757ca8103d058f8c9a37d0f8e9b9c00d9205c1faaf75022cc7b998c9c |
|      |       | *txsvalid*  | true                                                             |
|159     |34642 | *address*   | Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7                           |
|      |       | *hash*      | 040c275548130963d683b2ba0b4432f82618c7d787134a3b8937c556c0677fcd |
|      |       | *txsvalid*  | true                                                             |
|160     |34690 | *address*   | Tmh3je9zbnHAvPfwwHhQsFSJmKkeRTtKqmV                           |
|      |       | *hash*      | 02082c8dd3e7932721af004c3053e35779690cafebe36da45dfbbb617e21f72e |
|      |       | *txsvalid*  | true                                                             |
|161     |34761 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 03bbb02d9b1dfe5179c082e1c568bf00df7e87c1cbbfe4ec66fcbcbc54d85a47 |
|      |       | *txsvalid*  | true                                                             |
|162     |34778 | *address*   | TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws                           |
|      |       | *hash*      | 02899a2257106f17e6370d484a69f29f2153a1a569ec17cc99f2543bc978a2ad |
|      |       | *txsvalid*  | true                                                             |

### Reward Ranking

| Address                             |    |
|-------------------------------------|----|
| TmPvQaAAtHWDhzGdsd8gQ2q3rkuzfKJeaws | 19 |
| Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7 | 17 |
| TmfLjksDFTTwaPiNv6zBTGS8tkqMt6ci2Cj | 15 |
| TmdyfadD856zHHnwCMeLvTgpTTaYKGkVaxv | 11 |
| TmRShC5EEhKp9njDyH465xDdpoQY4bA3DpW | 11 |
| TmdynELSnV85MivyKsHWGjh3d6gSyDxjooD | 7 |
| TmUWxxq66VCdAJtf5kxPVGwHN8LzXyqRk8L | 6 |
| Tmh3je9zbnHAvPfwwHhQsFSJmKkeRTtKqmV | 5 |
| Tmg4w8T53Ptr2Lsrdtn6HK8EBoKDZ4qDRNn | 5 |
| TmTr6FyJeBWq8LHdj1ZbH84nKxDiz3tQGJG | 5 |
| TmRxANuoKN9xPvtAaLe1SA77B279pQC9Wbo | 5 |
| TmkuVRGjxor5iNJKThmkoRMKe4cBX8DRc6e | 4 |
| TmYd9W3KKqD76J4SXK8mkh9uXYmWuwgk3U2 | 4 |
| TmWRM7fk8SzBWvuUQv2cJ4T7nWPnNmzrbxi | 4 |
| TmRZH8a4nQhBkfYnq28x1HdTNQ5wMJLWPh4 | 4 |
| TmhKLnX4knKkr4e7iHfw1X7BfjDFWrXkfGC | 3 |
| TmgV2kTjryz7XTivVSfQcPjpUfenBEdTdJ9 | 3 |
| TmeGMabPcGPZLBXWXAuD9o1oEyfXU3ucMKh | 3 |
| Tmcubtx2XRRQMoBa4Cb2LHJhWnR1TLXEEjB | 3 |
| TmcwnXt3d3bMxp7p2KY4jWJnZ3RALX8cABH | 2 |
| TmborPU4WChv8SGo6rfRuTo8JL9cKFqmCXj | 2 |
| TmbNTwPm9aTN7TtStnJjrcjsaRJWm4gt3Fs | 2 |
| TmZs2BjFxze1oPG28tQ8NfSrw8mHUrti9rc | 2 |
| TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5 | 2 |
| TmRic3PMcHT8mPVfSYUdhUcjSDYmBXDyUmv | 2 |
| TmPK4WJRLqPjhjGEpVmofZ19apjD2rsnSoF | 2 |
| TmjpmECxxekaLjZA61f6T32cmmPmvjUr7iD | 1 |
| Tmj8PCEms1fAT9SYEWuxfHjbppataAM7rP2 | 1 |
| TmhDBBZgJ3HL4mBBW8cBGNECKDakZmvRFkc | 1 |
| Tmgw6ttQsrWRsWfXPw6BgWRcviLarZUca48 | 1 |
| Tmgi1VwHrAG9bb9hXSeEUv76uchPQGgt5pj | 1 |
| TmdyFy6HNfQAWooYGhyS4wQgNZQmRMZsQnN | 1 |
| Tmc2PkNwmPznWu4erZTbL1JXWFnVHXKdWxj | 1 |
| TmaAh9M9232fjkfPJaMPRK15i5Nn9YAYQCM | 1 |
| TmZjUjatt9yixCxWX7bwyB2DbM8PpPWSLPb | 1 |
| TmYWs9RgJmFnc5J9tGWsTYQ3PSyFgqVSTVx | 1 |
| TmYL7UUoFQNkZedhYWP4SbhtHtYVzgLjmis | 1 |
| TmUvv3cYB4TDtqSo47kKBvkvGn7JjgWwuGe | 1 |
| TmUQjNKPA3dLBB6ZfcKd4YSDThQ9Cqzmk5S | 1 |
| TmQ67Hr2hfHpYEU4ThKPRnVE62h8aBMY3DT | 1 |
