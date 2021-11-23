package blockdag

import (
	"github.com/Qitmeer/qng-core/common/hash"
)

//A collection that tries to imitate "lazy" operations
type LazySet struct {
	sets            []*HashSet
	positiveIndices map[int]Empty
}

func (ls *LazySet) update(other *HashSet) {
	if other == nil {
		return
	}
	ls.positiveIndices[len(ls.sets)] = Empty{}
	ls.sets = append(ls.sets, other)
}

func (ls *LazySet) differenceUpdate(other *HashSet) {
	if other == nil {
		return
	}
	ls.sets = append(ls.sets, other)
}

// The set is composed of all self and other elements.
func (ls *LazySet) Union(other *HashSet) *LazySet {
	result := ls.Clone()
	result.update(other)
	return result
}

// A collection consisting of all elements belonging to set self and other.
func (ls *LazySet) Intersection(other *HashSet) *LazySet {
	bs := ls.flattenToSet()
	if bs == nil || other == nil {
		return nil
	}
	is := bs.Intersection(other)
	result := NewLazySet()
	result.update(is)
	return result
}

// A collection of all elements that belong to self and not to other.
func (ls *LazySet) difference(other *HashSet) *LazySet {
	result := ls.Clone()
	result.differenceUpdate(other)
	return result
}

// Return a new copy
func (ls *LazySet) Clone() *LazySet {
	result := NewLazySet()
	result.sets = make([]*HashSet, len(ls.sets))
	copy(result.sets, ls.sets)
	for k := range ls.positiveIndices {
		result.positiveIndices[k] = Empty{}
	}
	return result

}

func (ls *LazySet) flattenToSet() *HashSet {
	baseSetIndex := -1
	for index := range ls.sets {
		if _, ok := ls.positiveIndices[index]; ok {
			baseSetIndex = index
		}
	}
	if baseSetIndex < 0 {
		return nil
	}
	baseSet := ls.sets[baseSetIndex].Clone()
	lenSets := len(ls.sets)
	for index := baseSetIndex + 1; index < lenSets; index++ {
		if _, ok := ls.positiveIndices[index]; ok {
			baseSet.AddSet(ls.sets[index])
		} else {
			baseSet.RemoveSet(ls.sets[index])
		}
	}
	return baseSet
}

func (ls *LazySet) Clear() {
	ls.sets = []*HashSet{}
	ls.positiveIndices = map[int]Empty{}
}

func (ls *LazySet) Add(elem *hash.Hash) {

}

func (ls *LazySet) Remove(elem *hash.Hash) {

}

// Create a new LazySet
func NewLazySet() *LazySet {
	return &LazySet{
		sets:            []*HashSet{},
		positiveIndices: map[int]Empty{},
	}
}
