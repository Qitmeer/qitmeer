// Copyright 2017-2018 The nox developers

package types

type Genesis struct {
	Config    *Config
	Nonce    uint64
}

type genesisJSON struct {
	Config  *Config           `json:"config"`
	Nonce   UInt64           `json:"nonce"`
}

