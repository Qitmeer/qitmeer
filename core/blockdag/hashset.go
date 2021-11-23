package blockdag

import (
	"github.com/Qitmeer/qng-core/common/hash"
	"sort"
)

// On the Set of hash, and the saved data can be of any type
type HashSet struct {
	m map[hash.Hash]interface{}
}

// Return the map
func (s *HashSet) GetMap() map[hash.Hash]interface{} {
	return s.m
}

// Add the key of element
func (s *HashSet) Add(elem *hash.Hash) {
	s.m[*elem] = Empty{}
}

// Add one pair of data
func (s *HashSet) AddPair(elem *hash.Hash, data interface{}) {
	s.m[*elem] = data
}

// Remove the element
func (s *HashSet) Remove(elem *hash.Hash) {
	delete(s.m, *elem)
}

// Add elements from set
func (s *HashSet) AddSet(other *HashSet) {
	if other == nil || other.Size() == 0 {
		return
	}
	for k, v := range other.GetMap() {
		s.AddPair(&k, v)
	}
}

func (s *HashSet) RemoveSet(other *HashSet) {
	if other == nil || other.Size() == 0 {
		return
	}
	for k := range other.GetMap() {
		s.Remove(&k)
	}
}

func (s *HashSet) AddList(list []*hash.Hash) {
	if len(list) == 0 {
		return
	}
	for _, v := range list {
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
		if other != nil && other.Size() > 0 {
			for k := range other.GetMap() {
				if s.Has(&k) {
					result.Add(&k)
				}
			}
		}
	}
	return result
}

func (s *HashSet) Exclude(other *HashSet) {
	if other != nil && other.Size() > 0 {
		for k := range other.GetMap() {
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
	return s.Size() == 1 && s.Has(elem)
}

func (s *HashSet) Get(elem *hash.Hash) interface{} {
	return s.m[*elem]
}

func (s *HashSet) Size() int {
	return len(s.m)
}

func (s *HashSet) IsEmpty() bool {
	return s.Size() == 0
}

func (s *HashSet) List() []*hash.Hash {
	list := []*hash.Hash{}
	for item := range s.m {
		kv := item
		list = append(list, &kv)
	}
	return list
}

func (s *HashSet) SortList(reverse bool) []*hash.Hash {
	list := HashSlice(s.List())
	if reverse {
		sort.Sort(sort.Reverse(list))
	} else {
		sort.Sort(list)
	}
	return []*hash.Hash(list)
}

func (s *HashSet) IsEqual(other *HashSet) bool {
	var k hash.Hash
	for k = range s.m {
		if !other.Has(&k) {
			return false
		}
	}
	for k = range other.m {
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
	for k := range other.GetMap() {
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
	s.m = map[hash.Hash]interface{}{}
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
			curNum = v.Size()
		} else {
			if v.Size() > curNum {
				curNum = v.Size()
				result = k
			} else if v.Size() == curNum {
				if k.String() < result.String() {
					result = k
				}
			}
		}

	}
	return &result
}

// This struct is empty
type Empty struct{}
