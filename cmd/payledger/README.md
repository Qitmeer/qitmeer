# Ledger

### How to generate ledger

* You can use this command to generating ledger for the next qitmeerd version.
```
~ cd ./tools/payledger
~ go build
~ ./payledger -h
```
* If you want to use all UTXOs from `srcdatadir`:
```
~ ./payledger --srcdatadir=[YourQitmeerDataPath] --endpoint="*"
```

* If you want to use specific UTXOs from `srcdatadir`:
```
~ ./payledger --srcdatadir=[YourQitmeerDataPath] --endpoint=000005fd233345570677bc257e7c35e300dfe9b6d384bd8a0659c6619ff7ab30
```

* Then, you can build the next qitmeerd version.
```
~ cd ./../../
~ go build
```

### How to generate locked ledger and save payouts
```
cd ./qitmeer/cmd/payledger
go run . --mixnet --srcdatadir=[YourQitmeerDataPath] --endpoint="*" --savefile --unlocksperheight=5000000000
```


### How to show last result of generated ledger
```
~ ./payledger --last
```

### How to find end point
* You can print a list of recommendations when using `showendpoints` to set the specific number.
```
~ ./payledger --showendpoints=100
```
* Or skip some blocks
```
~ ./payledger --showendpoints=100 --endpointskips=200
```
* You can check if the end point works
```
~ ./payledger --checkendpoint=000005fd233345570677bc257e7c35e300dfe9b6d384bd8a0659c6619ff7ab30
```

### How to debug address
```
~ ./payledger --debugaddress=[Qitmeer Address]
or
~ ./payledger --debugaddress=[Qitmeer Address] --debugaddrutxo
or
~ ./payledger --debugaddress=[Qitmeer Address] --debugaddrvalid
```


### How to show blocks info
```
~ ./payledger --blocksinfo
```