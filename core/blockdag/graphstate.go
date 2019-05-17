package blockdag

import (
	"fmt"
	"io"
	"qitmeer/common/hash"
	s "qitmeer/core/serialization"
)

// A general description of the whole state of DAG
type GraphState struct {
	tips *HashSet
	total uint
	layer uint
}

func (gs *GraphState) GetLayer() uint {
	return gs.layer
}

func (gs *GraphState) GetTotal() uint {
	return gs.total
}

func (gs *GraphState) GetTips() *HashSet {
	return gs.tips
}

func (gs *GraphState) IsEqual(other *GraphState) bool {
	if gs==other {
		return true
	}
	if gs.layer!=other.layer || gs.total != other.total {
		return false
	}
	return gs.tips.IsEqual(other.tips)
}

func (gs *GraphState) Equal(other *GraphState) {
	if gs.IsEqual(other) {
		return
	}
	gs.tips=other.tips.Clone()
	gs.layer=other.layer
	gs.total=other.total
}

func (gs *GraphState) Clone() *GraphState {
	result:=NewGraphState()
	result.Equal(gs)
	return result
}

func (gs *GraphState) String() string {
	return fmt.Sprintf("(%v,%d,%d)",gs.tips,gs.total,gs.layer)
}

func (gs *GraphState) IsExcellent(other *GraphState) bool {
	if gs.IsEqual(other) {
		return false
	}
	if gs.total<other.total {
		return false
	}else if gs.total>other.total {
		return true
	}
	if gs.layer<other.layer {
		return false
	}else if gs.layer>other.layer {
		return true
	}
	return false
}

func (gs *GraphState) Encode(w io.Writer,pver uint32) error {
	err:= s.WriteVarInt(w, pver, uint64(gs.total))
	if err != nil {
		return err
	}
	err= s.WriteVarInt(w, pver, uint64(gs.layer))
	if err != nil {
		return err
	}
	err= s.WriteVarInt(w, pver, uint64(gs.tips.Len()))
	if err != nil {
		return err
	}

	for k,_:= range gs.tips.GetMap() {
		err = s.WriteElements(w, &k)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gs *GraphState) Decode(r io.Reader,pver uint32) error {
	total, err := s.ReadVarInt(r,pver)
	if err != nil {
		return err
	}
	gs.total=uint(total)

	layer, err := s.ReadVarInt(r,pver)
	if err != nil {
		return err
	}
	gs.layer=uint(layer)

	count, err := s.ReadVarInt(r,pver)
	if count==0 {
		return fmt.Errorf("GraphState.Decode:tips count is zero.")
	}

	locatorHashes := make([]hash.Hash, count)
	for i := uint64(0); i < count; i++ {
		h := &locatorHashes[i]
		err := s.ReadElements(r,h)
		if err != nil {
			return err
		}
		gs.tips.Add(h)
	}
	return nil
}

// Create a new GraphState
func NewGraphState() *GraphState {
	return &GraphState{
		tips:NewHashSet(),
		total:0,
		layer:0,
	}
}