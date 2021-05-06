// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"encoding/binary"
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/ecc"
	"github.com/Qitmeer/qitmeer/params"
)

const (
	// MaxDataCarrierSize is the maximum number of bytes allowed in pushed
	// data to be considered a nulldata transaction.
	MaxDataCarrierSize = 256
)

// ScriptClass is an enumeration for the list of standard types of script.
type ScriptClass byte

// Classes of script payment known about in the blockchain.
const (
	NonStandardTy     ScriptClass = iota // None of the recognized forms.
	PubKeyTy                             // Pay pubkey.
	PubKeyHashTy                         // Pay pubkey hash.
	ScriptHashTy                         // Pay to script hash.
	MultiSigTy                           // Multi signature.
	NullDataTy                           // Empty data-only (provably prunable).
	StakeSubmissionTy                    // Stake submission.
	StakeGenTy                           // Stake generation
	StakeRevocationTy                    // Stake revocation.
	StakeSubChangeTy                     // Change for stake submission tx.
	PubkeyAltTy                          // Alternative signature pubkey.
	PubkeyHashAltTy                      // Alternative signature pubkey hash.
	CLTVPubKeyHashTy                     // Check Lock Time Verify Pay pubkey hash.
	TokenPubKeyHashTy                    // Token Pay pubkey hash.
)

// Script Interface provide a abstract layer to support new Script parsing from opcode
type Script interface {
	GetClass() ScriptClass
	Match(pops []ParsedOpcode) bool
	SetOpcode(pops []ParsedOpcode) error
	GetAddresses() []types.Address
	RequiredSigs() bool
}

// The registry where the add-on script been registered
var scriptRegistry = map[ScriptClass]Script{
	NonStandardTy:     &NonStandardScript{},
	TokenPubKeyHashTy: &TokenNewScript{},
}

func fromRegisteredScript(pops []ParsedOpcode) Script {
	for _, s := range scriptRegistry {
		if s.Match(pops) {
			s.SetOpcode(pops)
			return s
		}
	}
	return &NonStandardScript{}
}

// Register Script into the registry, which can be parsed from the outside of txscript engine
func RegisterScript(sin Script) error {
	for _, s := range scriptRegistry {
		if s.GetClass() == sin.GetClass() {
			return fmt.Errorf("%v has been registred as script %v ", sin, s)
		}
	}
	scriptRegistry[sin.GetClass()] = sin
	return nil
}

// ParsePkScript returns a parsed Script from the Script registry, which's addresses and
// required signatures are associated with. Note: It's methold only works for the
// registered Script
func ParsePkScript(pkScript []byte) (Script, error) {
	pops, err := parseScript(pkScript)
	if err != nil {
		return nil, err
	}
	return fromRegisteredScript(pops), nil
}

type NonStandardScript struct{}

func (s *NonStandardScript) Name() string {
	return scriptClassToName[NonStandardTy]
}
func (s *NonStandardScript) GetClass() ScriptClass {
	return NonStandardTy
}
func (s *NonStandardScript) Match(pops []ParsedOpcode) bool {
	return false
}
func (s *NonStandardScript) SetOpcode(pops []ParsedOpcode) error {
	return nil
}
func (s *NonStandardScript) GetAddresses() []types.Address {
	return []types.Address{}
}
func (s *NonStandardScript) RequiredSigs() bool {
	return true
}

var _ Script = (*NonStandardScript)(nil)

// scriptClassToName houses the human-readable strings which describe each
// script class.
var scriptClassToName = []string{
	NonStandardTy:     "nonstandard",
	PubKeyTy:          "pubkey",
	PubkeyAltTy:       "pubkeyalt",
	PubKeyHashTy:      "pubkeyhash",
	PubkeyHashAltTy:   "pubkeyhashalt",
	ScriptHashTy:      "scripthash",
	MultiSigTy:        "multisig",
	NullDataTy:        "nulldata",
	StakeSubmissionTy: "stakesubmission",
	StakeGenTy:        "stakegen",
	StakeRevocationTy: "stakerevoke",
	StakeSubChangeTy:  "sstxchange",
	CLTVPubKeyHashTy:  "cltvpubkeyhash",
	TokenPubKeyHashTy: "tokenpubkeyhash",
}

// String implements the Stringer interface by returning the name of
// the enum script class. If the enum is invalid then "Invalid" will be
// returned.
func (t ScriptClass) String() string {
	if int(t) > len(scriptClassToName) || int(t) < 0 {
		return "Invalid"
	}
	return scriptClassToName[t]
}

// isPubkey returns true if the script passed is a pay-to-pubkey transaction,
// false otherwise.
func isPubkey(pops []ParsedOpcode) bool {
	// Valid pubkeys are either 33 or 65 bytes.
	return len(pops) == 2 &&
		(len(pops[0].data) == 33 || len(pops[0].data) == 65) &&
		pops[1].opcode.value == OP_CHECKSIG
}

// isOneByteMaxDataPush returns true if the parsed opcode pushes exactly one
// byte to the stack.
func isOneByteMaxDataPush(po ParsedOpcode) bool {
	return po.opcode.value == OP_1 ||
		po.opcode.value == OP_2 ||
		po.opcode.value == OP_3 ||
		po.opcode.value == OP_4 ||
		po.opcode.value == OP_5 ||
		po.opcode.value == OP_6 ||
		po.opcode.value == OP_7 ||
		po.opcode.value == OP_8 ||
		po.opcode.value == OP_9 ||
		po.opcode.value == OP_10 ||
		po.opcode.value == OP_11 ||
		po.opcode.value == OP_12 ||
		po.opcode.value == OP_13 ||
		po.opcode.value == OP_14 ||
		po.opcode.value == OP_15 ||
		po.opcode.value == OP_16 ||
		po.opcode.value == OP_DATA_1
}

// isPubkey returns true if the script passed is an alternative pay-to-pubkey
// transaction, false otherwise.
func isPubkeyAlt(pops []ParsedOpcode) bool {
	// An alternative pubkey must be less than 512 bytes.
	return len(pops) == 3 &&
		len(pops[0].data) < 512 &&
		isOneByteMaxDataPush(pops[1]) &&
		pops[2].opcode.value == OP_CHECKSIGALT
}

// isPubkeyHash returns true if the script passed is a pay-to-pubkey-hash
// transaction, false otherwise.
func isPubkeyHash(pops []ParsedOpcode) bool {
	return len(pops) == 5 &&
		pops[0].opcode.value == OP_DUP &&
		pops[1].opcode.value == OP_HASH160 &&
		pops[2].opcode.value == OP_DATA_20 &&
		pops[3].opcode.value == OP_EQUALVERIFY &&
		pops[4].opcode.value == OP_CHECKSIG
}

// isPubkeyHashAlt returns true if the script passed is a pay-to-pubkey-hash
// transaction, false otherwise.
func isPubkeyHashAlt(pops []ParsedOpcode) bool {
	return len(pops) == 6 &&
		pops[0].opcode.value == OP_DUP &&
		pops[1].opcode.value == OP_HASH160 &&
		pops[2].opcode.value == OP_DATA_20 &&
		pops[3].opcode.value == OP_EQUALVERIFY &&
		isOneByteMaxDataPush(pops[4]) &&
		pops[5].opcode.value == OP_CHECKSIGALT
}

