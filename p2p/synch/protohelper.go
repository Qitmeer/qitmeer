/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"bytes"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	pb "github.com/Qitmeer/qitmeer/p2p/proto/v1"
)

func changePBHashsToHashs(hs []*pb.Hash) []*hash.Hash {
	result := []*hash.Hash{}
	for _, ha := range hs {
		h, err := hash.NewHash(ha.Hash)
		if err != nil {
			log.Warn(fmt.Sprintf("Can't NewHash:%v", ha.Hash))
			continue
		}
		result = append(result, h)
	}
	return result
}

func changePBHashToHash(ha *pb.Hash) *hash.Hash {
	h, err := hash.NewHash(ha.Hash)
	if err != nil {
		log.Warn(fmt.Sprintf("Can't NewHash:%v", ha.Hash))
		return nil
	}
	return h
}

func changeHashsToPBHashs(hs []*hash.Hash) []*pb.Hash {
	result := []*pb.Hash{}
	for _, ha := range hs {
		result = append(result, &pb.Hash{Hash: ha.Bytes()})
	}
	return result
}

func changePBTxToTx(tx *pb.Transaction) *types.Transaction {
	var transaction types.Transaction
	err := transaction.Deserialize(bytes.NewReader(tx.TxBytes))
	if err != nil {
		return nil
	}
	return &transaction
}
