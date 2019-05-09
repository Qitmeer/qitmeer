package blockdag

import (
	"qitmeer/common/hash"
	"testing"
)

func Test_Add(t *testing.T) {
	hs:=NewHashSet()
	hs.Add(&hash.ZeroHash)

	if !hs.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}

func Test_AddSet(t *testing.T) {
	hs:=NewHashSet()
	other:=NewHashSet()
	other.Add(&hash.ZeroHash)

	hs.AddSet(other)
	if !hs.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}

func Test_Remove(t *testing.T) {
	hs:=NewHashSet()
	hs.Add(&hash.ZeroHash)
	hs.Remove(&hash.ZeroHash)

	if hs.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}

func Test_RemoveSet(t *testing.T) {
	hs:=NewHashSet()
	other:=NewHashSet()
	other.Add(&hash.ZeroHash)

	hs.AddSet(other)
	hs.RemoveSet(other)

	if hs.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}