// isScriptHash returns true if the script passed is a pay-to-script-hash
// transaction, false otherwise.
func isScriptHash(pops []ParsedOpcode) bool {
	return len(pops) == 3 &&
		pops[0].opcode.value == OP_HASH160 &&
		pops[1].opcode.value == OP_DATA_20 &&
		pops[2].opcode.value == OP_EQUAL
}

// isAnyKindOfScriptHash returns true if the script passed is a pay-to-script-hash
// or stake pay-to-script-hash transaction, false otherwise. Used to make the
// engine have the correct behaviour.
func isAnyKindOfScriptHash(pops []ParsedOpcode) bool {
	standardP2SH := len(pops) == 3 &&
		pops[0].opcode.value == OP_HASH160 &&
		pops[1].opcode.value == OP_DATA_20 &&
		pops[2].opcode.value == OP_EQUAL
	if standardP2SH {
		return true
	}

	return len(pops) == 4 &&
		(pops[0].opcode.value >= 186 && pops[0].opcode.value <= 189) &&
		pops[1].opcode.value == OP_HASH160 &&
		pops[2].opcode.value == OP_DATA_20 &&
		pops[3].opcode.value == OP_EQUAL
}

// isMultiSig returns true if the passed script is a multisig transaction, false
// otherwise.
func isMultiSig(pops []ParsedOpcode) bool {
	// The absolute minimum is 1 pubkey:
	// OP_0/OP_1-16 <pubkey> OP_1 OP_CHECKMULTISIG
	l := len(pops)
	if l < 4 {
		return false
	}
	if !isSmallInt(pops[0].opcode) {
		return false
	}
	if !isSmallInt(pops[l-2].opcode) {
		return false
	}
	if pops[l-1].opcode.value != OP_CHECKMULTISIG {
		return false
	}

	// Verify the number of pubkeys specified matches the actual number
	// of pubkeys provided.
	if l-2-1 != asSmallInt(pops[l-2].opcode) {
		return false
	}

	for _, pop := range pops[1 : l-2] {
		// Valid pubkeys are either 33 or 65 bytes.
		if len(pop.data) != 33 && len(pop.data) != 65 {
			return false
		}
	}
	return true
}

// IsMultisigScript takes a script, parses it, then returns whether or
// not it is a multisignature script.
func IsMultisigScript(script []byte) (bool, error) {
	pops, err := parseScript(script)
	if err != nil {
		return false, err
	}
	return isMultiSig(pops), nil
}

// IsMultisigSigScript takes a script, parses it, then returns whether or
// not it is a multisignature script.
func IsMultisigSigScript(script []byte) bool {
	if len(script) == 0 || script == nil {
		return false
	}
	pops, err := parseScript(script)
	if err != nil {
		return false
	}
	subPops, err := parseScript(pops[len(pops)-1].data)
	if err != nil {
		return false
	}

	return isMultiSig(subPops)
}

// isNullData returns true if the passed script is a null data transaction,
// false otherwise.
func isNullData(pops []ParsedOpcode) bool {
	// A nulldata transaction is either a single OP_RETURN or an
	// OP_RETURN SMALLDATA (where SMALLDATA is a data push up to
	// MaxDataCarrierSize bytes).
	l := len(pops)
	if l == 1 && pops[0].opcode.value == OP_RETURN {
		return true
	}

	return l == 2 &&
		pops[0].opcode.value == OP_RETURN &&
		(isSmallInt(pops[1].opcode) || pops[1].opcode.value <=
			OP_PUSHDATA4) &&
		len(pops[1].data) <= MaxDataCarrierSize
}

// isStakeSubmission returns true if the script passed is a stake submission tx,
// false otherwise.
func isStakeSubmission(pops []ParsedOpcode) bool {
	if len(pops) == 6 &&
		pops[0].opcode.value == OP_SSTX &&
		pops[1].opcode.value == OP_DUP &&
		pops[2].opcode.value == OP_HASH160 &&
		pops[3].opcode.value == OP_DATA_20 &&
		pops[4].opcode.value == OP_EQUALVERIFY &&
		pops[5].opcode.value == OP_CHECKSIG {
		return true
	}

	if len(pops) == 4 &&
		pops[0].opcode.value == OP_SSTX &&
		pops[1].opcode.value == OP_HASH160 &&
		pops[2].opcode.value == OP_DATA_20 &&
		pops[3].opcode.value == OP_EQUAL {
		return true
	}

	return false
}

// isStakeGen returns true if the script passed is a stake generation tx,
// false otherwise.
func isStakeGen(pops []ParsedOpcode) bool {
	if len(pops) == 6 &&
		pops[0].opcode.value == OP_SSGEN &&
		pops[1].opcode.value == OP_DUP &&
		pops[2].opcode.value == OP_HASH160 &&
		pops[3].opcode.value == OP_DATA_20 &&
		pops[4].opcode.value == OP_EQUALVERIFY &&
		pops[5].opcode.value == OP_CHECKSIG {
		return true
	}

	if len(pops) == 4 &&
		pops[0].opcode.value == OP_SSGEN &&
		pops[1].opcode.value == OP_HASH160 &&
		pops[2].opcode.value == OP_DATA_20 &&
		pops[3].opcode.value == OP_EQUAL {
		return true
	}

	return false
}

// isStakeRevocation returns true if the script passed is a stake submission
// revocation tx, false otherwise.
func isStakeRevocation(pops []ParsedOpcode) bool {
	if len(pops) == 6 &&
		pops[0].opcode.value == OP_SSRTX &&
		pops[1].opcode.value == OP_DUP &&
		pops[2].opcode.value == OP_HASH160 &&
		pops[3].opcode.value == OP_DATA_20 &&
		pops[4].opcode.value == OP_EQUALVERIFY &&
		pops[5].opcode.value == OP_CHECKSIG {
		return true
	}

	if len(pops) == 4 &&
		pops[0].opcode.value == OP_SSRTX &&
		pops[1].opcode.value == OP_HASH160 &&
		pops[2].opcode.value == OP_DATA_20 &&
		pops[3].opcode.value == OP_EQUAL {
		return true
	}

	return false
}

// isSStxChange returns true if the script passed is a stake submission
// change tx, false otherwise.
func isSStxChange(pops []ParsedOpcode) bool {
	if len(pops) == 6 &&
		pops[0].opcode.value == OP_SSTXCHANGE &&
		pops[1].opcode.value == OP_DUP &&
		pops[2].opcode.value == OP_HASH160 &&
		pops[3].opcode.value == OP_DATA_20 &&
		pops[4].opcode.value == OP_EQUALVERIFY &&
		pops[5].opcode.value == OP_CHECKSIG {
		return true
	}

	if len(pops) == 4 &&
		pops[0].opcode.value == OP_SSTXCHANGE &&
		pops[1].opcode.value == OP_HASH160 &&
		pops[2].opcode.value == OP_DATA_20 &&
		pops[3].opcode.value == OP_EQUAL {
		return true
	}

	return false
}

