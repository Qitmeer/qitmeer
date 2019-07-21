// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package util

// timeSorter implements sort.Interface to allow a slice of timestamps to
// be sorted.
type TimeSorter []int64

// Len returns the number of timestamps in the slice.  It is part of the
// sort.Interface implementation.
func (s TimeSorter) Len() int {
	return len(s)
}

// Swap swaps the timestamps at the passed indices.  It is part of the
// sort.Interface implementation.
func (s TimeSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less returns whether the timstamp with index i should sort before the
// timestamp with index j.  It is part of the sort.Interface implementation.
func (s TimeSorter) Less(i, j int) bool {
	return s[i] < s[j]
}
