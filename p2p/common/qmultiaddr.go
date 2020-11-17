package common

import iaddr "github.com/ipfs/go-ipfs-addr"

type QMultiaddr interface {
	iaddr.IPFSAddr
}

func QMultiAddrFromString(address string) (QMultiaddr, error) {
	addr, err := iaddr.ParseString(address)
	if err != nil {
		return nil, err
	}
	return addr, nil
}