// isCLTVPubkeyHash returns true if the script passed is a pay-to-cltv-pubkey-hash
// transaction, false otherwise.
func isCLTVPubkeyHash(pops []ParsedOpcode) bool {
	return len(pops) == 8 &&
		pops[1].opcode.value == OP_CHECKLOCKTIMEVERIFY &&
		pops[2].opcode.value == OP_DROP &&
		pops[3].opcode.value == OP_DUP &&
		pops[4].opcode.value == OP_HASH160 &&
		pops[5].opcode.value == OP_DATA_20 &&
		pops[6].opcode.value == OP_EQUALVERIFY &&
		pops[7].opcode.value == OP_CHECKSIG
}

// isTokenPubkeyHash returns true if the script passed is a pay-to-token-pubkey-hash
// transaction, false otherwise.
func isTokenPubkeyHash(pops []ParsedOpcode) bool {
	return len(pops) == 11 &&
		pops[3].opcode.value == OP_TOKEN &&
		pops[4].opcode.value == OP_DROP &&
		pops[5].opcode.value == OP_2DROP &&
		pops[6].opcode.value == OP_DUP &&
		pops[7].opcode.value == OP_HASH160 &&
		pops[8].opcode.value == OP_DATA_20 &&
		pops[9].opcode.value == OP_EQUALVERIFY &&
		pops[10].opcode.value == OP_CHECKSIG
}

// scriptType returns the type of the script being inspected from the known
// standard types.
func typeOfScript(pops []ParsedOpcode) ScriptClass {
	if isPubkey(pops) {
		return PubKeyTy
	} else if isPubkeyAlt(pops) {
		return PubkeyAltTy
	} else if isPubkeyHash(pops) {
		return PubKeyHashTy
	} else if isPubkeyHashAlt(pops) {
		return PubkeyHashAltTy
	} else if isScriptHash(pops) {
		return ScriptHashTy
	} else if isMultiSig(pops) {
		return MultiSigTy
	} else if isNullData(pops) {
		return NullDataTy
	} else if isStakeSubmission(pops) {
		return StakeSubmissionTy
	} else if isStakeGen(pops) {
		return StakeGenTy
	} else if isStakeRevocation(pops) {
		return StakeRevocationTy
	} else if isSStxChange(pops) {
		return StakeSubChangeTy
	} else if isCLTVPubkeyHash(pops) {
		return CLTVPubKeyHashTy
	} else if isTokenPubkeyHash(pops) {
		return TokenPubKeyHashTy
	}

	return NonStandardTy
}

// GetScriptClass returns the class of the script passed.
//
// NonStandardTy will be returned when the script does not parse.
func GetScriptClass(version uint16, script []byte) ScriptClass {
	// NullDataTy outputs are allowed to have non-default script
	// versions. However, other types are not.
	if version != DefaultScriptVersion {
		return NonStandardTy
	}

	pops, err := parseScript(script)
	if err != nil {
		log.Error("parseScript error", "script", fmt.Sprintf("%x", script))
		return NonStandardTy
	}

	return typeOfScript(pops)
}

// expectedInputs returns the number of arguments required by a script.
// If the script is of unknown type such that the number can not be determined
// then -1 is returned. We are an internal function and thus assume that class
// is the real class of pops (and we can thus assume things that were determined
// while finding out the type).
func expectedInputs(pops []ParsedOpcode, class ScriptClass,
	subclass ScriptClass) int {
	switch class {
	case PubKeyTy:
		return 1

	case PubKeyHashTy:
		return 2

	case StakeSubmissionTy:
		if subclass == PubKeyHashTy {
			return 2
		}
		return 1 // P2SH

	case StakeGenTy:
		if subclass == PubKeyHashTy {
			return 2
		}
		return 1 // P2SH

	case StakeRevocationTy:
		if subclass == PubKeyHashTy {
			return 2
		}
		return 1 // P2SH

	case StakeSubChangeTy:
		if subclass == PubKeyHashTy {
			return 2
		}
		return 1 // P2SH

	case ScriptHashTy:
		// Not including script, handled below.
		return 1

	case MultiSigTy:
		// Standard multisig has a push a small number for the number
		// of sigs and number of keys.  Check the first push instruction
		// to see how many arguments are expected. typeOfScript already
		// checked this so we know it'll be a small int.  Also, due to
		// the original bitcoind bug where OP_CHECKMULTISIG pops an
		// additional item from the stack, add an extra expected input
		// for the extra push that is required to compensate.
		return asSmallInt(pops[0].opcode)

	case NullDataTy:
		fallthrough
	default:
		return -1
	}
}

// ScriptInfo houses information about a script pair that is determined by
// CalcScriptInfo.
type ScriptInfo struct {
	// PkScriptClass is the class of the public key script and is equivalent
	// to calling GetScriptClass on it.
	PkScriptClass ScriptClass

	// NumInputs is the number of inputs provided by the public key script.
	NumInputs int

	// ExpectedInputs is the number of outputs required by the signature
	// script and any pay-to-script-hash scripts. The number will be -1 if
	// unknown.
	ExpectedInputs int

	// SigOps is the number of signature operations in the script pair.
	SigOps int
}

// IsStakeOutput returns true is a script output is a stake type.
func IsStakeOutput(pkScript []byte) bool {
	pkPops, err := parseScript(pkScript)
	if err != nil {
		return false
	}

	class := typeOfScript(pkPops)
	return class == StakeSubmissionTy ||
		class == StakeGenTy ||
		class == StakeRevocationTy ||
		class == StakeSubChangeTy
}

// GetStakeOutSubclass extracts the subclass (P2PKH or P2SH)
// from a stake output.
func GetStakeOutSubclass(pkScript []byte) (ScriptClass, error) {
	pkPops, err := parseScript(pkScript)
	if err != nil {
		return 0, err
	}

	class := typeOfScript(pkPops)
	isStake := class == StakeSubmissionTy ||
		class == StakeGenTy ||
		class == StakeRevocationTy ||
		class == StakeSubChangeTy

	subClass := ScriptClass(0)
	if isStake {
		var stakeSubscript []ParsedOpcode
		for _, pop := range pkPops {
			if pop.opcode.value >= 186 && pop.opcode.value <= 189 {
				continue
			}
			stakeSubscript = append(stakeSubscript, pop)
		}

		subClass = typeOfScript(stakeSubscript)
	} else {
		return 0, fmt.Errorf("not a stake output")
	}

	return subClass, nil
}

// getStakeOutSubscript extracts the subscript (P2PKH or P2SH)
// from a stake output.
func getStakeOutSubscript(pkScript []byte) []byte {
	return pkScript[1:]
}

// ContainsStakeOpCodes returns whether or not a pkScript contains stake tagging
// OP codes.
func ContainsStakeOpCodes(pkScript []byte) (bool, error) {
	shPops, err := parseScript(pkScript)
	if err != nil {
		return false, err
	}

	for _, pop := range shPops {
		if pop.opcode.value >= 186 && pop.opcode.value <= 189 {
			return true, nil
		}
	}

	return false, nil
}

