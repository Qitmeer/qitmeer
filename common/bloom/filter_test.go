package bloom_test

import (
	"bytes"
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/common/bloom"
	chainhash "github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
	"testing"
)

// TestFilterLarge ensures a maximum sized filter can be created.
func TestFilterLarge(t *testing.T) {
	f := bloom.NewFilter(100000000, 0, 0.01, types.BloomUpdateNone)
	if len(f.MsgFilterLoad().Filter) > types.MaxFilterLoadFilterSize {
		t.Errorf("TestFilterLarge test failed: %d > %d",
			len(f.MsgFilterLoad().Filter), types.MaxFilterLoadFilterSize)
	}
}

// TestFilterLoad ensures loading and unloading of a filter pass.
func TestFilterLoad(t *testing.T) {
	merkle := types.MsgFilterLoad{}

	f := bloom.LoadFilter(&merkle)
	if !f.IsLoaded() {
		t.Errorf("TestFilterLoad IsLoaded test failed: want %v got %v",
			true, !f.IsLoaded())
		return
	}
	f.Unload()
	if f.IsLoaded() {
		t.Errorf("TestFilterLoad IsLoaded test failed: want %v got %v",
			f.IsLoaded(), false)
		return
	}
}

// TestFilterInsert ensures inserting data into the filter causes that data
// to be matched and the resulting serialized MsgFilterLoad is the expected
// value.
func TestFilterInsert(t *testing.T) {
	var tests = []struct {
		hex    string
		insert bool
	}{
		{"99108ad8ed9bb6274d3980bab5a85c048f0950c8", true},
		{"19108ad8ed9bb6274d3980bab5a85c048f0950c8", false},
		{"b5a2c786d9ef4658287ced5914b37a1b4aa32eee", true},
		{"b9300670b4c5366e95b2699e8b18bc75e5f729c5", true},
	}

	f := bloom.NewFilter(3, 0, 0.01, types.BloomUpdateAll)

	for i, test := range tests {
		data, err := hex.DecodeString(test.hex)
		if err != nil {
			t.Errorf("TestFilterInsert DecodeString failed: %v\n", err)
			return
		}
		if test.insert {
			f.Add(data)
		}

		result := f.Matches(data)
		if test.insert != result {
			t.Errorf("TestFilterInsert Matches test #%d failure: got %v want %v\n",
				i, result, test.insert)
			return
		}
	}

	want, err := hex.DecodeString("03614e9b050000000000000001")
	if err != nil {
		t.Errorf("TestFilterInsert DecodeString failed: %v\n", err)
		return
	}

	got := bytes.NewBuffer(nil)
	err = f.MsgFilterLoad().QitmeerEncode(got, protocol.ProtocolVersion, types.BaseEncoding)
	if err != nil {
		t.Errorf("TestFilterInsert BtcDecode failed: %v\n", err)
		return
	}

	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("TestFilterInsert failure: got %v want %v\n",
			got.Bytes(), want)
		return
	}
}

// TestFilterFPRange checks that new filters made with out of range
// false positive targets result in either max or min false positive rates.
func TestFilterFPRange(t *testing.T) {
	tests := []struct {
		name   string
		hash   string
		want   string
		filter *bloom.Filter
	}{
		{
			name:   "fprates > 1 should be clipped at 1",
			hash:   "02981fa052f0481dbc5868f4fc2166035a10f27a03cfd2de67326471df5bc041",
			want:   "00000000000000000001",
			filter: bloom.NewFilter(1, 0, 20.9999999769, types.BloomUpdateAll),
		},
		{
			name:   "fprates less than 1e-9 should be clipped at min",
			hash:   "02981fa052f0481dbc5868f4fc2166035a10f27a03cfd2de67326471df5bc041",
			want:   "0566d97a91a91b0000000000000001",
			filter: bloom.NewFilter(1, 0, 0, types.BloomUpdateAll),
		},
		{
			name:   "negative fprates should be clipped at min",
			hash:   "02981fa052f0481dbc5868f4fc2166035a10f27a03cfd2de67326471df5bc041",
			want:   "0566d97a91a91b0000000000000001",
			filter: bloom.NewFilter(1, 0, -1, types.BloomUpdateAll),
		},
	}

	for _, test := range tests {
		// Convert test input to appropriate types.
		hash, err := chainhash.NewHashFromStr(test.hash)
		if err != nil {
			t.Errorf("NewHashFromStr unexpected error: %v", err)
			continue
		}
		want, err := hex.DecodeString(test.want)
		if err != nil {
			t.Errorf("DecodeString unexpected error: %v\n", err)
			continue
		}

		// Add the test hash to the bloom filter and ensure the
		// filter serializes to the expected bytes.
		f := test.filter
		f.AddHash(hash)
		got := bytes.NewBuffer(nil)
		err = f.MsgFilterLoad().QitmeerEncode(got, protocol.ProtocolVersion, types.BaseEncoding)
		if err != nil {
			t.Errorf("BtcDecode unexpected error: %v\n", err)
			continue
		}
		if !bytes.Equal(got.Bytes(), want) {
			t.Errorf("serialized filter mismatch: got %x want %x\n",
				got.Bytes(), want)
			continue
		}
	}
}

