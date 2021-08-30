/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:idset.go
 * Date:3/29/20 8:56 PM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package blockdag

import (
	"sort"
)

// On the Set of hash, and the saved data can be of any type
type IdSet struct {
	m map[uint]interface{}
}

// Return the map
func (s *IdSet) GetMap() map[uint]interface{} {
	return s.m
}

// Add the key of element
func (s *IdSet) Add(elem uint) {
	s.m[elem] = Empty{}
}

// Add one pair of data
func (s *IdSet) AddPair(elem uint, data interface{}) {
	s.m[elem] = data
}

// Remove the element
func (s *IdSet) Remove(elem uint) {
	delete(s.m, elem)
}

// Add elements from set
func (s *IdSet) AddSet(other *IdSet) {
	if other == nil || other.Size() == 0 {
		return
	}
	for k, v := range other.GetMap() {
		s.AddPair(k, v)
	}
}

func (s *IdSet) RemoveSet(other *IdSet) {
	if other == nil || other.Size() == 0 {
		return
	}
	for k := range other.GetMap() {
		s.Remove(k)
	}
}

func (s *IdSet) AddList(list []uint) {
	if len(list) == 0 {
		return
	}
	for _, v := range list {
		s.Add(v)
	}
}

// return union of a set
func (s *IdSet) Union(other *IdSet) *IdSet {
	result := s.Clone()
	if s != other {
		result.AddSet(other)
	}
	return result
}

// This function returns a new open memory (IdSet)
// The intersection of a set
func (s *IdSet) Intersection(other *IdSet) *IdSet {
	result := NewIdSet()
	if s == other {
		result.AddSet(s)
	} else {
		if other != nil && other.Size() > 0 {
			for k, v := range other.GetMap() {
				if s.Has(k) {
					result.AddPair(k, v)
				}
			}
		}
	}
	return result
}

func (s *IdSet) Exclude(other *IdSet) {
	if other != nil && other.Size() > 0 {
		for k := range other.GetMap() {
			if s.Has(k) {
				s.Remove(k)
			}
		}
	}
}

func (s *IdSet) Has(elem uint) bool {
	_, ok := s.m[elem]
	return ok
}

func (s *IdSet) HasOnly(elem uint) bool {
	return s.Size() == 1 && s.Has(elem)
}

func (s *IdSet) Get(elem uint) interface{} {
	return s.m[elem]
}

func (s *IdSet) Size() int {
	return len(s.m)
}

func (s *IdSet) IsEmpty() bool {
	return s.Size() == 0
}

func (s *IdSet) List() []uint {
	list := []uint{}
	for item := range s.m {
		kv := item
		list = append(list, kv)
	}
	return list
}

func (s *IdSet) SortList(reverse bool) []uint {
	list := IdSlice(s.List())
	if reverse {
		sort.Sort(sort.Reverse(list))
	} else {
		sort.Sort(list)
	}
	return []uint(list)
}

// Value must be ensured
func (s *IdSet) SortHashList(reverse bool) []uint {
	list := BlockHashSlice{}
	for _, v := range s.m {
		item := v.(IBlock)
		list = append(list, item)
	}
	if reverse {
		sort.Sort(sort.Reverse(list))
	} else {
		sort.Sort(list)
	}

	result := []uint{}
	for _, v := range list {
		result = append(result, v.GetID())
	}
	return result
}

func (s *IdSet) SortPriorityList(reverse bool) []uint {
	list := BlockPrioritySlice{}
	for _, v := range s.m {
		item := v.(IBlock)
		list = append(list, item)
	}
	if reverse {
		sort.Sort(sort.Reverse(list))
	} else {
		sort.Sort(list)
	}

	result := []uint{}
	for _, v := range list {
		result = append(result, v.GetID())
	}
	return result
}

func (s *IdSet) IsEqual(other *IdSet) bool {
	var k uint
	for k = range s.m {
		if !other.Has(k) {
			return false
		}
	}
	for k = range other.m {
		if !s.Has(k) {
			return false
		}
	}

	return true
}

func (s *IdSet) Contain(other *IdSet) bool {
	if other.IsEmpty() {
		return false
	}
	for k := range other.GetMap() {
		if !s.Has(k) {
			return false
		}
	}
	return true
}

// return a new copy
func (s *IdSet) Clone() *IdSet {
	result := NewIdSet()
	result.AddSet(s)
	return result
}

func (s *IdSet) Clean() {
	s.m = map[uint]interface{}{}
}

// Create a new IdSet
func NewIdSet() *IdSet {
	return &IdSet{
		m: map[uint]interface{}{},
	}
}