// CalcScriptInfo returns a structure providing data about the provided script
// pair.  It will error if the pair is in someway invalid such that they can not
// be analysed, i.e. if they do not parse or the pkScript is not a push-only
// script
func CalcScriptInfo(sigScript, pkScript []byte, bip16 bool) (*ScriptInfo, error) {
	sigPops, err := parseScript(sigScript)
	if err != nil {
		return nil, err
	}

	pkPops, err := parseScript(pkScript)
	if err != nil {
		return nil, err
	}

	// Push only sigScript makes little sense.
	si := new(ScriptInfo)
	si.PkScriptClass = typeOfScript(pkPops)

	// Can't have a pkScript that doesn't just push data.
	if !isPushOnly(sigPops) {
		return nil, ErrStackNonPushOnly
	}

	subClass := ScriptClass(0)
	if si.PkScriptClass == StakeSubmissionTy ||
		si.PkScriptClass == StakeGenTy ||
		si.PkScriptClass == StakeRevocationTy ||
		si.PkScriptClass == StakeSubChangeTy {
		subClass, err = GetStakeOutSubclass(pkScript)
		if err != nil {
			return nil, err
		}
	}

	si.ExpectedInputs = expectedInputs(pkPops, si.PkScriptClass, subClass)

	// All entries pushed to stack (or are OP_RESERVED and exec will fail).
	si.NumInputs = len(sigPops)

	// Count sigops taking into account pay-to-script-hash.
	if (si.PkScriptClass == ScriptHashTy || subClass == ScriptHashTy) && bip16 {
		// The pay-to-hash-script is the final data push of the
		// signature script.
		script := sigPops[len(sigPops)-1].data
		shPops, err := parseScript(script)
		if err != nil {
			return nil, err
		}

		shInputs := expectedInputs(shPops, typeOfScript(shPops), 0)
		if shInputs == -1 {
			si.ExpectedInputs = -1
		} else {
			si.ExpectedInputs += shInputs
		}
		si.SigOps = getSigOpCount(shPops, true)
	} else {
		si.SigOps = getSigOpCount(pkPops, true)
	}

	return si, nil
}

// CalcMultiSigStats returns the number of public keys and signatures from
// a multi-signature transaction script.  The passed script MUST already be
// known to be a multi-signature script.
func CalcMultiSigStats(script []byte) (int, int, error) {
	pops, err := parseScript(script)
	if err != nil {
		return 0, 0, err
	}

	// A multi-signature script is of the pattern:
	//  NUM_SIGS PUBKEY PUBKEY PUBKEY... NUM_PUBKEYS OP_CHECKMULTISIG
	// Therefore the number of signatures is the oldest item on the stack
	// and the number of pubkeys is the 2nd to last.  Also, the absolute
	// minimum for a multi-signature script is 1 pubkey, so at least 4
	// items must be on the stack per:
	//  OP_1 PUBKEY OP_1 OP_CHECKMULTISIG
	if len(pops) < 4 {
		return 0, 0, ErrStackUnderflow
	}

	numSigs := asSmallInt(pops[0].opcode)
	numPubKeys := asSmallInt(pops[len(pops)-2].opcode)
	return numPubKeys, numSigs, nil
}

// MultisigRedeemScriptFromScriptSig attempts to extract a multi-
// signature redeem script from a P2SH-redeeming input. It returns
// nil if the signature script is not a multisignature script.
func MultisigRedeemScriptFromScriptSig(script []byte) ([]byte, error) {
	pops, err := parseScript(script)
	if err != nil {
		return nil, err
	}

	// The redeemScript is always the last item on the stack of
	// the script sig.
	return pops[len(pops)-1].data, nil
}

// payToPubKeyHashScript creates a new script to pay a transaction
// output to a 20-byte pubkey hash. It is expected that the input is a valid
// hash.
func payToPubKeyHashScript(pubKeyHash []byte) ([]byte, error) {
	return NewScriptBuilder().AddOp(OP_DUP).AddOp(OP_HASH160).
		AddData(pubKeyHash).AddOp(OP_EQUALVERIFY).AddOp(OP_CHECKSIG).
		Script()
}

// payToPubKeyHashEdwardsScript creates a new script to pay a transaction
// output to a 20-byte pubkey hash of an Edwards public key. It is expected
// that the input is a valid hash.
func payToPubKeyHashEdwardsScript(pubKeyHash []byte) ([]byte, error) {
	edwardsData := []byte{byte(edwards)}
	return NewScriptBuilder().AddOp(OP_DUP).AddOp(OP_HASH160).
		AddData(pubKeyHash).AddOp(OP_EQUALVERIFY).AddData(edwardsData).
		AddOp(OP_CHECKSIGALT).Script()
}

// payToPubKeyHashSchnorrScript creates a new script to pay a transaction
// output to a 20-byte pubkey hash of a secp256k1 public key, but expecting
// a schnorr signature instead of a classic secp256k1 signature. It is
// expected that the input is a valid hash.
func payToPubKeyHashSchnorrScript(pubKeyHash []byte) ([]byte, error) {
	schnorrData := []byte{byte(secSchnorr)}
	return NewScriptBuilder().AddOp(OP_DUP).AddOp(OP_HASH160).
		AddData(pubKeyHash).AddOp(OP_EQUALVERIFY).AddData(schnorrData).
		AddOp(OP_CHECKSIGALT).Script()
}

// payToScriptHashScript creates a new script to pay a transaction output to a
// script hash. It is expected that the input is a valid hash.
func payToScriptHashScript(scriptHash []byte) ([]byte, error) {
	return NewScriptBuilder().AddOp(OP_HASH160).AddData(scriptHash).
		AddOp(OP_EQUAL).Script()
}

// GetScriptHashFromP2SHScript extracts the script hash from a valid
// P2SH pkScript.
func GetScriptHashFromP2SHScript(pkScript []byte) ([]byte, error) {
	pops, err := parseScript(pkScript)
	if err != nil {
		return nil, err
	}

	var sh []byte
	reachedHash160DataPush := false
	for _, p := range pops {
		if p.opcode.value == OP_HASH160 {
			reachedHash160DataPush = true
			continue
		}
		if reachedHash160DataPush {
			sh = p.data
			break
		}
	}

	return sh, nil
}

// PayToScriptHashScript is the exported version of payToScriptHashScript.
func PayToScriptHashScript(scriptHash []byte) ([]byte, error) {
	return payToScriptHashScript(scriptHash)
}

// payToPubkeyScript creates a new script to pay a transaction output to a
// public key. It is expected that the input is a valid pubkey.
func payToPubKeyScript(serializedPubKey []byte) ([]byte, error) {
	return NewScriptBuilder().AddData(serializedPubKey).
		AddOp(OP_CHECKSIG).Script()
}

