package txscript

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
)

// PkScript is a wrapper struct around a byte array, allowing it to be used
// as a map index.
type PkScript struct {
	// class is the type of the script encoded within the byte array. This
	// is used to determine the correct length of the script within the byte
	// array.
	class ScriptClass

	// script is the script contained within a byte array. If the script is
	// smaller than the length of the byte array, it will be padded with 0s
	// at the end.
	script [maxLen]byte
}

// Class returns the script type.
func (s PkScript) Class() ScriptClass {
	return s.class
}

// Script returns the script as a byte slice without any padding.
func (s PkScript) Script() []byte {
	var script []byte

	switch s.class {
	case PubKeyHashTy:
		script = make([]byte, pubKeyHashLen)
		copy(script, s.script[:pubKeyHashLen])

	case ScriptHashTy:
		script = make([]byte, scriptHashLen)
		copy(script, s.script[:scriptHashLen])

	default:
		// Unsupported script type.
		return nil
	}

	return script
}

// Address encodes the script into an address for the given chain.
func (s PkScript) Address(chainParams *params.Params) (types.Address, error) {
	_, addrs, _, err := ExtractPkScriptAddrs(s.Script(), chainParams)
	if err != nil {
		return nil, fmt.Errorf("unable to parse address: %v", err)
	}

	return addrs[0], nil
}

// String returns a hex-encoded string representation of the script.
func (s PkScript) String() string {
	str, _ := DisasmString(s.Script())
	return str
}

// ComputePkScript computes the script of an output by looking at the spending
//
// NOTE: Only P2PKH, P2SH redeem scripts are supported.
func ComputePkScript(sigScript []byte) (PkScript, error) {
	switch {
	case len(sigScript) > 0:
		return computeStandardPkScript(sigScript)
	default:
		return PkScript{}, ErrUnsupportedScriptType
	}
}

// computeStandardPkScript computes the script of an output by looking at the
// spending input's signature script.
func computeStandardPkScript(sigScript []byte) (PkScript, error) {
	switch {
	// script types, we should expect to see a push only script.
	case !IsPushOnlyScript(sigScript):
		return PkScript{}, ErrUnsupportedScriptType

	// If a signature script is provided with a length long enough to
	// represent a P2PKH script, then we'll attempt to parse the compressed
	// public key from it.
	case len(sigScript) >= minPubKeyHashSigScriptLen &&
		len(sigScript) <= maxPubKeyHashSigScriptLen:

		// The public key should be found as the last part of the
		// signature script. We'll attempt to parse it to ensure this is
		// a P2PKH redeem script.
		pubKey := sigScript[len(sigScript)-compressedPubKeyLen:]
		if IsStrictCompressedPubKeyEncoding(pubKey) {
			pubKeyHash := hash.Hash160(pubKey)
			script, err := payToPubKeyHashScript(pubKeyHash)
			if err != nil {
				return PkScript{}, err
			}

			pkScript := PkScript{class: PubKeyHashTy}
			copy(pkScript.script[:], script)
			return pkScript, nil
		}

		fallthrough

	// If we failed to parse a compressed public key from the script in the
	// case above, or if the script length is not that of a P2PKH one, we
	// can assume it's a P2SH signature script.
	default:
		// The redeem script will always be the last data push of the
		// signature script, so we'll parse the script into opcodes to
		// obtain it.
		parsedOpcodes, err := parseScript(sigScript)
		if err != nil {
			return PkScript{}, err
		}
		redeemScript := parsedOpcodes[len(parsedOpcodes)-1].data

		scriptHash := hash.Hash160(redeemScript)
		script, err := payToScriptHashScript(scriptHash)
		if err != nil {
			return PkScript{}, err
		}

		pkScript := PkScript{class: ScriptHashTy}
		copy(pkScript.script[:], script)
		return pkScript, nil
	}
}
