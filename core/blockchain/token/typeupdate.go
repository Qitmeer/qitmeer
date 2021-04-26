package token

import (
	"github.com/Qitmeer/qitmeer/common/hash"
)

type TypeUpdate struct {
	*TokenUpdate
	Tt TokenType

	cacheHash *hash.Hash
}

func (tu *TypeUpdate) Serialize() ([]byte, error) {
	return nil, nil
}

func (tu *TypeUpdate) Deserialize(data []byte) (int, error) {
	return 0, nil
}

func (tu *TypeUpdate) GetHash() *hash.Hash {
	if tu.cacheHash != nil {
		return tu.cacheHash
	}
	return tu.CacheHash()
}

func (tu *TypeUpdate) CacheHash() *hash.Hash {
	tu.cacheHash = nil
	bs, err := tu.Serialize()
	if err != nil {
		log.Error(err.Error())
		return tu.cacheHash
	}
	h := hash.DoubleHashH(bs)
	tu.cacheHash = &h
	return tu.cacheHash
}
