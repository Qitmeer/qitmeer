package token

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
)

type TypeUpdate struct {
	*TokenUpdate
	Tt TokenType

	cacheHash *hash.Hash
}

func (tu *TypeUpdate) Serialize() ([]byte, error) {
	tuSerialized, err := tu.TokenUpdate.Serialize()
	if err != nil {
		return nil, err
	}

	serializeSize := serialization.SerializeSizeVLQ(uint64(tu.Tt.Id))
	serializeSize += serialization.SerializeSizeVLQ(uint64(len(tu.Tt.Owners)))
	serializeSize += len(tu.Tt.Owners)
	serializeSize += serialization.SerializeSizeVLQ(tu.Tt.UpLimit)
	serializeSize += 1
	serializeSize += serialization.SerializeSizeVLQ(uint64(len(tu.Tt.Name)))
	serializeSize += len(tu.Tt.Name)

	serialized := make([]byte, serializeSize)
	offset := 0

	offset += serialization.PutVLQ(serialized[offset:], uint64(tu.Tt.Id))
	offset += serialization.PutVLQ(serialized[offset:], uint64(len(tu.Tt.Owners)))
	copy(serialized[offset:offset+len(tu.Tt.Owners)], tu.Tt.Owners)
	offset += len(tu.Tt.Owners)
	offset += serialization.PutVLQ(serialized[offset:], tu.Tt.UpLimit)

	if tu.Tt.Enable {
		offset += serialization.PutVLQ(serialized[offset:], uint64(1))
	} else {
		offset += serialization.PutVLQ(serialized[offset:], uint64(0))
	}
	offset += serialization.PutVLQ(serialized[offset:], uint64(len(tu.Tt.Name)))
	copy(serialized[offset:offset+len(tu.Tt.Name)], tu.Tt.Name)
	offset += len(tu.Tt.Name)

	serialized = append(tuSerialized, serialized...)
	return serialized, nil
}

func (tu *TypeUpdate) Deserialize(data []byte) (int, error) {
	bytesRead, err := tu.TokenUpdate.Deserialize(data)
	if err != nil {
		return bytesRead, err
	}
	offset := bytesRead
	//tokenId
	Id, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading coin id at update")
	}
	offset += bytesRead
	//Owners
	ownersLength, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading owners at update")
	}
	Owners := data[offset : offset+int(ownersLength)]
	bytesRead = len(Owners)
	offset += bytesRead

	//UpLimit
	UpLimit, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading UpLimit at update")
	}
	offset += bytesRead

	//Enable
	enableB, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading Enable at update")
	}
	offset += bytesRead

	//Name
	nameLength, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading owners at update")
	}
	Name := data[offset : offset+int(nameLength)]
	bytesRead = len(Name)
	offset += bytesRead

	tu.Tt = TokenType{
		Id:      types.CoinID(Id),
		Owners:  Owners,
		UpLimit: UpLimit,
		Enable:  enableB > 0,
		Name:    string(Name),
	}
	return offset, nil
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

func (tu *TypeUpdate) CheckSanity() error {
	if tu.GetType() == types.TxTypeTokenNew {
		class, _, _, err := txscript.ExtractPkScriptAddrs(tu.Tt.Owners, params.ActiveNetParams.Params)
		if err != nil || class == txscript.NonStandardTy {
			return err
		}
		if tu.Tt.UpLimit == 0 {
			return fmt.Errorf("UpLimit cannot be zero")
		}
		if len(tu.Tt.Name) > MaxTokenNameLength {
			return fmt.Errorf("Token name (%s) exceeds the maximum length (%d).\n", tu.Tt.Name, MaxTokenNameLength)
		}
		if len(tu.Tt.Name) <= 0 {
			return fmt.Errorf("Must have token name.\n")
		}
	} else {
		return fmt.Errorf("This type (%v) is not supported\n", tu.GetType())
	}
	return nil
}
