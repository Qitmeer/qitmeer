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

func Test_AddPair(t *testing.T) {
	var intData int=123
	hs:=NewHashSet()
	hs.AddPair(&hash.ZeroHash,int(intData))

	if !hs.Has(&hash.ZeroHash) || hs.Get(&hash.ZeroHash).(int)!=intData{
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