// payToEdwardsPubKeyScript creates a new script to pay a transaction output
// to an Ed25519 public key. It is expected that the input is a valid pubkey.
func payToEdwardsPubKeyScript(serializedPubKey []byte) ([]byte, error) {
	edwardsData := []byte{byte(edwards)}
	return NewScriptBuilder().AddData(serializedPubKey).AddData(edwardsData).
		AddOp(OP_CHECKSIGALT).Script()
}

// payToSchnorrPubKeyScript creates a new script to pay a transaction output
// to a secp256k1 public key, but to be signed by Schnorr type signature. It
// is expected that the input is a valid pubkey.
func payToSchnorrPubKeyScript(serializedPubKey []byte) ([]byte, error) {
	schnorrData := []byte{byte(secSchnorr)}
	return NewScriptBuilder().AddData(serializedPubKey).AddData(schnorrData).
		AddOp(OP_CHECKSIGALT).Script()
}

// PayToSStx creates a new script to pay a transaction output to a script hash or
// public key hash, but tags the output with OP_SSTX. For use in constructing
// valid SStxs.
func PayToSStx(addr types.Address) ([]byte, error) {
	if addr == nil {
		return nil, ErrUnsupportedAddress
	}

	// Only pay to pubkey hash and pay to script hash are
	// supported.
	scriptType := PubKeyHashTy
	switch addr := addr.(type) {
	case *address.PubKeyHashAddress:
		if addr.EcType() != ecc.ECDSA_Secp256k1 {
			return nil, ErrUnsupportedAddress
		}
	case *address.ScriptHashAddress:
		scriptType = ScriptHashTy
	default:
		return nil, ErrUnsupportedAddress
	}

	hash := addr.Script()

	if scriptType == PubKeyHashTy {
		return NewScriptBuilder().AddOp(OP_SSTX).AddOp(OP_DUP).
			AddOp(OP_HASH160).AddData(hash).AddOp(OP_EQUALVERIFY).
			AddOp(OP_CHECKSIG).Script()
	}
	return NewScriptBuilder().AddOp(OP_SSTX).AddOp(OP_HASH160).
		AddData(hash).AddOp(OP_EQUAL).Script()
}

// PayToSStxChange creates a new script to pay a transaction output to a
// public key hash, but tags the output with OP_SSTXCHANGE. For use in constructing
// valid SStxs.
func PayToSStxChange(addr types.Address) ([]byte, error) {
	if addr == nil {
		return nil, ErrUnsupportedAddress
	}

	// Only pay to pubkey hash and pay to script hash are
	// supported.
	scriptType := PubKeyHashTy
	switch addr := addr.(type) {
	case *address.PubKeyHashAddress:
		if addr.EcType() != ecc.ECDSA_Secp256k1 {
			return nil, ErrUnsupportedAddress
		}
	case *address.ScriptHashAddress:
		scriptType = ScriptHashTy
	default:
		return nil, ErrUnsupportedAddress
	}

	h := addr.Script()

	if scriptType == PubKeyHashTy {
		return NewScriptBuilder().AddOp(OP_SSTXCHANGE).AddOp(OP_DUP).
			AddOp(OP_HASH160).AddData(h).AddOp(OP_EQUALVERIFY).
			AddOp(OP_CHECKSIG).Script()
	}
	return NewScriptBuilder().AddOp(OP_SSTXCHANGE).AddOp(OP_HASH160).
		AddData(h).AddOp(OP_EQUAL).Script()
}

// PayToSSGen creates a new script to pay a transaction output to a public key
// hash or script hash, but tags the output with OP_SSGEN. For use in constructing
// valid SSGen txs.
func PayToSSGen(addr types.Address) ([]byte, error) {
	if addr == nil {
		return nil, ErrUnsupportedAddress
	}

	// Only pay to pubkey hash and pay to script hash are
	// supported.
	scriptType := PubKeyHashTy
	switch addr := addr.(type) {
	case *address.PubKeyHashAddress:
		if addr.EcType() != ecc.ECDSA_Secp256k1 {
			return nil, ErrUnsupportedAddress
		}
	case *address.ScriptHashAddress:
		scriptType = ScriptHashTy
	default:
		return nil, ErrUnsupportedAddress
	}

	h := addr.Script()

	if scriptType == PubKeyHashTy {
		return NewScriptBuilder().AddOp(OP_SSGEN).AddOp(OP_DUP).
			AddOp(OP_HASH160).AddData(h).AddOp(OP_EQUALVERIFY).
			AddOp(OP_CHECKSIG).Script()
	}
	return NewScriptBuilder().AddOp(OP_SSGEN).AddOp(OP_HASH160).
		AddData(h).AddOp(OP_EQUAL).Script()
}

// PayToSSGenPKHDirect creates a new script to pay a transaction output to a
// public key hash, but tags the output with OP_SSGEN. For use in constructing
// valid SSGen txs. Unlike PayToSSGen, this function directly uses the HASH160
// pubkeyhash (instead of an address).
func PayToSSGenPKHDirect(pkh []byte) ([]byte, error) {
	if pkh == nil {
		return nil, ErrUnsupportedAddress
	}

	return NewScriptBuilder().AddOp(OP_SSGEN).AddOp(OP_DUP).
		AddOp(OP_HASH160).AddData(pkh).AddOp(OP_EQUALVERIFY).
		AddOp(OP_CHECKSIG).Script()
}

// PayToSSGenSHDirect creates a new script to pay a transaction output to a
// script hash, but tags the output with OP_SSGEN. For use in constructing
// valid SSGen txs. Unlike PayToSSGen, this function directly uses the HASH160
// script hash (instead of an address).
func PayToSSGenSHDirect(sh []byte) ([]byte, error) {
	if sh == nil {
		return nil, ErrUnsupportedAddress
	}

	return NewScriptBuilder().AddOp(OP_SSGEN).AddOp(OP_HASH160).
		AddData(sh).AddOp(OP_EQUAL).Script()
}

// PayToSSRtx creates a new script to pay a transaction output to a
// public key hash, but tags the output with OP_SSRTX. For use in constructing
// valid SSRtx.
func PayToSSRtx(addr types.Address) ([]byte, error) {
	if addr == nil {
		return nil, ErrUnsupportedAddress
	}

	// Only pay to pubkey hash and pay to script hash are
	// supported.
	scriptType := PubKeyHashTy
	switch addr := addr.(type) {
	case *address.PubKeyHashAddress:
		if addr.EcType() != ecc.ECDSA_Secp256k1 {
			return nil, ErrUnsupportedAddress
		}
	case *address.ScriptHashAddress:
		scriptType = ScriptHashTy
	default:
		return nil, ErrUnsupportedAddress
	}

	h := addr.Script()

	if scriptType == PubKeyHashTy {
		return NewScriptBuilder().AddOp(OP_SSRTX).AddOp(OP_DUP).
			AddOp(OP_HASH160).AddData(h).AddOp(OP_EQUALVERIFY).
			AddOp(OP_CHECKSIG).Script()
	}
	return NewScriptBuilder().AddOp(OP_SSRTX).AddOp(OP_HASH160).
		AddData(h).AddOp(OP_EQUAL).Script()
}

