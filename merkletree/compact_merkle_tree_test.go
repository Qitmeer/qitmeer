// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package merkle

import (
	"bytes"
	"encoding/hex"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"github.com/kylelemons/godebug/pretty"

	"github.com/dindinw/dagproject/merkletree/rfc6962"
)

// MustHexDecode decodes its input string from hex and panics if this fails
func MustHexDecode(b string) []byte {
	r, err := hex.DecodeString(b)
	if err != nil {
		panic(err)
	}
	return r
}

// MustDecodeBase64 expects a base 64 encoded string input and panics if it cannot be decoded
func MustDecodeBase64(b64 string) []byte {
	r, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		panic(r)
	}
	return r
}

// MerkleTreeLeafTestInputs returns a slice of leaf inputs that may be used in
// compact Merkle tree test cases.  They are intended to be added successively,
// so that after each addition the corresponding root from MerkleTreeLeafTestRoots
// gives the expected Merkle tree root hash.
func MerkleTreeLeafTestInputs() [][]byte {
	return [][]byte{
		[]byte(""), []byte("\x00"), []byte("\x10"), []byte("\x20\x21"), []byte("\x30\x31"),
		[]byte("\x40\x41\x42\x43"), []byte("\x50\x51\x52\x53\x54\x55\x56\x57"),
		[]byte("\x60\x61\x62\x63\x64\x65\x66\x67\x68\x69\x6a\x6b\x6c\x6d\x6e\x6f")}
}

// MerkleTreeLeafTestRootHashes returns a slice of Merkle tree root hashes that
// correspond to the expected tree state for the leaf additions returned by
// MerkleTreeLeafTestInputs(), as described above.
func MerkleTreeLeafTestRootHashes() [][]byte {
	return [][]byte{
		// constants from C++ test: https://github.com/google/certificate-transparency/blob/master/cpp/merkletree/merkle_tree_test.cc#L277
		MustHexDecode("6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"),
		MustHexDecode("fac54203e7cc696cf0dfcb42c92a1d9dbaf70ad9e621f4bd8d98662f00e3c125"),
		MustHexDecode("aeb6bcfe274b70a14fb067a5e5578264db0fa9b51af5e0ba159158f329e06e77"),
		MustHexDecode("d37ee418976dd95753c1c73862b9398fa2a2cf9b4ff0fdfe8b30cd95209614b7"),
		MustHexDecode("4e3bbb1f7b478dcfe71fb631631519a3bca12c9aefca1612bfce4c13a86264d4"),
		MustHexDecode("76e67dadbcdf1e10e1b74ddc608abd2f98dfb16fbce75277b5232a127f2087ef"),
		MustHexDecode("ddb89be403809e325750d3d263cd78929c2942b7942a34b77e122c9594a74c8c"),
		MustHexDecode("5dc9da79a70659a9ad559cb701ded9a2ab9d823aad2f4960cfe370eff4604328")}
}

// CompactMerkleTreeLeafTestNodeHashes returns the CompactMerkleTree.node state
// that must result after each of the leaf additions returned by
// MerkleTreeLeafTestInputs(), as described above.
func CompactMerkleTreeLeafTestNodeHashes() [][][]byte {
	return [][][]byte{
		nil, // perfect tree size, 2^0
		nil, // perfect tree size, 2^1
		{MustDecodeBase64("ApjRIpBtz8EIkstTpzmS/FufST6kybrbJ7eRtBJ6f+c="), MustDecodeBase64("+sVCA+fMaWzw38tCySodnbr3CtnmIfS9jZhmLwDjwSU=")},
		nil, // perfect tree size, 2^2
		{MustDecodeBase64("vBoGQ7EuTS18d5GPROD095qDi2z57FtcKD4fTYhZnms="), nil, MustDecodeBase64("037kGJdt2VdTwcc4Yrk5j6Kiz5tP8P3+izDNlSCWFLc=")},
		{nil, MustDecodeBase64("DrxdNDf74tsVi58Sah0RjjCBgQMdCpSfje3t68VY72o="), MustDecodeBase64("037kGJdt2VdTwcc4Yrk5j6Kiz5tP8P3+izDNlSCWFLc=")},
		{MustDecodeBase64("sIaT7C5yFZcTBkHoIR5+7cy0wmQTlj7ubB4u0W/7Gl8="), MustDecodeBase64("DrxdNDf74tsVi58Sah0RjjCBgQMdCpSfje3t68VY72o="), MustDecodeBase64("037kGJdt2VdTwcc4Yrk5j6Kiz5tP8P3+izDNlSCWFLc=")},
		nil, // perfect tree size, 2^3
	}
}

