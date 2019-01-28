package blockdag

import (
	"github.com/noxproject/nox/common/hash"
	"sort"
)

type BlockSet struct {
	m map[hash.Hash]bool
}

func NewBlockSet() *BlockSet {
	return &BlockSet{
		m: map[hash.Hash]bool{},
	}
}
func (s *BlockSet) GetMap() map[hash.Hash]bool {
	return s.m
}
func (s *BlockSet) Add(item *hash.Hash) {
	s.m[*item] = true
}
func (s *BlockSet) AddSet(other *BlockSet) {
	if other == nil || other.Len() == 0 {
		return
	}
	for k, _ := range other.GetMap() {
		s.Add(&k)
	}
}
func (s *BlockSet) AddList(list []*hash.Hash) {
	if len(list) == 0 {
		return
	}
	for _,v:= range list {
		s.Add(v)
	}
}

/*This function returns a new open memory (BlockSet)*/
func (s *BlockSet) Intersection(other *BlockSet) *BlockSet {
	result := NewBlockSet()
	if s == other {
		result.AddSet(s)
	} else {
		if other != nil && other.Len() > 0 {
			for k, _ := range other.GetMap() {
				if s.Has(&k) {
					result.Add(&k)
				}
			}
		}
	}
	return result
}

func (s *BlockSet) Remove(item *hash.Hash) {
	delete(s.m, *item)
}
func (s *BlockSet) Exclude(other *BlockSet) {
	if other != nil && other.Len() > 0 {
		for k, _ := range other.GetMap() {
			if s.Has(&k) {
				s.Remove(&k)
			}
		}
	}
}

func (s *BlockSet) Has(item *hash.Hash) bool {
	_, ok := s.m[*item]
	return ok
}
func (s *BlockSet) HasOnly(item *hash.Hash) bool {
	return s.Len() == 1 && s.Has(item)
}

func (s *BlockSet) Len() int {
	return len(s.List())
}

func (s *BlockSet) Clear() {
	s.m = map[hash.Hash]bool{}
}

func (s *BlockSet) IsEmpty() bool {
	if s.Len() == 0 {
		return true
	}
	return false
}

func (s *BlockSet) List() []*hash.Hash {
	list := []*hash.Hash{}
	for item,_ := range s.m {
		kv:=item
		list = append(list, &kv)
	}
	return list
}

func (s *BlockSet) OrderList() []*hash.Hash {
	list := SortHashs(s.List())
	sort.Sort(list)
	return []*hash.Hash(list)
}

func (s *BlockSet) IsEqual(other *BlockSet) bool {
	var k hash.Hash
	for k, _ = range s.m {
		if !other.Has(&k) {
			return false
		}
	}
	for k, _ = range other.m {
		if !s.Has(&k) {
			return false
		}
	}

	return true
}
func (s *BlockSet) Contain(other *BlockSet) bool {
	if other.IsEmpty() {
		return false
	}
	for k, _ := range other.GetMap() {
		if !s.Has(&k) {
			return false
		}
	}
	return true
}
func (s *BlockSet) Clone() *BlockSet {
	result := NewBlockSet()
	result.AddSet(s)
	return result
}
func GetMaxLenBlockSet(bsm map[hash.Hash]*BlockSet) *hash.Hash {

	var result hash.Hash
	var curNum int = 0
	for k, v := range bsm {
		if curNum == 0 {
			result = k
			curNum = v.Len()
		} else {
			if v.Len() > curNum {
				curNum = v.Len()
				result = k
			} else if v.Len() == curNum {
				if k.String() < result.String() {
					result = k
				}
			}
		}

	}
	return &result
}

// SortHashs is used to sort hash list
type SortHashs []*hash.Hash

func (sh SortHashs) Len() int {
	return len(sh)
}

func (sh SortHashs) Less(i, j int) bool {
	return sh[i].String() < sh[j].String()
}

func (sh SortHashs) Swap(i, j int) {
	sh[i], sh[j] = sh[j], sh[i]
}