// TestFilterInsert ensures inserting data into the filter with a tweak causes
// that data to be matched and the resulting serialized MsgFilterLoad is the
// expected value.
func TestFilterInsertWithTweak(t *testing.T) {
	var tests = []struct {
		hex    string
		insert bool
	}{
		{"99108ad8ed9bb6274d3980bab5a85c048f0950c8", true},
		{"19108ad8ed9bb6274d3980bab5a85c048f0950c8", false},
		{"b5a2c786d9ef4658287ced5914b37a1b4aa32eee", true},
		{"b9300670b4c5366e95b2699e8b18bc75e5f729c5", true},
	}

	f := bloom.NewFilter(3, 2147483649, 0.01, types.BloomUpdateAll)

	for i, test := range tests {
		data, err := hex.DecodeString(test.hex)
		if err != nil {
			t.Errorf("TestFilterInsertWithTweak DecodeString failed: %v\n", err)
			return
		}
		if test.insert {
			f.Add(data)
		}

		result := f.Matches(data)
		if test.insert != result {
			t.Errorf("TestFilterInsertWithTweak Matches test #%d failure: got %v want %v\n",
				i, result, test.insert)
			return
		}
	}

	want, err := hex.DecodeString("03ce4299050000000100008001")
	if err != nil {
		t.Errorf("TestFilterInsertWithTweak DecodeString failed: %v\n", err)
		return
	}
	got := bytes.NewBuffer(nil)
	err = f.MsgFilterLoad().QitmeerEncode(got, protocol.ProtocolVersion, types.BaseEncoding)
	if err != nil {
		t.Errorf("TestFilterInsertWithTweak BtcDecode failed: %v\n", err)
		return
	}

	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("TestFilterInsertWithTweak failure: got %v want %v\n",
			got.Bytes(), want)
		return
	}
}