// EmptyMerkleTreeRootHash returns the expected root hash for an empty Merkle Tree
// that uses SHA256 hashing.
func EmptyMerkleTreeRootHash() []byte {
	const sha256EmptyTreeHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	return MustHexDecode(sha256EmptyTreeHash)
}

func checkUnusedNodesInvariant(c *CompactMerkleTree) error {
	// The structure of this invariant check mirrors the structure in
	// NewCompactMerkleTreeWithState in which only the nodes which
	// should be present for a tree of given size are fetched from the
	// backing store via GetNodeFunc.
	size := c.size
	sizeBits := bitLen(size)
	if isPerfectTree(size) {
		for i, n := range c.nodes {
			expectNil := i != sizeBits-1
			if expectNil && n != nil {
				return fmt.Errorf("perfect Tree size %d has non-nil node at index %d, wanted nil", size, i)
			}
			if !expectNil && n == nil {
				return fmt.Errorf("perfect Tree size %d has nil node at index %d, wanted non-nil", size, i)
			}
		}
	} else {
		for depth := 0; depth < sizeBits; depth++ {
			if size&1 == 1 {
				if c.nodes[depth] == nil {
					return fmt.Errorf("imperfect Tree size %d has nil node at index %d, wanted non-nil", c.size, depth)
				}
			} else {
				if c.nodes[depth] != nil {
					return fmt.Errorf("imperfect Tree size %d has non-nil node at index %d, wanted nil", c.size, depth)
				}
			}
			size >>= 1
		}
	}
	return nil
}

