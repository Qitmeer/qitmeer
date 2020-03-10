package blockdag

import (
	"bytes"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/database"
	"io"
)

// update db to new version
func (bd *BlockDAG) UpgradeDB(dbTx database.Tx, blockTotal uint) error {
	blocks := NewHashSet()
	for i := uint(0); i < blockTotal; i++ {
		block := Block{id: i}
		ib := bd.instance.CreateBlock(&block)
		err := DBGetDAGBlockOld(dbTx, ib)
		if err != nil {
			return err
		}
		blocks.AddPair(ib.GetHash(), ib)

		// add child
		if ib.HasParents() {
			//parentsSet := NewHashSet()
			for k := range ib.GetParents().GetMap() {
				parent := blocks.Get(&k).(*PhantomBlock)
				parent.AddChild(&block)
			}
		}

		/*		err = DBPutDAGBlock(dbTx, ib)
				if err != nil {
					return err
				}*/
	}

	for _, v := range blocks.GetMap() {
		err := DBPutDAGBlock(dbTx, v.(*PhantomBlock))
		if err != nil {
			return err
		}
	}
	return nil
}

// DBGetDAGBlock get dag block data by resouce ID
func DBGetDAGBlockOld(dbTx database.Tx, block IBlock) error {
	bucket := dbTx.Metadata().Bucket(dbnamespace.BlockIndexBucketName)
	var serializedID [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedID[:], uint32(block.GetID()))

	data := bucket.Get(serializedID[:])
	if data == nil {
		return fmt.Errorf("get dag block error")
	}

	pb := block.(*PhantomBlock)
	return BlockDecodeOld(pb, bytes.NewReader(data))
}

// decode
func BlockDecodeOld(pb *PhantomBlock, r io.Reader) error {
	var b *Block = pb.Block

	var id uint32
	err := s.ReadElements(r, &id)
	if err != nil {
		return err
	}
	b.id = uint(id)

	err = s.ReadElements(r, &b.hash)
	if err != nil {
		return err
	}
	// parents
	var parentsSize uint32
	err = s.ReadElements(r, &parentsSize)
	if err != nil {
		return err
	}
	if parentsSize > 0 {
		b.parents = NewHashSet()
		for i := uint32(0); i < parentsSize; i++ {
			var parent hash.Hash
			err := s.ReadElements(r, &parent)
			if err != nil {
				return err
			}
			b.parents.Add(&parent)
		}
	}
	// children
	/*var childrenSize uint32
	err=s.ReadElements(r,&childrenSize)
	if err != nil {
		return err
	}
	if childrenSize>0 {
		b.children = NewHashSet()
		for i:=uint32(0);i<childrenSize ;i++  {
			var children hash.Hash
			err:=s.ReadElements(r,&children)
			if err != nil {
				return err
			}
			b.children.Add(&children)
		}
	}*/
	// mainParent
	var mainParent hash.Hash
	err = s.ReadElements(r, &mainParent)
	if err != nil {
		return err
	}
	if mainParent.IsEqual(&hash.ZeroHash) {
		b.mainParent = nil
	} else {
		b.mainParent = &mainParent
	}

	var weight uint64
	err = s.ReadElements(r, &weight)
	if err != nil {
		return err
	}
	b.weight = uint64(weight)

	var order uint32
	err = s.ReadElements(r, &order)
	if err != nil {
		return err
	}
	b.order = uint(order)

	var layer uint32
	err = s.ReadElements(r, &layer)
	if err != nil {
		return err
	}
	b.layer = uint(layer)

	var height uint32
	err = s.ReadElements(r, &height)
	if err != nil {
		return err
	}
	b.height = uint(height)

	var status byte
	err = s.ReadElements(r, &status)
	if err != nil {
		return err
	}
	b.status = BlockStatus(status)

	var blueNum uint32
	err = s.ReadElements(r, &blueNum)
	if err != nil {
		return err
	}
	pb.blueNum = uint(blueNum)

	// blueDiffAnticone
	var blueDiffAnticoneSize uint32
	err = s.ReadElements(r, &blueDiffAnticoneSize)
	if err != nil {
		return err
	}
	pb.blueDiffAnticone = NewHashSet()
	if blueDiffAnticoneSize > 0 {
		for i := uint32(0); i < blueDiffAnticoneSize; i++ {
			var bda hash.Hash
			err := s.ReadElements(r, &bda)
			if err != nil {
				return err
			}

			var order uint32
			err = s.ReadElements(r, &order)
			if err != nil {
				return err
			}

			pb.blueDiffAnticone.AddPair(&bda, uint(order))
		}
	}

	// blueDiffAnticone
	var redDiffAnticoneSize uint32
	err = s.ReadElements(r, &redDiffAnticoneSize)
	if err != nil {
		return err
	}
	pb.redDiffAnticone = NewHashSet()
	if redDiffAnticoneSize > 0 {
		for i := uint32(0); i < redDiffAnticoneSize; i++ {
			var bda hash.Hash
			err := s.ReadElements(r, &bda)
			if err != nil {
				return err
			}
			var order uint32
			err = s.ReadElements(r, &order)
			if err != nil {
				return err
			}

			pb.redDiffAnticone.AddPair(&bda, uint(order))
		}
	}

	return nil
}