func TestFilterBloomMatch(t *testing.T) {
	// genesis coinbase tx bytes
	str := "01000000020000000000000000000000000000000000000000000000000000000000000000ffffffffffffffff0000000000000000000000000000000000000000000000000000000000000000ffffffffffffffff0200000088526a740000001976a91437733b37b9f09ce024a5ffbd4570fc1e242c907a88ac010000743ba40b0000001976a91437733b37b9f09ce024a5ffbd4570fc1e242c907a88ac000000000000000000096e88024d04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b7300"
	strBytes, err := hex.DecodeString(str)
	if err != nil {
		t.Errorf("TestFilterBloomMatch DecodeString failure: %v", err)
		return
	}
	tx, err := types.NewTxFromBytes(strBytes)
	if err != nil {
		t.Errorf("TestFilterBloomMatch NewTxFromBytes failure: %v", err)
		return
	}
	spendingTxBytes, _ := hex.DecodeString("0100000002767b98dc030a071654483cd22ad4265b4dcbcaed92f04b42119da626950330ff00000000ffffffffdce4f5433828893593f175a9c918cd7ae4e7837407a725a995177b884cd6430b00000000ffffffff0200000088526a740000001976a9143208f7eb599914e3418e2def4ee0de8f66fefd9488ac0000ce643ba40b0000001976a914f8328c6b050e67d7c76cd7cf0a2673c38a78ce6388ac000000000000000056197c60026b483045022100c57ba688a867f3c10883a0c995a7983009a48fce4e0a2f24bb2679f602af24b102200e874322c7631ff710abc943b4c48c3d877c22bd4bdcfd3fc7c2f05c118b8f480121023c6682e0628c701ead657b43c01c316c06d15cd0090e82f14caa004e414217916a47304402202c54b92007c2be1bdc16d6affd8e9ae274d1a346cd72360cc1c5e349f3e7507f02200f8227d35c4a71f41bc7d63a64bbea10697daedd1f5f289e24993535d571744c0121023c6682e0628c701ead657b43c01c316c06d15cd0090e82f14caa004e41421791")

	spendingTx, err := types.NewTxFromBytes(spendingTxBytes)
	if err != nil {
		t.Errorf("TestFilterBloomMatch NewTxFromBytes failure: %v", err)
		return
	}

	f := bloom.NewFilter(10, 0, 0.000001, types.BloomUpdateAll)
	// pubkey
	inputStr := "37733b37b9f09ce024a5ffbd4570fc1e242c907a"
	hashBytes, err := hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterBloomMatch DecodeString failed: %v\n", err)
		return
	}
	f.Add(hashBytes)
	if !f.MatchTxAndUpdate(tx) {
		t.Errorf("TestFilterBloomMatch didn't match hash %s", inputStr)
	}
	f = bloom.NewFilter(10, 0, 0.000001, types.BloomUpdateAll)
	outpoint := types.NewOutPoint(tx.Hash(), 0)
	f.AddOutPoint(outpoint)
	if !f.MatchTxAndUpdate(spendingTx) {
		t.Errorf("TestFilterBloomMatch matched outpoint %s", inputStr)
	}

}

func TestFilterInsertUpdateNone(t *testing.T) {
	f := bloom.NewFilter(10, 0, 0.000001, types.BloomUpdateNone)

	// Add the generation pubkey
	inputStr := "04eaafc2314def4ca98ac970241bcab022b9c1e1f4ea423a20f134c" +
		"876f2c01ec0f0dd5b2e86e7168cefe0d81113c3807420ce13ad1357231a" +
		"2252247d97a46a91"
	inputBytes, err := hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertUpdateNone DecodeString failed: %v", err)
		return
	}
	f.Add(inputBytes)

	// Add the output address for the 4th transaction
	inputStr = "b6efd80d99179f4f4ff6f4dd0a007d018c385d21"
	inputBytes, err = hex.DecodeString(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertUpdateNone DecodeString failed: %v", err)
		return
	}
	f.Add(inputBytes)

	inputStr = "147caa76786596590baa4e98f5d9f48b86c7765e489f7a6ff3360fe5c674360b"
	hash, err := chainhash.NewHashFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertUpdateNone NewHashFromStr failed: %v", err)
		return
	}
	outpoint := types.NewOutPoint(hash, 0)

	if f.MatchesOutPoint(outpoint) {
		t.Errorf("TestFilterInsertUpdateNone matched outpoint %s", inputStr)
		return
	}

	inputStr = "02981fa052f0481dbc5868f4fc2166035a10f27a03cfd2de67326471df5bc041"
	hash, err = chainhash.NewHashFromStr(inputStr)
	if err != nil {
		t.Errorf("TestFilterInsertUpdateNone NewHashFromStr failed: %v", err)
		return
	}
	outpoint = types.NewOutPoint(hash, 0)

	if f.MatchesOutPoint(outpoint) {
		t.Errorf("TestFilterInsertUpdateNone matched outpoint %s", inputStr)
		return
	}
}

func TestFilterReload(t *testing.T) {
	f := bloom.NewFilter(10, 0, 0.000001, types.BloomUpdateAll)

	bFilter := bloom.LoadFilter(f.MsgFilterLoad())
	if bFilter.MsgFilterLoad() == nil {
		t.Errorf("TestFilterReload LoadFilter test failed")
		return
	}
	bFilter.Reload(nil)

	if bFilter.MsgFilterLoad() != nil {
		t.Errorf("TestFilterReload Reload test failed")
	}
}