func TestAddingLeaves(t *testing.T) {
	inputs := MerkleTreeLeafTestInputs()
	roots := MerkleTreeLeafTestRootHashes()
	hashes := CompactMerkleTreeLeafTestNodeHashes()

	// We test the "same" thing 3 different ways this is to ensure than any lazy
	// update strategy being employed by the implementation doesn't affect the
	// api-visible calculation of root & size.
	{
		// First tree, add nodes one-by-one
		tree := NewCompactMerkleTree(rfc6962.DefaultHasher)
		if got, want := tree.Size(), int64(0); got != want {
			t.Errorf("Size()=%d, want %d", got, want)
		}
		if got, want := tree.CurrentRoot(), EmptyMerkleTreeRootHash(); !bytes.Equal(got, want) {
			t.Errorf("CurrentRoot()=%x, want %x", got, want)
		}

		for i := 0; i < 8; i++ {
			tree.AddLeaf(inputs[i], func(int, int64, []byte) error {
				return nil
			})
			if err := checkUnusedNodesInvariant(tree); err != nil {
				t.Fatalf("UnusedNodesInvariant check failed: %v", err)
			}
			if got, want := tree.Size(), int64(i+1); got != want {
				t.Errorf("Size()=%d, want %d", got, want)
			}
			if got, want := tree.CurrentRoot(), roots[i]; !bytes.Equal(got, want) {
				t.Errorf("CurrentRoot()=%v, want %v", got, want)
			}
			if diff := pretty.Compare(tree.Hashes(), hashes[i]); diff != "" {
				t.Errorf("post-Hashes() diff:\n%v", diff)
			}
		}
	}

	{
		// Second tree, add nodes all at once
		tree := NewCompactMerkleTree(rfc6962.DefaultHasher)
		for i := 0; i < 8; i++ {
			tree.AddLeaf(inputs[i], func(int, int64, []byte) error {
				return nil
			})
			if err := checkUnusedNodesInvariant(tree); err != nil {
				t.Fatalf("UnusedNodesInvariant check failed: %v", err)
			}
		}
		if got, want := tree.Size(), int64(8); got != want {
			t.Errorf("Size()=%d, want %d", got, want)
		}
		if got, want := tree.CurrentRoot(), roots[7]; !bytes.Equal(got, want) {
			t.Errorf("CurrentRoot()=%v, want %v", got, want)
		}
		if diff := pretty.Compare(tree.Hashes(), hashes[7]); diff != "" {
			t.Errorf("post-Hashes() diff:\n%v", diff)
		}
	}

	{
		// Third tree, add nodes in two chunks
		tree := NewCompactMerkleTree(rfc6962.DefaultHasher)
		for i := 0; i < 3; i++ {
			tree.AddLeaf(inputs[i], func(int, int64, []byte) error {
				return nil
			})
			if err := checkUnusedNodesInvariant(tree); err != nil {
				t.Fatalf("UnusedNodesInvariant check failed: %v", err)
			}
		}
		if got, want := tree.Size(), int64(3); got != want {
			t.Errorf("Size()=%d, want %d", got, want)
		}
		if got, want := tree.CurrentRoot(), roots[2]; !bytes.Equal(got, want) {
			t.Errorf("CurrentRoot()=%v, want %v", got, want)
		}
		if diff := pretty.Compare(tree.Hashes(), hashes[2]); diff != "" {
			t.Errorf("post-Hashes() diff:\n%v", diff)
		}

		for i := 3; i < 8; i++ {
			tree.AddLeaf(inputs[i], func(int, int64, []byte) error {
				return nil
			})
			if err := checkUnusedNodesInvariant(tree); err != nil {
				t.Fatalf("UnusedNodesInvariant check failed: %v", err)
			}
		}
		if got, want := tree.Size(), int64(8); got != want {
			t.Errorf("Size()=%d, want %d", got, want)
		}
		if got, want := tree.CurrentRoot(), roots[7]; !bytes.Equal(got, want) {
			t.Errorf("CurrentRoot()=%v, want %v", got, want)
		}
		if diff := pretty.Compare(tree.Hashes(), hashes[7]); diff != "" {
			t.Errorf("post-Hashes() diff:\n%v", diff)
		}
	}
}

func failingGetNodeFunc(int, int64) ([]byte, error) {
	return []byte{}, errors.New("bang")
}

// This returns something that won't result in a valid root hash match, doesn't really
// matter what it is but it must be correct length for an SHA256 hash as if it was real
func fixedHashGetNodeFunc(int, int64) ([]byte, error) {
	return []byte("12345678901234567890123456789012"), nil
}

func TestLoadingTreeFailsNodeFetch(t *testing.T) {
	_, err := NewCompactMerkleTreeWithState(rfc6962.DefaultHasher, 237, failingGetNodeFunc, []byte("notimportant"))

	if err == nil || !strings.Contains(err.Error(), "bang") {
		t.Errorf("Did not return correctly on failed node fetch: %v", err)
	}
}

func TestLoadingTreeFailsBadRootHash(t *testing.T) {
	// Supply a root hash that can't possibly match the result of the SHA 256 hashing on our dummy
	// data
	_, err := NewCompactMerkleTreeWithState(rfc6962.DefaultHasher, 237, fixedHashGetNodeFunc, []byte("nomatch!nomatch!nomatch!nomatch!"))
	_, ok := err.(RootHashMismatchError)

	if err == nil || !ok {
		t.Errorf("Did not return correct error type on root mismatch: %v", err)
	}
}


// NodeID uniquely identifies a Node within a versioned MerkleTree.
type NodeID struct {
	// path is effectively a BigEndian bit set, with path[0] being the MSB
	// (identifying the root child), and successive bits identifying the lower
	// level children down to the leaf.
	Path []byte
	// PrefixLenBits is the number of MSB in Path which are considered part of
	// this NodeID.
	//
	// e.g. if Path contains two bytes, and PrefixLenBits is 9, then the 8 bits
	// in Path[0] are included, along with the lowest bit of Path[1]
	PrefixLenBits int
}

