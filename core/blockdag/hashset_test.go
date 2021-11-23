package blockdag

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"testing"
)

func Test_Add(t *testing.T) {
	hs := NewHashSet()
	hs.Add(&hash.ZeroHash)

	if !hs.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}

func Test_AddSet(t *testing.T) {
	hs := NewHashSet()
	other := NewHashSet()
	other.Add(&hash.ZeroHash)

	hs.AddSet(other)
	if !hs.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}

func Test_AddPair(t *testing.T) {
	var intData int = 123
	hs := NewHashSet()
	hs.AddPair(&hash.ZeroHash, int(intData))

	if !hs.Has(&hash.ZeroHash) || hs.Get(&hash.ZeroHash).(int) != intData {
		t.FailNow()
	}
}

func Test_Remove(t *testing.T) {
	hs := NewHashSet()
	hs.Add(&hash.ZeroHash)
	hs.Remove(&hash.ZeroHash)

	if hs.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}

func Test_RemoveSet(t *testing.T) {
	hs := NewHashSet()
	other := NewHashSet()
	other.Add(&hash.ZeroHash)

	hs.AddSet(other)
	hs.RemoveSet(other)

	if hs.Has(&hash.ZeroHash) {
		t.FailNow()
	}
}

func Test_SortList(t *testing.T) {
	hs := NewHashSet()
	hl := HashSlice{}
	var hashNum int = 5
	for i := 0; i < hashNum; i++ {
		hashStr := fmt.Sprintf("%d", i)
		h := hash.MustHexToDecodedHash(hashStr)
		hs.Add(&h)
		hl = append(hl, &h)
	}
	shs := hs.SortList(false)

	for i := 0; i < hashNum; i++ {
		if !hl[i].IsEqual(shs[i]) {
			t.FailNow()
		}
	}
	rshs := hs.SortList(true)

	for i := 0; i < hashNum; i++ {
		if !hl[i].IsEqual(rshs[hashNum-i-1]) {
			t.FailNow()
		}
	}
}

func Test_For(t *testing.T) {
	hs := NewHashSet()
	var hashNum int = 5
	for i := 0; i < hashNum; i++ {
		hashStr := fmt.Sprintf("%d", i)
		h := hash.MustHexToDecodedHash(hashStr)
		hs.AddPair(&h, hashStr)
	}
	for k, v := range hs.GetMap() {
		fmt.Printf("%s - %s\n", v.(string), k)
	}
}
