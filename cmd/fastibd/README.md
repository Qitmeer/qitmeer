# FastIBD
This tool is used for quick start Qitmeer nodes.
### Install
```
~ cd ./cmd/fastibd
~ go build
~ ./fastibd
```

### How to export the data of blocks from node
```
~ ./fastibd export
or
~ ./fastibd export --path=[Output directory]
```

### How to import the data of blocks to node
```
~ ./fastibd import
or
~ ./fastibd import --path=[Input directory]
```

### How to upgrade the data of blocks to node

```
~ ./fastibd --testnet upgrade
```

### First aid mode under consensus error

```
~ ./fastibd --testnet upgrade --aidmode
```