// PayToSSRtxPKHDirect creates a new script to pay a transaction output to a
// public key hash, but tags the output with OP_SSRTX. For use in constructing
// valid SSRtx. Unlike PayToSSRtx, this function directly uses the HASH160
// pubkeyhash (instead of an address).
func PayToSSRtxPKHDirect(pkh []byte) ([]byte, error) {
	if pkh == nil {
		return nil, ErrUnsupportedAddress
	}

	return NewScriptBuilder().AddOp(OP_SSRTX).AddOp(OP_DUP).
		AddOp(OP_HASH160).AddData(pkh).AddOp(OP_EQUALVERIFY).
		AddOp(OP_CHECKSIG).Script()
}

// PayToSSRtxSHDirect creates a new script to pay a transaction output to a
// script hash, but tags the output with OP_SSRTX. For use in constructing
// valid SSRtx. Unlike PayToSSRtx, this function directly uses the HASH160
// script hash (instead of an address).
func PayToSSRtxSHDirect(sh []byte) ([]byte, error) {
	if sh == nil {
		return nil, ErrUnsupportedAddress
	}

	return NewScriptBuilder().AddOp(OP_SSRTX).AddOp(OP_HASH160).
		AddData(sh).AddOp(OP_EQUAL).Script()
}

// PayToCLTVPubKeyHashScript creates a new script to pay a transaction
// output to a 20-byte pubkey hash and lockTime. It is expected that the input is a valid
// hash.
func PayToCLTVPubKeyHashScript(pubKeyHash []byte, lockTime int64) ([]byte, error) {
	return NewScriptBuilder().AddInt64(lockTime).AddOp(OP_CHECKLOCKTIMEVERIFY).AddOp(OP_DROP).AddOp(OP_DUP).AddOp(OP_HASH160).
		AddData(pubKeyHash).AddOp(OP_EQUALVERIFY).AddOp(OP_CHECKSIG).Script()
}

func PayToTokenPubKeyHashScript(pubKeyHash []byte, coinId types.CoinID, upLimit uint64, name string) ([]byte, error) {
	return NewScriptBuilder().AddInt64(int64(coinId)).AddInt64(int64(upLimit)).AddData([]byte(name)).AddOp(OP_TOKEN).AddOp(OP_DROP).AddOp(OP_2DROP).AddOp(OP_DUP).AddOp(OP_HASH160).
		AddData(pubKeyHash).AddOp(OP_EQUALVERIFY).AddOp(OP_CHECKSIG).Script()
}

// GenerateSStxAddrPush generates an OP_RETURN push for SSGen payment addresses in
// an SStx.
func GenerateSStxAddrPush(addr types.Address, amount uint64,
	limits uint16) ([]byte, error) {
	if addr == nil {
		return nil, ErrUnsupportedAddress
	}

	// Only pay to pubkey hash and pay to script hash are
	// supported.
	scriptType := PubKeyHashTy
	switch addr := addr.(type) {
	case *address.PubKeyHashAddress:
		if addr.EcType() != ecc.ECDSA_Secp256k1 {
			return nil, ErrUnsupportedAddress
		}
	case *address.ScriptHashAddress:
		scriptType = ScriptHashTy
	default:
		return nil, ErrUnsupportedAddress
	}

	// Prefix
	dataPushes := []byte{
		0x6a, // OP_RETURN
		0x1e, // OP_DATA_30
	}

	hash := addr.Script()

	amountBuffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountBuffer, uint64(amount))

	// Set the bit flag indicating pay to script hash.
	if scriptType == ScriptHashTy {
		amountBuffer[7] |= 1 << 7
	}

	limitsBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(limitsBuffer, limits)

	// Concatenate the prefix, pubkeyhash, and amount.
	addrOut := append(dataPushes, hash...)
	addrOut = append(addrOut, amountBuffer...)
	addrOut = append(addrOut, limitsBuffer...)

	return addrOut, nil
}

// GenerateSSGenBlockRef generates an OP_RETURN push for the block header hash and
// height which the block votes on.
func GenerateSSGenBlockRef(blockHash hash.Hash, height uint32) ([]byte,
	error) {
	// Prefix
	dataPushes := []byte{
		0x6a, // OP_RETURN
		0x24, // OP_DATA_36
	}

	// Serialize the block hash and height
	blockHeightBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(blockHeightBytes, height)

	blockData := append(blockHash[:], blockHeightBytes...)

	// Concatenate the prefix and block data
	blockDataOut := append(dataPushes, blockData...)

	return blockDataOut, nil
}

// GenerateSSGenVotes generates an OP_RETURN push for the vote bits in an SSGen tx.
func GenerateSSGenVotes(votebits uint16) ([]byte, error) {
	// Prefix
	dataPushes := []byte{
		0x6a, // OP_RETURN
		0x02, // OP_DATA_2
	}

	// Serialize the votebits
	voteBitsBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(voteBitsBytes, votebits)

	// Concatenate the prefix and vote bits
	voteBitsOut := append(dataPushes, voteBitsBytes...)

	return voteBitsOut, nil
}

// GenerateProvablyPruneableOut creates an OP_RETURN push of arbitrary data.
func GenerateProvablyPruneableOut(data []byte) ([]byte, error) {
	if len(data) > MaxDataCarrierSize {
		return nil, ErrStackLongScript
	}

	return NewScriptBuilder().AddOp(OP_RETURN).AddData(data).Script()
}

// PayToAddrScript creates a new script to pay a transaction output to a the
// specified address.
func PayToAddrScript(addr types.Address) ([]byte, error) {
	switch addr := addr.(type) {
	case *address.PubKeyHashAddress:
		if addr == nil {
			return nil, ErrUnsupportedAddress
		}
		switch addr.EcType() {
		case ecc.ECDSA_Secp256k1:
			return payToPubKeyHashScript(addr.Script())
		case ecc.EdDSA_Ed25519:
			return payToPubKeyHashEdwardsScript(addr.Script())
		case ecc.ECDSA_SecpSchnorr:
			return payToPubKeyHashSchnorrScript(addr.Script())
		}
	case *address.ScriptHashAddress:
		if addr == nil {
			return nil, ErrUnsupportedAddress
		}
		return payToScriptHashScript(addr.Script())

	case *address.SecpPubKeyAddress:
		if addr == nil {
			return nil, ErrUnsupportedAddress
		}
		return payToPubKeyScript(addr.Script())

	case *address.EdwardsPubKeyAddress:
		if addr == nil {
			return nil, ErrUnsupportedAddress
		}
		return payToEdwardsPubKeyScript(addr.Script())

	case *address.SecSchnorrPubKeyAddress:
		if addr == nil {
			return nil, ErrUnsupportedAddress
		}
		return payToSchnorrPubKeyScript(addr.Script())
	}

	return nil, ErrUnsupportedAddress
}

