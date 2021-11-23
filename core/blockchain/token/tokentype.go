package token

import (
	"fmt"
	"github.com/Qitmeer/qng-core/core/serialization"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/engine/txscript"
)

const MaxTokenNameLength = 6

type TokenType struct {
	Id      types.CoinID
	Owners  []byte
	UpLimit uint64
	Enable  bool
	Name    string
	FeeCfg  TokenFeeConfig
}

func (tt *TokenType) Serialize() ([]byte, error) {

	serializeSize := serialization.SerializeSizeVLQ(uint64(tt.Id))
	serializeSize += serialization.SerializeSizeVLQ(uint64(len(tt.Owners)))
	serializeSize += len(tt.Owners)
	serializeSize += serialization.SerializeSizeVLQ(tt.UpLimit)
	serializeSize += 1
	serializeSize += serialization.SerializeSizeVLQ(uint64(len(tt.Name)))
	serializeSize += len(tt.Name)
	serializeSize += serialization.SerializeSizeVLQ(uint64(tt.FeeCfg.Type))
	serializeSize += serialization.SerializeSizeVLQ(uint64(tt.FeeCfg.Value))

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
	offset += serialization.PutVLQ(serialized[offset:], uint64(tt.FeeCfg.Type))
	offset += serialization.PutVLQ(serialized[offset:], uint64(tt.FeeCfg.Value))

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

	//fee type
	feeType, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading fee type")
	}
	offset += bytesRead

	//fee value
	feeValue, bytesRead := serialization.DeserializeVLQ(data[offset:])
	if bytesRead == 0 {
		return offset, fmt.Errorf("unexpected end of data while reading fee value")
	}
	offset += bytesRead

	tt.Id = types.CoinID(Id)
	tt.Owners = Owners
	tt.UpLimit = UpLimit
	tt.Enable = enableB > 0
	tt.Name = string(Name)
	tt.FeeCfg = TokenFeeConfig{Type: types.FeeType(feeType), Value: int64(feeValue)}
	return offset, nil
}

func (tt *TokenType) GetAddress() types.Address {
	script, err := txscript.ParsePkScript(tt.Owners)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	if tnScript, ok := script.(*txscript.TokenScript); ok {
		addr := tnScript.GetAddresses()
		if len(addr) > 0 {
			return addr[0]
		}
	}
	return nil
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
		log.Trace(fmt.Sprintf("Token type update: new %s(%d)", update.Tt.Name, update.Tt.Id))
	case types.TxTypeTokenRenew:
		tt, ok := (*ttm)[update.Tt.Id]
		if !ok {
			return fmt.Errorf("It doesn't exist: Coin id (%d)\n", update.Tt.Id)
		}
		if tt.Enable {
			return fmt.Errorf("Renew is allowed only when disable: Coin id (%d)\n", update.Tt.Id)
		}
		tt.Owners = update.Tt.Owners
		tt.UpLimit = update.Tt.UpLimit
		tt.Name = update.Tt.Name
		(*ttm)[tt.Id] = tt
		log.Trace(fmt.Sprintf("Token type update: renew %s(%d)", update.Tt.Name, update.Tt.Id))
	case types.TxTypeTokenValidate:
		tt, ok := (*ttm)[update.Tt.Id]
		if !ok {
			return fmt.Errorf("It doesn't exist: Coin id (%d)\n", update.Tt.Id)
		}
		if tt.Enable {
			return fmt.Errorf("Validate is allowed only when disable: Coin id (%d)\n", update.Tt.Id)
		}
		tt.Enable = true
		(*ttm)[tt.Id] = tt
		log.Trace(fmt.Sprintf("Token type update: validate %s(%d)", update.Tt.Name, update.Tt.Id))
	case types.TxTypeTokenInvalidate:
		tt, ok := (*ttm)[update.Tt.Id]
		if !ok {
			return fmt.Errorf("It doesn't exist: Coin id (%d)\n", update.Tt.Id)
		}
		if !tt.Enable {
			return fmt.Errorf("Invalidate is allowed only when enable: Coin id (%d)\n", update.Tt.Id)
		}
		tt.Enable = false
		(*ttm)[tt.Id] = tt
		log.Trace(fmt.Sprintf("Token type update: invalidate %s(%d)", update.Tt.Name, update.Tt.Id))
	default:
		return fmt.Errorf("unknown update type %v", update.Typ)
	}

	return nil
}
