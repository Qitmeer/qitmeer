/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package cmds

import (
	"bytes"
	"encoding/hex"
	"github.com/Qitmeer/qng-core/core/types"
	"github.com/Qitmeer/qng-core/core/types/pow"
)

type GetBlockTemplateCmd struct {
	Capabilities []string
	PowType      byte
}

func NewGetBlockTemplateCmd(capabilities []string, powType byte) *GetBlockTemplateCmd {
	return &GetBlockTemplateCmd{
		Capabilities: capabilities,
		PowType:      powType,
	}
}

type SubmitBlockCmd struct {
	HexBlock string
}

func NewSubmitBlockCmd(hexBlock string) *SubmitBlockCmd {
	return &SubmitBlockCmd{
		HexBlock: hexBlock,
	}
}

type GenerateCmd struct {
	NumBlocks uint32
	PowType   pow.PowType
}

func NewGenerateCmd(numBlocks uint32, powType pow.PowType) *GenerateCmd {
	return &GenerateCmd{
		NumBlocks: numBlocks,
		PowType:   powType,
	}
}

type GetRemoteGBTCmd struct {
	PowType pow.PowType
}

func NewGetRemoteGBTCmd(powType pow.PowType) *GetRemoteGBTCmd {
	return &GetRemoteGBTCmd{
		PowType: powType,
	}
}

type SubmitBlockHeaderCmd struct {
	HexBlockHeader string
}

func NewSubmitBlockHeaderCmd(blockHeader *types.BlockHeader) *SubmitBlockHeaderCmd {
	var headerBuf bytes.Buffer
	err := blockHeader.Serialize(&headerBuf)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return &SubmitBlockHeaderCmd{
		HexBlockHeader: hex.EncodeToString(headerBuf.Bytes()),
	}
}

func init() {
	flags := UsageFlag(0)

	MustRegisterCmd("getBlockTemplate", (*GetBlockTemplateCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("submitBlock", (*SubmitBlockCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("getRemoteGBT", (*GetRemoteGBTCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("submitBlockHeader", (*SubmitBlockHeaderCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("generate", (*GenerateCmd)(nil), flags, MinerNameSpace)
}
