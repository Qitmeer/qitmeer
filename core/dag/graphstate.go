package dag

import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	s "github.com/HalalChain/qitmeer-lib/core/serialization"
	"io"
)

// Maximum number of the DAG tip
const MaxTips=100

// A general description of the whole state of DAG
type GraphState struct {
	// The terminal block is in block dag,this block have not any connecting at present.
	tips *HashSet

	// The total number blocks that this dag currently owned
	total uint

	// At present, the whole graph nodes has the last layer level.
	layer uint
}

// Return the DAG layer
func (gs *GraphState) GetLayer() uint {
	return gs.layer
}

func (gs *GraphState) SetLayer(layer uint) {
	gs.layer=layer
}

// Return the total of DAG
func (gs *GraphState) GetTotal() uint {
	return gs.total
}

func (gs *GraphState) SetTotal(total uint) {
	gs.total=total
}

// Return all tips of DAG
func (gs *GraphState) GetTips() *HashSet {
	return gs.tips
}

func (gs *GraphState) SetTips(tips *HashSet) {
	gs.tips=tips
}

// Judging whether it is equal to other
func (gs *GraphState) IsEqual(other *GraphState) bool {
	if gs==other {
		return true
	}
	if gs.layer!=other.layer || gs.total != other.total {
		return false
	}
	return gs.tips.IsEqual(other.tips)
}

// Setting vaules from other
func (gs *GraphState) Equal(other *GraphState) {
	if gs.IsEqual(other) {
		return
	}
	gs.tips=other.tips.Clone()
	gs.layer=other.layer
	gs.total=other.total
}

// Copy self and return
func (gs *GraphState) Clone() *GraphState {
	result:=NewGraphState()
	result.Equal(gs)
	return result
}

// Return one string contain info
func (gs *GraphState) String() string {
	return fmt.Sprintf("(%d,%d,%d)",gs.tips.Size(),gs.total,gs.layer)
}

// Judging whether it is better than other
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

// Encode itself to bytes buff
func (gs *GraphState) Encode(w io.Writer,pver uint32) error {
	err:= s.WriteVarInt(w, pver, uint64(gs.total))
	if err != nil {
		return err
	}
	err= s.WriteVarInt(w, pver, uint64(gs.layer))
	if err != nil {
		return err
	}
	err= s.WriteVarInt(w, pver, uint64(gs.tips.Size()))
	if err != nil {
		return err
	}

	for k:= range gs.tips.GetMap() {
		err = s.WriteElements(w, &k)
		if err != nil {
			return err
		}
	}

	return nil
}

// Decode itself from bytes buff
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

func (gs *GraphState) MaxPayloadLength() uint32 {
	return 8 + 4 + (MaxTips * hash.HashSize)
}

// Create a new GraphState
func NewGraphState() *GraphState {
	return &GraphState{
		tips:NewHashSet(),
		total:0,
		layer:0,
	}
}

