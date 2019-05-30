package blockdag

import (
	"qitmeer/common/hash"
	"sort"
)

type HashSet struct {
	m map[hash.Hash]interface{}
}

func (s *HashSet) GetMap() map[hash.Hash]interface{} {
	return s.m
}

func (s *HashSet) Add(elem *hash.Hash) {
	s.m[*elem] = Empty{}
}

func (s *HashSet) AddPair(elem *hash.Hash,data interface{}) {
	s.m[*elem] = data
}

func (s *HashSet) Remove(elem *hash.Hash) {
	delete(s.m, *elem)
}

func (s *HashSet) AddSet(other *HashSet) {
	if other == nil || other.Len() == 0 {
		return
	}
	for k, v := range other.GetMap() {
		s.AddPair(&k,v)
	}
}

func (s *HashSet) RemoveSet(other *HashSet) {
	if other == nil || other.Len() == 0 {
		return
	}
	for k:= range other.GetMap() {
		s.Remove(&k)
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

// return union of a set
func (s *HashSet) Union(other *HashSet) *HashSet {
	result := s.Clone()
	if s != other {
		result.AddSet(other)
	}
	return result
}

// This function returns a new open memory (HashSet)
// The intersection of a set
func (s *HashSet) Intersection(other *HashSet) *HashSet {
	result := NewHashSet()
	if s == other {
		result.AddSet(s)
	} else {
		if other != nil && other.Len() > 0 {
			for k:= range other.GetMap() {
				if s.Has(&k) {
					result.Add(&k)
				}
			}
		}
	}
	return result
}



func (s *HashSet) Exclude(other *HashSet) {
	if other != nil && other.Len() > 0 {
		for k:= range other.GetMap() {
			if s.Has(&k) {
				s.Remove(&k)
			}
		}
	}
}

func (s *HashSet) Has(elem *hash.Hash) bool {
	_, ok := s.m[*elem]
	return ok
}

func (s *HashSet) HasOnly(elem *hash.Hash) bool {
	return s.Len() == 1 && s.Has(elem)
}

func (s *HashSet) Len() int {
	return len(s.List())
}

func (s *HashSet) Clear() {
	s.m = map[hash.Hash]interface{}{}
}

func (s *HashSet) IsEmpty() bool {
	return s.Len()==0
}

func (s *HashSet) List() []*hash.Hash {
	list := []*hash.Hash{}
	for item:= range s.m {
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
	for k= range s.m {
		if !other.Has(&k) {
			return false
		}
	}
	for k= range other.m {
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
	for k:= range other.GetMap() {
		if !s.Has(&k) {
			return false
		}
	}
	return true
}

// return a new copy
func (s *HashSet) Clone() *HashSet {
	result := NewHashSet()
	result.AddSet(s)
	return result
}

func (s *HashSet) Clean() {
	s.m=map[hash.Hash]interface{}{}
}

// Create a new HashSet
func NewHashSet() *HashSet {
	return &HashSet{
		m: map[hash.Hash]interface{}{},
	}
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
