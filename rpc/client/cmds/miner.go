/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package cmds

import "github.com/Qitmeer/qitmeer/core/types/pow"

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

func init() {
	flags := UsageFlag(0)

	MustRegisterCmd("getBlockTemplate", (*GetBlockTemplateCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("submitBlock", (*SubmitBlockCmd)(nil), flags, DefaultServiceNameSpace)
	MustRegisterCmd("generate", (*GenerateCmd)(nil), flags, MinerNameSpace)
}
