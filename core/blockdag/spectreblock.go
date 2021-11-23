package blockdag

import "github.com/Qitmeer/qng-core/common/hash"

type ISpectre interface {
	Vote(x IBlock, y IBlock) int
}

type SpectreBlock struct {
	hash           hash.Hash
	Votes1, Votes2 int // votes in future set, -1 means not voted yet
}

func (sb *SpectreBlock) GetHash() *hash.Hash {
	return &sb.hash
}

type SpectreBlockData struct {
	hash      hash.Hash
	parents   []*hash.Hash
	timestamp int64
}

func (sd *SpectreBlockData) GetHash() *hash.Hash {
	return &sd.hash
}

func (sd *SpectreBlockData) GetParents() []*hash.Hash {
	return sd.parents
}

func (sd *SpectreBlockData) GetTimestamp() int64 {
	return sd.timestamp
}

// Acquire the weight of block
func (sd *SpectreBlockData) GetWeight() uint64 {
	return 1
}

func (sd *SpectreBlockData) GetPriority() int {
	return 1
}