// PathLenBits returns 8 * len(path).
func (n NodeID) PathLenBits() int {
	return len(n.Path) * 8
}

// String returns a string representation of the binary value of the NodeID.
// The left-most bit is the MSB (i.e. nearer the root of the tree).
func (n *NodeID) String() string {
	var r bytes.Buffer
	limit := n.PathLenBits() - n.PrefixLenBits
	for i := n.PathLenBits() - 1; i >= limit; i-- {
		r.WriteRune(rune('0' + n.Bit(i)))
	}
	return r.String()
}

// Bit returns 1 if the ith bit from the right is true, and false otherwise.
func (n *NodeID) Bit(i int) uint {
	if got, want := i, n.PathLenBits()-1; got > want {
		panic(fmt.Sprintf("storage: Bit(%v) > (PathLenBits() -1): %v", got, want))
	}
	bIndex := (n.PathLenBits() - i - 1) / 8
	return uint((n.Path[bIndex] >> uint(i%8)) & 0x01)
}

// NewEmptyNodeID creates a new zero-length NodeID with sufficient underlying
// capacity to store a maximum of maxLenBits.
func NewEmptyNodeID(maxLenBits int) NodeID {
	if got, want := maxLenBits%8, 0; got != want {
		panic(fmt.Sprintf("storeage: NewEmptyNodeID() maxLenBits mod 8: %v, want %v", got, want))
	}
	return NodeID{
		Path:          make([]byte, maxLenBits/8),
		PrefixLenBits: 0,
	}
}

// index is the horizontal index into the tree at level depth, so the returned
// NodeID will be zero padded on the right by depth places.
func NewNodeIDForTreeCoords(depth int64, index int64, maxPathBits int) (NodeID, error) {
	bl := bitLen(index)
	if index < 0 || depth < 0 ||
		bl > int(maxPathBits-int(depth)) ||
		maxPathBits%8 != 0 {
		return NodeID{}, fmt.Errorf("depth/index combination out of range: depth=%d index=%d maxPathBits=%v", depth, index, maxPathBits)
	}
	// This node is effectively a prefix of the subtree underneath (for non-leaf
	// depths), so we shift the index accordingly.
	uidx := uint64(index) << uint(depth)
	r := NewEmptyNodeID(maxPathBits)
	for i := len(r.Path) - 1; uidx > 0 && i >= 0; i-- {
		r.Path[i] = byte(uidx & 0xff)
		uidx >>= 8
	}
	// In the storage model nodes closer to the leaves have longer nodeIDs, so
	// we "reverse" depth here:
	r.PrefixLenBits = int(maxPathBits - int(depth))
	return r, nil
}

func nodeKey(d int, i int64) (string, error) {
	n, err := NewNodeIDForTreeCoords(int64(d), i, 64)
	if err != nil {
		return "", err
	}
	return n.String(), nil
}

