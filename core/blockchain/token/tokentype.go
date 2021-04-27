package token

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
)

const MaxTokenNameLength = 6

type TokenType struct {
	Id      types.CoinID
	Owners  []byte
	UpLimit uint64
	Enable  bool
	Name    string
}

func (tt *TokenType) Serialize() ([]byte, error) {

	serializeSize := serialization.SerializeSizeVLQ(uint64(tt.Id))
	serializeSize += serialization.SerializeSizeVLQ(uint64(len(tt.Owners)))
	serializeSize += len(tt.Owners)
	serializeSize += serialization.SerializeSizeVLQ(tt.UpLimit)
	serializeSize += 1
	serializeSize += serialization.SerializeSizeVLQ(uint64(len(tt.Name)))
	serializeSize += len(tt.Name)

	serialized := make([]byte, serializeSize)
	offset := 0

	offset += serialization.PutVLQ(serialized[offset:], uint64(tt.Id))
	offset += serialization.PutVLQ(serialized[offset:], uint64(len(tt.Owners)))
	copy(serialized[offset:offset+len(tt.Owners)], tt.Owners)
	offset += len(tt.Owners)
	offset += serialization.PutVLQ(serialized[offset:], tt.UpLimit)

	if tt.Enable {
		offset += serialization.PutVLQ(serialized[offset:], uint64(1))
	} else {
		offset += serialization.PutVLQ(serialized[offset:], uint64(0))
	}
	offset += serialization.PutVLQ(serialized[offset:], uint64(len(tt.Name)))
	copy(serialized[offset:offset+len(tt.Name)], tt.Name)
	offset += len(tt.Name)

	return serialized, nil
}

func (tt *TokenType) Deserialize(data []byte) (int, error) {
	offset := 0
	//tokenId
	Id, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading coin id")
	}
	offset += bytesRead
	//Owners
	ownersLength, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading owners")
	}
	offset += bytesRead

	Owners := data[offset : offset+int(ownersLength)]
	bytesRead = len(Owners)
	offset += bytesRead

	//UpLimit
	UpLimit, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading UpLimit")
	}
	offset += bytesRead

	//Enable
	enableB, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading Enable")
	}
	offset += bytesRead

	//Name
	nameLength, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading owners")
	}
	offset += bytesRead

	Name := data[offset : offset+int(nameLength)]
	bytesRead = len(Name)
	offset += bytesRead

	tt.Id = types.CoinID(Id)
	tt.Owners = Owners
	tt.UpLimit = UpLimit
	tt.Enable = enableB > 0
	tt.Name = string(Name)
	return offset, nil
}

type TokenTypesMap map[types.CoinID]TokenType

func (ttm *TokenTypesMap) Update(update *TypeUpdate) error {
	switch update.Typ {
	case types.TxTypeTokenNew:
		_, ok := (*ttm)[update.Tt.Id]
		if ok {
			return fmt.Errorf("It already exists:%s\n", update.Tt.Name)
		}
		(*ttm)[update.Tt.Id] = update.Tt
	default:
		return fmt.Errorf("unknown update type %v", update.Typ)
	}

	return nil
}
