package qx

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qng-core/engine/txscript"
)

func ScriptDecode(rawScriptStr string) {
	scriptBytes, err := hex.DecodeString(rawScriptStr)
	if err != nil {
		ErrExit(err)
	}
	out, err := txscript.DisasmString(scriptBytes)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%v\n", out)
}

func ScriptEncode(rawOps string) {
	bytes, err := txscript.PkStringToScript(rawOps)
	if err != nil {
		ErrExit(err)
	}
	fmt.Printf("%x\n", bytes)
}