func TestCompactVsFullTree(t *testing.T) {
	imt := NewInMemoryMerkleTree(rfc6962.DefaultHasher)
	nodes := make(map[string][]byte)

	for i := int64(0); i < 1024; i++ {
		cmt, err := NewCompactMerkleTreeWithState(
			rfc6962.DefaultHasher,
			imt.LeafCount(),
			func(depth int, index int64) ([]byte, error) {
				k, err := nodeKey(depth, index)
				if err != nil {
					t.Errorf("failed to create nodeID: %v", err)
				}
				h := nodes[k]
				return h, nil
			}, imt.CurrentRoot().Hash())

		if err != nil {
			t.Errorf("interation %d: failed to create CMT with state: %v", i, err)
		}
		if a, b := imt.CurrentRoot().Hash(), cmt.CurrentRoot(); !bytes.Equal(a, b) {
			t.Errorf("iteration %d: Got in-memory root of %v, but compact tree has root %v", i, a, b)
		}

		newLeaf := []byte(fmt.Sprintf("Leaf %d", i))

		iSeq, iHash, err := imt.AddLeaf(newLeaf)
		if err != nil {
			t.Errorf("AddLeaf(): %v", err)
		}

		cSeq, cHash, err := cmt.AddLeaf(newLeaf,
			func(depth int, index int64, hash []byte) error {
				k, err := nodeKey(depth, index)
				if err != nil {
					return fmt.Errorf("failed to create nodeID: %v", err)
				}
				nodes[k] = hash
				return nil
			})
		if err != nil {
			t.Fatalf("mt update failed: %v", err)
		}

		// In-Memory tree is 1-based for sequence numbers, since it's based on the original CT C++ impl.
		if got, want := iSeq, i+1; got != want {
			t.Errorf("iteration %d: Got in-memory sequence number of %d, expected %d", i, got, want)
		}
		if int64(iSeq) != cSeq+1 {
			t.Errorf("iteration %d: Got in-memory sequence number of %d but %d (zero based) from compact tree", i, iSeq, cSeq)
		}
		if a, b := iHash.Hash(), cHash; !bytes.Equal(a, b) {
			t.Errorf("iteration %d: Got leaf hash %v from in-memory tree, but %v from compact tree", i, a, b)
		}
		if a, b := imt.CurrentRoot().Hash(), cmt.CurrentRoot(); !bytes.Equal(a, b) {
			t.Errorf("iteration %d: Got in-memory root of %v, but compact tree has root %v", i, a, b)
		}

	}
}

func TestRootHashForVariousTreeSizes(t *testing.T) {
	tests := []struct {
		size     int64
		wantRoot []byte
	}{
		{10, MustDecodeBase64("VjWMPSYNtCuCNlF/RLnQy6HcwSk6CIipfxm+hettA+4=")},
		{15, MustDecodeBase64("j4SulYmocFuxdeyp12xXCIgK6PekBcxzAIj4zbQzNEI=")},
		{16, MustDecodeBase64("c+4Uc6BCMOZf/v3NZK1kqTUJe+bBoFtOhP+P3SayKRE=")},
		{100, MustDecodeBase64("dUh9hYH88p0CMoHkdr1wC2szbhcLAXOejWpINIooKUY=")},
		{255, MustDecodeBase64("SmdsuKUqiod3RX2jyF2M6JnbdE4QuTwwipfAowI4/i0=")},
		{256, MustDecodeBase64("qFI0t/tZ1MdOYgyPpPzHFiZVw86koScXy9q3FU5casA=")},
		{1000, MustDecodeBase64("RXrgb8xHd55Y48FbfotJwCbV82Kx22LZfEbmBGAvwlQ=")},
		{4095, MustDecodeBase64("cWRFdQhPcjn9WyBXE/r1f04ejxIm5lvg40DEpRBVS0w=")},
		{4096, MustDecodeBase64("6uU/phfHg1n/GksYT6TO9aN8EauMCCJRl3dIK0HDs2M=")},
		{10000,MustDecodeBase64("VZcav65F9haHVRk3wre2axFoBXRNeUh/1d9d5FQfxIg=")},
		{65535,MustDecodeBase64("iPuVYJhP6SEE4gUFp8qbafd2rYv9YTCDYqAxCj8HdLM=")},
	}

	b64e := func(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

	for _, test := range tests {
		tree := NewCompactMerkleTree(rfc6962.DefaultHasher)
		for i := int64(0); i < test.size; i++ {
			l := []byte{byte(i & 0xff), byte((i >> 8) & 0xff)}
			tree.AddLeaf(l, func(int, int64, []byte) error {
				return nil
			})
		}
		if got, want := tree.CurrentRoot(), test.wantRoot; !bytes.Equal(got, want) {
			t.Errorf("Test (treesize=%v) got root %v, want %v", test.size, b64e(got), b64e(want))
		}
	}
}
