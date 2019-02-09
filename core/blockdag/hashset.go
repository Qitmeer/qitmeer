package blockdag

import (
	"github.com/noxproject/nox/common/hash"
	"sort"
)

type HashSet struct {
	m map[hash.Hash]Empty
}

func NewHashSet() *HashSet {
	return &HashSet{
		m: map[hash.Hash]Empty{},
	}
}

func (s *HashSet) GetMap() map[hash.Hash]Empty {
	return s.m
}

func (s *HashSet) Add(item *hash.Hash) {
	s.m[*item] = Empty{}
}

func (s *HashSet) AddSet(other *HashSet) {
	if other == nil || other.Len() == 0 {
		return
	}
	for k, _ := range other.GetMap() {
		s.Add(&k)
	}
}

func (s *HashSet) AddList(list []*hash.Hash) {
	if len(list) == 0 {
		return
	}
	for _,v:= range list {
		s.Add(v)
	}
}

/*This function returns a new open memory (HashSet)*/
func (s *HashSet) Intersection(other *HashSet) *HashSet {
	result := NewHashSet()
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

func (s *HashSet) Remove(item *hash.Hash) {
	delete(s.m, *item)
}

func (s *HashSet) Exclude(other *HashSet) {
	if other != nil && other.Len() > 0 {
		for k, _ := range other.GetMap() {
			if s.Has(&k) {
				s.Remove(&k)
			}
		}
	}
}

func (s *HashSet) Has(item *hash.Hash) bool {
	_, ok := s.m[*item]
	return ok
}

func (s *HashSet) HasOnly(item *hash.Hash) bool {
	return s.Len() == 1 && s.Has(item)
}

func (s *HashSet) Len() int {
	return len(s.List())
}

func (s *HashSet) Clear() {
	s.m = map[hash.Hash]Empty{}
}

func (s *HashSet) IsEmpty() bool {
	if s.Len() == 0 {
		return true
	}
	return false
}

func (s *HashSet) List() []*hash.Hash {
	list := []*hash.Hash{}
	for item,_ := range s.m {
		kv:=item
		list = append(list, &kv)
	}
	return list
}

func (s *HashSet) SortList() []*hash.Hash {
	list := SortHashs(s.List())
	sort.Sort(list)
	return []*hash.Hash(list)
}

func (s *HashSet) IsEqual(other *HashSet) bool {
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
func (s *HashSet) Contain(other *HashSet) bool {
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
func (s *HashSet) Clone() *HashSet {
	result := NewHashSet()
	result.AddSet(s)
	return result
}
func GetMaxLenHashSet(bsm map[hash.Hash]*HashSet) *hash.Hash {

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

// This struct is empty
type Empty struct {}