// MultiSigScript returns a valid script for a multisignature redemption where
// nrequired of the keys in pubkeys are required to have signed the transaction
// for success.  An ErrBadNumRequired will be returned if nrequired is larger
// than the number of keys provided.
func MultiSigScript(pubkeys []*address.SecpPubKeyAddress, nrequired int) ([]byte,
	error) {
	if len(pubkeys) < nrequired {
		return nil, ErrBadNumRequired
	}

	builder := NewScriptBuilder().AddInt64(int64(nrequired))
	for _, key := range pubkeys {
		builder.AddData(key.Script())
	}
	builder.AddInt64(int64(len(pubkeys)))
	builder.AddOp(OP_CHECKMULTISIG)

	return builder.Script()
}

// PushedData returns an array of byte slices containing any pushed data found
// in the passed script.  This includes OP_0, but not OP_1 - OP_16.
func PushedData(script []byte) ([][]byte, error) {
	pops, err := parseScript(script)
	if err != nil {
		return nil, err
	}

	var data [][]byte
	for _, pop := range pops {
		if pop.data != nil {
			data = append(data, pop.data)
		} else if pop.opcode.value == OP_0 {
			data = append(data, nil)
		}
	}
	return data, nil
}

// GetMultisigMandN returns the number of public keys and the number of
// signatures required to redeem the multisignature script.
func GetMultisigMandN(script []byte) (uint8, uint8, error) {
	// No valid addresses or required signatures if the script doesn't
	// parse.
	pops, err := parseScript(script)
	if err != nil {
		return 0, 0, err
	}

	requiredSigs := uint8(asSmallInt(pops[0].opcode))
	numPubKeys := uint8(asSmallInt(pops[len(pops)-2].opcode))

	return requiredSigs, numPubKeys, nil
}

// ExtractPkScriptAddrs returns the type of script, addresses and required
// signatures associated with the passed PkScript.  Note that it only works for
// 'standard' transaction script types.  Any data such as public keys which are
// invalid are omitted from the results.
func ExtractPkScriptAddrs(pkScript []byte,
	chainParams *params.Params) (ScriptClass, []types.Address, int, error) {

	var addrs []types.Address
	var requiredSigs int

	// No valid addresses or required signatures if the script doesn't
	// parse.
	pops, err := parseScript(pkScript)
	if err != nil {
		return NonStandardTy, nil, 0, err
	}

	scriptClass := typeOfScript(pops)

	switch scriptClass {
	case PubKeyHashTy:
		// A pay-to-pubkey-hash script is of the form:
		//  OP_DUP OP_HASH160 <hash> OP_EQUALVERIFY OP_CHECKSIG
		// Therefore the pubkey hash is the 3rd item on the stack.
		// Skip the pubkey hash if it's invalid for some reason.
		requiredSigs = 1
		addr, err := address.NewPubKeyHashAddress(pops[2].data,
			chainParams, ecc.ECDSA_Secp256k1)
		if err == nil {
			addrs = append(addrs, addr)
		}

	case PubkeyHashAltTy:
		// A pay-to-pubkey-hash script is of the form:
		// OP_DUP OP_HASH160 <hash> OP_EQUALVERIFY <type> OP_CHECKSIGALT
		// Therefore the pubkey hash is the 3rd item on the stack.
		// Skip the pubkey hash if it's invalid for some reason.
		requiredSigs = 1
		suite, _ := ExtractPkScriptAltSigType(pkScript)
		addr, err := address.NewPubKeyHashAddress(pops[2].data,
			chainParams, suite)
		if err == nil {
			addrs = append(addrs, addr)
		}

	case PubKeyTy:
		// A pay-to-pubkey script is of the form:
		//  <pubkey> OP_CHECKSIG
		// Therefore the pubkey is the first item on the stack.
		// Skip the pubkey if it's invalid for some reason.
		requiredSigs = 1
		pk, err := ecc.Secp256k1.ParsePubKey(pops[0].data)
		if err == nil {
			addr, err := address.NewSecpPubKeyCompressedAddress(pk, chainParams)
			if err == nil {
				addrs = append(addrs, addr)
			}
		}

	case PubkeyAltTy:
		// A pay-to-pubkey alt script is of the form:
		//  <pubkey> <type> OP_CHECKSIGALT
		// Therefore the pubkey is the first item on the stack.
		// Skip the pubkey if it's invalid for some reason.
		requiredSigs = 1
		suite, _ := ExtractPkScriptAltSigType(pkScript)
		var addr types.Address
		err := fmt.Errorf("invalid signature suite for alt sig")
		switch suite {
		case ecc.EdDSA_Ed25519:
			addr, err = address.NewEdwardsPubKeyAddress(pops[0].data,
				chainParams)
		case ecc.ECDSA_SecpSchnorr:
			addr, err = address.NewSecSchnorrPubKeyAddress(pops[0].data,
				chainParams)
		}
		if err == nil {
			addrs = append(addrs, addr)
		}

	case StakeSubmissionTy:
		// A pay-to-stake-submission-hash script is of the form:
		//  OP_SSTX ... P2PKH or P2SH
		var localAddrs []types.Address
		_, localAddrs, requiredSigs, err =
			ExtractPkScriptAddrs(getStakeOutSubscript(pkScript),
				chainParams)
		if err == nil {
			addrs = append(addrs, localAddrs...)
		}

	case StakeGenTy:
		// A pay-to-stake-generation-hash script is of the form:
		//  OP_SSGEN  ... P2PKH or P2SH
		var localAddrs []types.Address
		_, localAddrs, requiredSigs, err = ExtractPkScriptAddrs(
			getStakeOutSubscript(pkScript), chainParams)
		if err == nil {
			addrs = append(addrs, localAddrs...)
		}

	case StakeRevocationTy:
		// A pay-to-stake-revocation-hash script is of the form:
		//  OP_SSRTX  ... P2PKH or P2SH
		var localAddrs []types.Address
		_, localAddrs, requiredSigs, err =
			ExtractPkScriptAddrs(getStakeOutSubscript(pkScript),
				chainParams)
		if err == nil {
			addrs = append(addrs, localAddrs...)
		}

	case StakeSubChangeTy:
		// A pay-to-stake-submission-change-hash script is of the form:
		// OP_SSTXCHANGE ... P2PKH or P2SH
		var localAddrs []types.Address
		_, localAddrs, requiredSigs, err =
			ExtractPkScriptAddrs(getStakeOutSubscript(pkScript),
				chainParams)
		if err == nil {
			addrs = append(addrs, localAddrs...)
		}

	case ScriptHashTy:
		// A pay-to-script-hash script is of the form:
		//  OP_HASH160 <scripthash> OP_EQUAL
		// Therefore the script hash is the 2nd item on the stack.
		// Skip the script hash if it's invalid for some reason.
		requiredSigs = 1
		addr, err := address.NewScriptHashAddressFromHash(pops[1].data,
			chainParams)
		if err == nil {
			addrs = append(addrs, addr)
		}

	case MultiSigTy:
		// A multi-signature script is of the form:
		//  <numsigs> <pubkey> <pubkey> <pubkey>... <numpubkeys> OP_CHECKMULTISIG
		// Therefore the number of required signatures is the 1st item
		// on the stack and the number of public keys is the 2nd to last
		// item on the stack.
		requiredSigs = asSmallInt(pops[0].opcode)
		numPubKeys := asSmallInt(pops[len(pops)-2].opcode)

		// Extract the public keys while skipping any that are invalid.
		addrs = make([]types.Address, 0, numPubKeys)
		for i := 0; i < numPubKeys; i++ {
			pubkey, err := ecc.Secp256k1.ParsePubKey(pops[i+1].data)
			if err == nil {
				addr, err := address.NewSecpPubKeyCompressedAddress(pubkey,
					chainParams)
				if err == nil {
					addrs = append(addrs, addr)
				}
			}
		}

	case NullDataTy:
		// Null data transactions have no addresses or required
		// signatures.

	case NonStandardTy:
		// Don't attempt to extract addresses or required signatures for
		// nonstandard transactions.

	case CLTVPubKeyHashTy:
		// A pay-to-cltv-pubkey-hash script is of the form:
		//  <nLockTime> OP_CHECKLOCKTIMEVERIFY OP_DROP OP_DUP OP_HASH160 <hash> OP_EQUALVERIFY OP_CHECKSIG
		// Therefore the pubkey hash is the 3rd item on the stack.
		// Skip the pubkey hash if it's invalid for some reason.
		requiredSigs = 1
		addr, err := address.NewPubKeyHashAddress(pops[5].data,
			chainParams, ecc.ECDSA_Secp256k1)
		if err == nil {
			addrs = append(addrs, addr)
		}

	case TokenPubKeyHashTy:
		requiredSigs = 1
		addr, err := address.NewPubKeyHashAddress(pops[8].data,
			chainParams, ecc.ECDSA_Secp256k1)
		if err == nil {
			addrs = append(addrs, addr)
		}
	}

	return scriptClass, addrs, requiredSigs, nil
}

