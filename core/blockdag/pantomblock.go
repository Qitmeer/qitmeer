package blockdag

import (
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/dag"
	s "github.com/Qitmeer/qitmeer/core/serialization"
	"io"
)

type PhantomBlock struct {
	*Block
	blueNum uint

	blueDiffAnticone *dag.HashSet
	redDiffAnticone *dag.HashSet
}

func (pb *PhantomBlock) IsBluer(other *PhantomBlock) bool {
	if pb.blueNum > other.blueNum ||
		(pb.blueNum == other.blueNum && pb.GetHash().String() < other.GetHash().String()) {
		return true
	}
	return false
}

// encode
func (pb *PhantomBlock) Encode(w io.Writer) error {
	err:=pb.Block.Encode(w)
	if err != nil {
		return err
	}
	err=s.WriteElements(w,uint32(pb.blueNum))
	if err != nil {
		return err
	}

	// blueDiffAnticone
	blueDiffAnticone:=[]*hash.Hash{}
	if pb.blueDiffAnticone!=nil && pb.blueDiffAnticone.Size()>0 {
		blueDiffAnticone=pb.blueDiffAnticone.List()
	}
	blueDiffAnticoneSize:=len(blueDiffAnticone)
	err=s.WriteElements(w,uint32(blueDiffAnticoneSize))
	if err != nil {
		return err
	}
	for i:=0;i<blueDiffAnticoneSize ;i++  {
		err=s.WriteElements(w,blueDiffAnticone[i])
		if err != nil {
			return err
		}
	}
	// redDiffAnticone
	redDiffAnticone:=[]*hash.Hash{}
	if pb.redDiffAnticone!=nil && pb.redDiffAnticone.Size()>0 {
		redDiffAnticone=pb.redDiffAnticone.List()
	}
	redDiffAnticoneSize:=len(redDiffAnticone)
	err=s.WriteElements(w,uint32(redDiffAnticoneSize))
	if err != nil {
		return err
	}
	for i:=0;i<redDiffAnticoneSize ;i++  {
		err=s.WriteElements(w,redDiffAnticone[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// decode
func (pb *PhantomBlock) Decode(r io.Reader) error {
	err:=pb.Block.Decode(r)
	if err != nil {
		return err
	}

	var blueNum uint32
	err=s.ReadElements(r,&blueNum)
	if err != nil {
		return err
	}
	pb.blueNum=uint(blueNum)

	// blueDiffAnticone
	var blueDiffAnticoneSize uint32
	err=s.ReadElements(r,&blueDiffAnticoneSize)
	if err != nil {
		return err
	}
	pb.blueDiffAnticone = dag.NewHashSet()
	if blueDiffAnticoneSize>0 {
		for i:=uint32(0);i<blueDiffAnticoneSize ;i++  {
			var bda hash.Hash
			err:=s.ReadElements(r,&bda)
			if err != nil {
				return err
			}
			pb.blueDiffAnticone.Add(&bda)
		}
	}

	// blueDiffAnticone
	var redDiffAnticoneSize uint32
	err=s.ReadElements(r,&redDiffAnticoneSize)
	if err != nil {
		return err
	}
	pb.redDiffAnticone = dag.NewHashSet()
	if redDiffAnticoneSize>0 {
		for i:=uint32(0);i<redDiffAnticoneSize ;i++  {
			var bda hash.Hash
			err:=s.ReadElements(r,&bda)
			if err != nil {
				return err
			}
			pb.redDiffAnticone.Add(&bda)
		}
	}

	return nil
}