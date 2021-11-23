package blockdag

import (
	s "github.com/Qitmeer/qng-core/core/serialization"
	"io"
)

type PhantomBlock struct {
	*Block
	blueNum uint

	blueDiffAnticone *IdSet
	redDiffAnticone  *IdSet
}

func (pb *PhantomBlock) IsBluer(other *PhantomBlock) bool {
	if pb.blueNum > other.blueNum {
		return true
	} else if pb.blueNum == other.blueNum {
		if pb.GetData().GetPriority() > other.GetData().GetPriority() {
			return true
		} else if pb.GetData().GetPriority() == other.GetData().GetPriority() {
			if pb.GetHash().String() < other.GetHash().String() {
				return true
			}
		}
	}
	return false
}

// encode
func (pb *PhantomBlock) Encode(w io.Writer) error {
	err := pb.Block.Encode(w)
	if err != nil {
		return err
	}
	err = s.WriteElements(w, uint32(pb.blueNum))
	if err != nil {
		return err
	}

	// blueDiffAnticone
	blueDiffAnticone := []uint{}
	if pb.blueDiffAnticone != nil && pb.blueDiffAnticone.Size() > 0 {
		blueDiffAnticone = pb.blueDiffAnticone.List()
	}
	blueDiffAnticoneSize := len(blueDiffAnticone)
	err = s.WriteElements(w, uint32(blueDiffAnticoneSize))
	if err != nil {
		return err
	}
	for i := 0; i < blueDiffAnticoneSize; i++ {
		err = s.WriteElements(w, uint32(blueDiffAnticone[i]))
		if err != nil {
			return err
		}
		order := pb.blueDiffAnticone.Get(blueDiffAnticone[i]).(uint)
		err = s.WriteElements(w, uint32(order))
		if err != nil {
			return err
		}
	}
	// redDiffAnticone
	redDiffAnticone := []uint{}
	if pb.redDiffAnticone != nil && pb.redDiffAnticone.Size() > 0 {
		redDiffAnticone = pb.redDiffAnticone.List()
	}
	redDiffAnticoneSize := len(redDiffAnticone)
	err = s.WriteElements(w, uint32(redDiffAnticoneSize))
	if err != nil {
		return err
	}
	for i := 0; i < redDiffAnticoneSize; i++ {
		err = s.WriteElements(w, uint32(redDiffAnticone[i]))
		if err != nil {
			return err
		}
		order := pb.redDiffAnticone.Get(redDiffAnticone[i]).(uint)
		err = s.WriteElements(w, uint32(order))
		if err != nil {
			return err
		}
	}
	return nil
}

// decode
func (pb *PhantomBlock) Decode(r io.Reader) error {
	err := pb.Block.Decode(r)
	if err != nil {
		return err
	}

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
	pb.blueDiffAnticone = NewIdSet()
	if blueDiffAnticoneSize > 0 {
		for i := uint32(0); i < blueDiffAnticoneSize; i++ {
			var bda uint32
			err := s.ReadElements(r, &bda)
			if err != nil {
				return err
			}

			var order uint32
			err = s.ReadElements(r, &order)
			if err != nil {
				return err
			}

			pb.blueDiffAnticone.AddPair(uint(bda), uint(order))
		}
	}

	// redDiffAnticone
	var redDiffAnticoneSize uint32
	err = s.ReadElements(r, &redDiffAnticoneSize)
	if err != nil {
		return err
	}
	pb.redDiffAnticone = NewIdSet()
	if redDiffAnticoneSize > 0 {
		for i := uint32(0); i < redDiffAnticoneSize; i++ {
			var bda uint32
			err := s.ReadElements(r, &bda)
			if err != nil {
				return err
			}
			var order uint32
			err = s.ReadElements(r, &order)
			if err != nil {
				return err
			}

			pb.redDiffAnticone.AddPair(uint(bda), uint(order))
		}
	}

	return nil
}

// GetBlueNum
func (pb *PhantomBlock) GetBlueNum() uint {
	return pb.blueNum
}

func (pb *PhantomBlock) GetBlueDiffAnticone() *IdSet {
	return pb.blueDiffAnticone
}

func (pb *PhantomBlock) GetRedDiffAnticone() *IdSet {
	return pb.redDiffAnticone
}