// extractOneBytePush returns the value of a one byte push.
func extractOneBytePush(po ParsedOpcode) int {
	if !isOneByteMaxDataPush(po) {
		return -1
	}

	if po.opcode.value == OP_1 ||
		po.opcode.value == OP_2 ||
		po.opcode.value == OP_3 ||
		po.opcode.value == OP_4 ||
		po.opcode.value == OP_5 ||
		po.opcode.value == OP_6 ||
		po.opcode.value == OP_7 ||
		po.opcode.value == OP_8 ||
		po.opcode.value == OP_9 ||
		po.opcode.value == OP_10 ||
		po.opcode.value == OP_11 ||
		po.opcode.value == OP_12 ||
		po.opcode.value == OP_13 ||
		po.opcode.value == OP_14 ||
		po.opcode.value == OP_15 ||
		po.opcode.value == OP_16 {
		return int(po.opcode.value - 80)
	}

	return int(po.data[0])
}

// ExtractPkScriptAltSigType returns the signature scheme to use for an
// alternative check signature script.
func ExtractPkScriptAltSigType(pkScript []byte) (ecc.EcType, error) {
	pops, err := parseScript(pkScript)
	if err != nil {
		return 0, err
	}

	isPKA := isPubkeyAlt(pops)
	isPKHA := isPubkeyHashAlt(pops)
	if !(isPKA || isPKHA) {
		return -1, fmt.Errorf("wrong script type")
	}

	sigTypeLoc := 1
	if isPKHA {
		sigTypeLoc = 4
	}

	valInt := extractOneBytePush(pops[sigTypeLoc])
	if valInt < 0 {
		return 0, fmt.Errorf("bad type push")
	}
	val := sigTypes(valInt)
	switch val {
	case edwards:
		return ecc.EcType(val), nil
	case secSchnorr:
		return ecc.EcType(val), nil
	default:
		break
	}

	return -1, fmt.Errorf("bad signature scheme type")
}

// AtomicSwapDataPushes houses the data pushes found in atomic swap contracts.
type AtomicSwapDataPushes struct {
	RecipientHash160 [20]byte
	RefundHash160    [20]byte
	SecretHash       [32]byte
	SecretSize       int64
	LockTime         int64
}

// TODO, refactor & design of the Atomics Swaps of btc-alike chains & eth-alike chains
// ExtractAtomicSwapDataPushes returns the data pushes from an atomic swap
// contract.  If the script is not an atomic swap contract,
// ExtractAtomicSwapDataPushes returns (nil, nil).  Non-nil errors are returned
// for unparsable scripts.
//
// NOTE: Atomic swaps are not considered standard script types by the mempool policy
// and should be used with P2SH.  The atomic swap format is also
// expected to change to use a more secure hash function in the future.
//
// This function is only defined in the txscript package due to API limitations
// which prevent callers using txscript to parse nonstandard scripts.
func ExtractAtomicSwapDataPushes(version uint16, pkScript []byte) (*AtomicSwapDataPushes, error) {
	pops, err := parseScript(pkScript)
	if err != nil {
		return nil, err
	}

	if len(pops) != 20 {
		return nil, nil
	}
	isAtomicSwap := pops[0].opcode.value == OP_IF &&
		pops[1].opcode.value == OP_SIZE &&
		canonicalPush(pops[2]) &&
		pops[3].opcode.value == OP_EQUALVERIFY &&
		pops[4].opcode.value == OP_SHA256 &&
		pops[5].opcode.value == OP_DATA_32 &&
		pops[6].opcode.value == OP_EQUALVERIFY &&
		pops[7].opcode.value == OP_DUP &&
		pops[8].opcode.value == OP_HASH160 &&
		pops[9].opcode.value == OP_DATA_20 &&
		pops[10].opcode.value == OP_ELSE &&
		canonicalPush(pops[11]) &&
		pops[12].opcode.value == OP_CHECKLOCKTIMEVERIFY &&
		pops[13].opcode.value == OP_DROP &&
		pops[14].opcode.value == OP_DUP &&
		pops[15].opcode.value == OP_HASH160 &&
		pops[16].opcode.value == OP_DATA_20 &&
		pops[17].opcode.value == OP_ENDIF &&
		pops[18].opcode.value == OP_EQUALVERIFY &&
		pops[19].opcode.value == OP_CHECKSIG
	if !isAtomicSwap {
		return nil, nil
	}

	pushes := new(AtomicSwapDataPushes)
	copy(pushes.SecretHash[:], pops[5].data)
	copy(pushes.RecipientHash160[:], pops[9].data)
	copy(pushes.RefundHash160[:], pops[16].data)
	if pops[2].data != nil {
		locktime, err := makeScriptNum(pops[2].data, true, 5)
		if err != nil {
			return nil, nil
		}
		pushes.SecretSize = int64(locktime)
	} else if op := pops[2].opcode; isSmallInt(op) {
		pushes.SecretSize = int64(asSmallInt(op))
	} else {
		return nil, nil
	}
	if pops[11].data != nil {
		locktime, err := makeScriptNum(pops[11].data, true, 5)
		if err != nil {
			return nil, nil
		}
		pushes.LockTime = int64(locktime)
	} else if op := pops[11].opcode; isSmallInt(op) {
		pushes.LockTime = int64(asSmallInt(op))
	} else {
		return nil, nil
	}
	return pushes, nil
}
