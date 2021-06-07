// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"errors"
	"fmt"

	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/ecc"
	"github.com/Qitmeer/qitmeer/params"
)

// RawTxInSignature returns the serialized ECDSA signature for the input idx of
// the given transaction, with hashType appended to it.
func RawTxInSignature(tx *types.Transaction, idx int, subScript []byte,
	hashType SigHashType, key ecc.PrivateKey) ([]byte, error) {

	parsedScript, err := parseScript(subScript)
	if err != nil {
		return nil, fmt.Errorf("cannot parse output script: %v", err)
	}
	h, err := calcSignatureHash(parsedScript, hashType, tx, idx, nil)
	if err != nil {
		return nil, err
	}

	r, s, err := ecc.Secp256k1.Sign(key, h)
	if err != nil {
		return nil, fmt.Errorf("cannot sign tx input: %s", err)
	}
	sig := ecc.Secp256k1.NewSignature(r, s)

	return append(sig.Serialize(), byte(hashType)), nil
}

// RawTxInSignatureAlt returns the serialized ECDSA signature for the input idx of
// the given transaction, with hashType appended to it.
func RawTxInSignatureAlt(tx *types.Transaction, idx int, subScript []byte,
	hashType SigHashType, key ecc.PrivateKey, sigType sigTypes) ([]byte,
	error) {

	parsedScript, err := parseScript(subScript)
	if err != nil {
		return nil, fmt.Errorf("cannot parse output script: %v", err)
	}
	hash, err := calcSignatureHash(parsedScript, hashType, tx, idx, nil)
	if err != nil {
		return nil, err
	}

	var sig ecc.Signature
	switch sigType {
	case edwards:
		r, s, err := ecc.Ed25519.Sign(key, hash)
		if err != nil {
			return nil, fmt.Errorf("cannot sign tx input: %s", err)
		}
		sig = ecc.Ed25519.NewSignature(r, s)
	case secSchnorr:
		r, s, err := ecc.SecSchnorr.Sign(key, hash)
		if err != nil {
			return nil, fmt.Errorf("cannot sign tx input: %s", err)
		}
		sig = ecc.SecSchnorr.NewSignature(r, s)
	default:
		return nil, fmt.Errorf("unknown alt sig type %v", sigType)
	}

	return append(sig.Serialize(), byte(hashType)), nil
}

// SignatureScript creates an input signature script for tx to spend coins sent
// from a previous output to the owner of privKey. tx must include all
// transaction inputs and outputs, however txin scripts are allowed to be filled
// or empty. The returned script is calculated to be used as the idx'th txin
// sigscript for tx. subscript is the PkScript of the previous output being used
// as the idx'th input. privKey is serialized in either a compressed or
// uncompressed format based on compress. This format must match the same format
// used to generate the payment address, or the script validation will fail.
func SignatureScript(tx *types.Transaction, idx int, subscript []byte,
	hashType SigHashType, privKey ecc.PrivateKey, compress bool) ([]byte,
	error) {
	sig, err := RawTxInSignature(tx, idx, subscript, hashType, privKey)
	if err != nil {
		return nil, err
	}

	pubx, puby := privKey.Public()
	pub := ecc.Secp256k1.NewPublicKey(pubx, puby)
	var pkData []byte
	if compress {
		pkData = pub.SerializeCompressed()
	} else {
		pkData = pub.SerializeUncompressed()
	}

	return NewScriptBuilder().AddData(sig).AddData(pkData).Script()
}

// SignatureScriptAlt creates an input signature script for tx to spend coins sent
// from a previous output to the owner of privKey. tx must include all
// transaction inputs and outputs, however txin scripts are allowed to be filled
// or empty. The returned script is calculated to be used as the idx'th txin
// sigscript for tx. subscript is the PkScript of the previous output being used
// as the idx'th input. privKey is serialized in the respective format for the
// ECDSA type. This format must match the same format used to generate the payment
// address, or the script validation will fail.
func SignatureScriptAlt(tx *types.Transaction, idx int, subscript []byte,
	hashType SigHashType, privKey ecc.PrivateKey, compress bool,
	sigType int) ([]byte,
	error) {
	sig, err := RawTxInSignatureAlt(tx, idx, subscript, hashType, privKey,
		sigTypes(sigType))
	if err != nil {
		return nil, err
	}

	pubx, puby := privKey.Public()
	var pub ecc.PublicKey
	switch sigTypes(sigType) {
	case edwards:
		pub = ecc.Ed25519.NewPublicKey(pubx, puby)
	case secSchnorr:
		pub = ecc.SecSchnorr.NewPublicKey(pubx, puby)
	}
	pkData := pub.Serialize()

	return NewScriptBuilder().AddData(sig).AddData(pkData).Script()
}

// p2pkSignatureScript constructs a pay-to-pubkey signature script.
func p2pkSignatureScript(tx *types.Transaction, idx int, subScript []byte,
	hashType SigHashType, privKey ecc.PrivateKey) ([]byte, error) {
	sig, err := RawTxInSignature(tx, idx, subScript, hashType, privKey)
	if err != nil {
		return nil, err
	}

	return NewScriptBuilder().AddData(sig).Script()
}

// p2pkSignatureScript constructs a pay-to-pubkey signature script for alternative
// ECDSA types.
func p2pkSignatureScriptAlt(tx *types.Transaction, idx int, subScript []byte,
	hashType SigHashType, privKey ecc.PrivateKey, sigType sigTypes) ([]byte,
	error) {
	sig, err := RawTxInSignatureAlt(tx, idx, subScript, hashType, privKey,
		sigType)
	if err != nil {
		return nil, err
	}

	return NewScriptBuilder().AddData(sig).Script()
}

// signMultiSig signs as many of the outputs in the provided multisig script as
// possible. It returns the generated script and a boolean if the script fulfils
// the contract (i.e. nrequired signatures are provided).  Since it is arguably
// legal to not be able to sign any of the outputs, no error is returned.
func signMultiSig(tx *types.Transaction, idx int, subScript []byte, hashType SigHashType,
	addresses []types.Address, nRequired int, kdb KeyDB) ([]byte, bool) {
	// No need to add dummy.
	// TODO, revisit the bitcoin multi-sig script bug
	builder := NewScriptBuilder()
	signed := 0
	for _, addr := range addresses {
		key, _, err := kdb.GetKey(addr)
		if err != nil {
			continue
		}
		sig, err := RawTxInSignature(tx, idx, subScript, hashType, key)
		if err != nil {
			continue
		}

		builder.AddData(sig)
		signed++
		if signed == nRequired {
			break
		}
	}

	script, _ := builder.Script()
	return script, signed == nRequired
}

// handleStakeOutSign is a convenience function for reducing code clutter in
// sign. It handles the signing of stake outputs.
func handleStakeOutSign(chainParams *params.Params, tx *types.Transaction, idx int,
	subScript []byte, hashType SigHashType, kdb KeyDB, sdb ScriptDB,
	addresses []types.Address, class ScriptClass, subClass ScriptClass,
	nrequired int) ([]byte, ScriptClass, []types.Address, int, error) {

	// look up key for address
	switch subClass {
	case PubKeyHashTy:
		key, compressed, err := kdb.GetKey(addresses[0])
		if err != nil {
			return nil, class, nil, 0, err
		}
		txscript, err := SignatureScript(tx, idx, subScript, hashType,
			key, compressed)
		if err != nil {
			return nil, class, nil, 0, err
		}
		return txscript, class, addresses, nrequired, nil
	case ScriptHashTy:
		script, err := sdb.GetScript(addresses[0])
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil
	}

	return nil, class, nil, 0, fmt.Errorf("unknown subclass for stake output " +
		"to sign")
}

// sign is the main signing workhorse. It takes a script, its input transaction,
// its input index, a database of keys, a database of scripts, and information
// about the type of signature and returns a signature, script class, the
// addresses involved, and the number of signatures required.
func sign(chainParams *params.Params, tx *types.Transaction, idx int,
	subScript []byte, hashType SigHashType, kdb KeyDB, sdb ScriptDB,
	sigType sigTypes) ([]byte,
	ScriptClass, []types.Address, int, error) {

	class, addresses, nrequired, err := ExtractPkScriptAddrs(
		subScript, chainParams)
	if err != nil {
		return nil, NonStandardTy, nil, 0, err
	}

	subClass := class
	isStakeType := class == StakeSubmissionTy ||
		class == StakeSubChangeTy ||
		class == StakeGenTy ||
		class == StakeRevocationTy
	if isStakeType {
		subClass, err = GetStakeOutSubclass(subScript)
		if err != nil {
			return nil, 0, nil, 0,
				fmt.Errorf("unknown stake output subclass encountered")
		}
	}

	switch class {
	case PubKeyTy:
		// look up key for address
		key, _, err := kdb.GetKey(addresses[0])
		if err != nil {
			return nil, class, nil, 0, err
		}

		script, err := p2pkSignatureScript(tx, idx, subScript, hashType,
			key)
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil

	case PubkeyAltTy:
		// look up key for address
		key, _, err := kdb.GetKey(addresses[0])
		if err != nil {
			return nil, class, nil, 0, err
		}

		script, err := p2pkSignatureScriptAlt(tx, idx, subScript, hashType,
			key, sigType)
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil

	case PubKeyHashTy:
		// look up key for address
		key, compressed, err := kdb.GetKey(addresses[0])
		if err != nil {
			return nil, class, nil, 0, err
		}

		script, err := SignatureScript(tx, idx, subScript, hashType,
			key, compressed)
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil

	case PubkeyHashAltTy:
		// look up key for address
		key, compressed, err := kdb.GetKey(addresses[0])
		if err != nil {
			return nil, class, nil, 0, err
		}

		script, err := SignatureScriptAlt(tx, idx, subScript, hashType,
			key, compressed, int(sigType))
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil

	case ScriptHashTy:
		script, err := sdb.GetScript(addresses[0])
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil

	case MultiSigTy:
		script, _ := signMultiSig(tx, idx, subScript, hashType,
			addresses, nrequired, kdb)
		return script, class, addresses, nrequired, nil

	case StakeSubmissionTy:
		return handleStakeOutSign(chainParams, tx, idx, subScript, hashType, kdb,
			sdb, addresses, class, subClass, nrequired)

	case StakeGenTy:
		return handleStakeOutSign(chainParams, tx, idx, subScript, hashType, kdb,
			sdb, addresses, class, subClass, nrequired)

	case StakeRevocationTy:
		return handleStakeOutSign(chainParams, tx, idx, subScript, hashType, kdb,
			sdb, addresses, class, subClass, nrequired)

	case StakeSubChangeTy:
		return handleStakeOutSign(chainParams, tx, idx, subScript, hashType, kdb,
			sdb, addresses, class, subClass, nrequired)

	case NullDataTy:
		return nil, class, nil, 0,
			errors.New("can't sign NULLDATA transactions")

	case CLTVPubKeyHashTy:
		key, compressed, err := kdb.GetKey(addresses[0])
		if err != nil {
			return nil, class, nil, 0, err
		}

		script, err := SignatureScript(tx, idx, subScript, hashType,
			key, compressed)
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil

	case TokenPubKeyHashTy:
		key, compressed, err := kdb.GetKey(addresses[0])
		if err != nil {
			return nil, class, nil, 0, err
		}

		script, err := SignatureScript(tx, idx, subScript, hashType,
			key, compressed)
		if err != nil {
			return nil, class, nil, 0, err
		}

		return script, class, addresses, nrequired, nil
	default:
		return nil, class, nil, 0,
			errors.New("can't sign unknown transactions")
	}
}

// mergeScripts merges sigScript and prevScript assuming they are both
// partial solutions for pkScript spending output idx of tx. class, addresses
// and nrequired are the result of extracting the addresses from pkscript.
// The return value is the best effort merging of the two scripts. Calling this
// function with addresses, class and nrequired that do not match pkScript is
// an error and results in undefined behaviour.
func mergeScripts(chainParams *params.Params, tx *types.Transaction, idx int,
	pkScript []byte, class ScriptClass, addresses []types.Address,
	nRequired int, sigScript, prevScript []byte) []byte {

	// TODO(oga) the scripthash and multisig paths here are overly
	// inefficient in that they will recompute already known data.
	// some internal refactoring could probably make this avoid needless
	// extra calculations.
	switch class {
	case ScriptHashTy:
		// Remove the last push in the script and then recurse.
		// this could be a lot less inefficient.
		sigPops, err := parseScript(sigScript)
		if err != nil || len(sigPops) == 0 {
			return prevScript
		}
		prevPops, err := parseScript(prevScript)
		if err != nil || len(prevPops) == 0 {
			return sigScript
		}

		// assume that script in sigPops is the correct one, we just
		// made it.
		script := sigPops[len(sigPops)-1].data

		// We already know this information somewhere up the stack,
		// therefore the error is ignored.
		class, addresses, nrequired, _ :=
			ExtractPkScriptAddrs(script, chainParams)

		// regenerate scripts.
		sigScript, _ := unparseScript(sigPops)
		prevScript, _ := unparseScript(prevPops)

		// Merge
		mergedScript := mergeScripts(chainParams, tx, idx, script,
			class, addresses, nrequired, sigScript, prevScript)

		// Reappend the script and return the result.
		builder := NewScriptBuilder()
		builder.AddOps(mergedScript)
		builder.AddData(script)
		finalScript, _ := builder.Script()
		return finalScript
	case MultiSigTy:
		return mergeMultiSig(tx, idx, addresses, nRequired, pkScript,
			sigScript, prevScript)

	// It doesn't actually make sense to merge anything other than multiig
	// and scripthash (because it could contain multisig). Everything else
	// has either zero signature, can't be spent, or has a single signature
	// which is either present or not. The other two cases are handled
	// above. In the conflict case here we just assume the longest is
	// correct (this matches behaviour of the reference implementation).
	default:
		if len(sigScript) > len(prevScript) {
			return sigScript
		}
		return prevScript
	}
}

// mergeMultiSig combines the two signature scripts sigScript and prevScript
// that both provide signatures for pkScript in output idx of tx. addresses
// and nRequired should be the results from extracting the addresses from
// pkScript. Since this function is internal only we assume that the arguments
// have come from other functions internally and thus are all consistent with
// each other, behaviour is undefined if this contract is broken.
func mergeMultiSig(tx *types.Transaction, idx int, addresses []types.Address,
	nRequired int, pkScript, sigScript, prevScript []byte) []byte {

	// This is an internal only function and we already parsed this script
	// as ok for multisig (this is how we got here), so if this fails then
	// all assumptions are broken and who knows which way is up?
	pkPops, _ := parseScript(pkScript)

	sigPops, err := parseScript(sigScript)
	if err != nil || len(sigPops) == 0 {
		return prevScript
	}

	prevPops, err := parseScript(prevScript)
	if err != nil || len(prevPops) == 0 {
		return sigScript
	}

	// Convenience function to avoid duplication.
	extractSigs := func(pops []ParsedOpcode, sigs [][]byte) [][]byte {
		for _, pop := range pops {
			if len(pop.data) != 0 {
				sigs = append(sigs, pop.data)
			}
		}
		return sigs
	}

	possibleSigs := make([][]byte, 0, len(sigPops)+len(prevPops))
	possibleSigs = extractSigs(sigPops, possibleSigs)
	possibleSigs = extractSigs(prevPops, possibleSigs)

	// Now we need to match the signatures to pubkeys, the only real way to
	// do that is to try to verify them all and match it to the pubkey
	// that verifies it. we then can go through the addresses in order
	// to build our script. Anything that doesn't parse or doesn't verify we
	// throw away.
	addrToSig := make(map[string][]byte)
sigLoop:
	for _, sig := range possibleSigs {

		// can't have a valid signature that doesn't at least have a
		// hashtype, in practise it is even longer than this. but
		// that'll be checked next.
		if len(sig) < 1 {
			continue
		}
		tSig := sig[:len(sig)-1]
		hashType := SigHashType(sig[len(sig)-1])

		pSig, err := ecc.Secp256k1.ParseDERSignature(tSig)
		if err != nil {
			continue
		}

		// We have to do this each round since hash types may vary
		// between signatures and so the hash will vary. We can,
		// however, assume no sigs etc are in the script since that
		// would make the transaction nonstandard and thus not
		// MultiSigTy, so we just need to hash the full thing.
		hash, err := calcSignatureHash(pkPops, hashType, tx, idx, nil)
		if err != nil {
			// is this the right handling for SIGHASH_SINGLE error ?
			// make sure this doesn't break anything.
			// TODO revisit the SIGHASH_SINGLE design
			continue
		}

		for _, addr := range addresses {
			// All multisig addresses should be pubkey addreses
			// it is an error to call this internal function with
			// bad input.
			pkaddr := addr.(*address.SecpPubKeyAddress)

			pubKey := pkaddr.PubKey()

			// If it matches we put it in the map. We only
			// can take one signature per public key so if we
			// already have one, we can throw this away.
			r := pSig.GetR()
			s := pSig.GetS()
			if ecc.Secp256k1.Verify(pubKey, hash, r, s) {
				aStr := addr.Encode()
				if _, ok := addrToSig[aStr]; !ok {
					addrToSig[aStr] = sig
				}
				continue sigLoop
			}
		}
	}

	// Extra opcode to handle the extra arg consumed (due to previous bugs
	// in the reference implementation).
	builder := NewScriptBuilder() //.AddOp(OP_FALSE)
	doneSigs := 0
	// This assumes that addresses are in the same order as in the script.
	for _, addr := range addresses {
		sig, ok := addrToSig[addr.Encode()]
		if !ok {
			continue
		}
		builder.AddData(sig)
		doneSigs++
		if doneSigs == nRequired {
			break
		}
	}

	// padding for missing ones.
	for i := doneSigs; i < nRequired; i++ {
		builder.AddOp(OP_0)
	}

	script, _ := builder.Script()
	return script
}

// KeyDB is an interface type provided to SignTxOutput, it encapsulates
// any user state required to get the private keys for an address.
type KeyDB interface {
	GetKey(types.Address) (ecc.PrivateKey, bool, error)
}

// KeyClosure implements KeyDB with a closure.
type KeyClosure func(types.Address) (ecc.PrivateKey, bool, error)

// GetKey implements KeyDB by returning the result of calling the closure.
func (kc KeyClosure) GetKey(address types.Address) (ecc.PrivateKey,
	bool, error) {
	return kc(address)
}

// ScriptDB is an interface type provided to SignTxOutput, it encapsulates any
// user state required to get the scripts for an pay-to-script-hash address.
type ScriptDB interface {
	GetScript(types.Address) ([]byte, error)
}

// ScriptClosure implements ScriptDB with a closure.
type ScriptClosure func(types.Address) ([]byte, error)

// GetScript implements ScriptDB by returning the result of calling the closure.
func (sc ScriptClosure) GetScript(address types.Address) ([]byte, error) {
	return sc(address)
}

// SignTxOutput signs output idx of the given tx to resolve the script given in
// pkScript with a signature type of hashType. Any keys required will be
// looked up by calling getKey() with the string of the given address.
// Any pay-to-script-hash signatures will be similarly looked up by calling
// getScript. If previousScript is provided then the results in previousScript
// will be merged in a type-dependent manner with the newly generated.
// signature script.
func SignTxOutput(chainParams *params.Params, tx *types.Transaction, idx int,
	pkScript []byte, hashType SigHashType, kdb KeyDB, sdb ScriptDB,
	previousScript []byte, sigType ecc.EcType) ([]byte, error) {
	sigScript, class, addresses, nrequired, err := sign(chainParams, tx,
		idx, pkScript, hashType, kdb, sdb, sigTypes(sigType))
	if err != nil {
		return nil, err
	}

	isStakeType := class == StakeSubmissionTy ||
		class == StakeSubChangeTy ||
		class == StakeGenTy ||
		class == StakeRevocationTy
	if isStakeType {
		class, err = GetStakeOutSubclass(pkScript)
		if err != nil {
			return nil, fmt.Errorf("unknown stake output subclass encountered")
		}
	}

	if class == ScriptHashTy {
		// TODO keep the sub addressed and pass down to merge.
		realSigScript, _, _, _, err := sign(chainParams, tx, idx,
			sigScript, hashType, kdb, sdb, sigTypes(sigType))
		if err != nil {
			return nil, err
		}

		// Append the p2sh script as the last push in the script.
		builder := NewScriptBuilder()
		builder.AddOps(realSigScript)
		builder.AddData(sigScript)

		sigScript, _ = builder.Script()
		// TODO keep a copy of the script for merging.
	}

	// Merge scripts. with any previous data, if any.
	mergedScript := mergeScripts(chainParams, tx, idx, pkScript, class,
		addresses, nrequired, sigScript, previousScript)
	return mergedScript, nil
}

//TODO refactor SignTxOut remove depends on params & types.Transaction
func SignTxOut(tx types.ScriptTx, idx int,
	pkScript []byte, hashType SigHashType, kdb KeyDB, sdb ScriptDB,
	previousScript []byte, sigType ecc.EcType) ([]byte, error) {
	sigScript, class, addresses, nrequired, err := sign2(tx,
		idx, pkScript, hashType, kdb, sdb, sigTypes(sigType))
	if err != nil {
		return nil, err
	}
	// Merge scripts. with any previous data, if any.
	mergedScript := mergeScripts2(tx, idx, pkScript, class,
		addresses, nrequired, sigScript, previousScript)
	return mergedScript, nil
}

// sign2 (refactor sign)
func sign2(tx types.ScriptTx, idx int,
	subScript []byte, hashType SigHashType, kdb KeyDB, sdb ScriptDB,
	sigType sigTypes) ([]byte,
	ScriptClass, []types.Address, int, error) {

	s, err := ParsePkScript(subScript)

	if err != nil {
		return nil, NonStandardTy, nil, 0, err
	}
	class := s.GetClass()
	addresses := s.GetAddresses()
	nrequired := 0
	if s.RequiredSigs() {
		nrequired = 1
	}
	switch class {
	case PubKeyTy:
		//TODO
	case PubkeyAltTy:
		//TODO
	case PubKeyHashTy:
		// look up key for address
		key, compressed, err := kdb.GetKey(addresses[0])
		if err != nil {
			return nil, class, nil, 0, err
		}

		script, err := SignatureScript2(tx, idx, subScript, hashType,
			key, compressed)
		if err != nil {
			return nil, class, nil, 0, err
		}
		return script, class, addresses, nrequired, nil
	case PubkeyHashAltTy:
		// TODO
	case ScriptHashTy:
		// TODO
	case MultiSigTy:
		// TODO
		return nil, class, nil, 0,
			fmt.Errorf("NOT support %s transactions", class)
	case NullDataTy:
		return nil, class, nil, 0,
			errors.New("can't sign NULLDATA transactions")
	default:
		return nil, class, nil, 0,
			errors.New("can't sign unknown transactions")
	}
	//TODO should not go here
	return nil, class, nil, 0,
		fmt.Errorf("NOT support %s transactions", class)
}

// SignatureScript2, ( refactor of SignatureScript)
func SignatureScript2(tx types.ScriptTx, idx int, subscript []byte,
	hashType SigHashType, privKey ecc.PrivateKey, compress bool) ([]byte,
	error) {
	sig, err := RawTxInSignature2(tx, idx, subscript, hashType, privKey)
	if err != nil {
		return nil, err
	}

	pubx, puby := privKey.Public()
	pub := ecc.Secp256k1.NewPublicKey(pubx, puby)
	var pkData []byte
	if compress {
		pkData = pub.SerializeCompressed()
	} else {
		pkData = pub.SerializeUncompressed()
	}

	return NewScriptBuilder().AddData(sig).AddData(pkData).Script()
}

// RawTxInSignature2 (refactor of  RawTxInSignature)
func RawTxInSignature2(tx types.ScriptTx, idx int, subScript []byte,
	hashType SigHashType, key ecc.PrivateKey) ([]byte, error) {

	parsedScript, err := parseScript(subScript)
	if err != nil {
		return nil, fmt.Errorf("cannot parse output script: %v", err)
	}
	var h []byte
	// TODO, need to abstract SignatureHash calculator, instead of switch by type
	switch tx.GetType() {
	case types.QitmeerScriptTx:
		h, err = calcSignatureHash2(parsedScript, hashType, tx, idx, nil)
	}
	if err != nil {
		return nil, err
	}

	r, s, err := ecc.Secp256k1.Sign(key, h)
	if err != nil {
		return nil, fmt.Errorf("cannot sign tx input: %s", err)
	}
	sig := ecc.Secp256k1.NewSignature(r, s)

	return append(sig.Serialize(), byte(hashType)), nil
}

// calcSignatureHash2 (refactor of calcSignatureHash)
// 2 -> normal
func calcSignatureHash2(prevOutScript []ParsedOpcode, hashType SigHashType, txScript types.ScriptTx, idx int, cachedPrefix *hash.Hash) ([]byte, error) {
	// TODO, error handling
	tx, _ := txScript.(*types.Transaction)

	// The SigHashSingle signature type signs only the corresponding input
	// and output (the output with the same index number as the input).
	//
	// Since transactions can have more inputs than outputs, this means it
	// is improper to use SigHashSingle on input indices that don't have a
	// corresponding output.
	if hashType&sigHashMask == SigHashSingle && idx >= len(tx.TxOut) {
		return nil, ErrSighashSingleIdx
	}

	// Remove all instances of OP_CODESEPARATOR from the script.
	//
	// The call to unparseScript cannot fail here because removeOpcode
	// only returns a valid script.
	prevOutScript = removeOpcode(prevOutScript, OP_CODESEPARATOR)
	signScript, _ := unparseScript(prevOutScript)

	// Choose the inputs that will be committed to based on the signature
	// hash type.
	//
	// The SigHashAnyOneCanPay flag specifies that the signature will only
	// commit to the input being signed.  Otherwise, it will commit to all
	// inputs.
	txIns := tx.TxIn
	signTxInIdx := idx
	if hashType&SigHashAnyOneCanPay != 0 {
		txIns = tx.TxIn[idx : idx+1]
		signTxInIdx = 0
	}

	// The prefix hash commits to the non-witness data depending on the
	// signature hash type.  In particular, the specific inputs and output
	// semantics which are committed to are modified depending on the
	// signature hash type as follows:
	//
	// SigHashAll (and undefined signature hash types):
	//   Commits to all outputs.
	// SigHashNone:
	//   Commits to no outputs with all input sequences except the input
	//   being signed replaced with 0.
	// SigHashSingle:
	//   Commits to a single output at the same index as the input being
	//   signed.  All outputs before that index are cleared by setting the
	//   value to -1 and pkscript to nil and all outputs after that index
	//   are removed.  Like SigHashNone, all input sequences except the
	//   input being signed are replaced by 0.
	// SigHashAnyOneCanPay:
	//   Commits to only the input being signed.  Bit flag that can be
	//   combined with the other signature hash types.  Without this flag
	//   set, commits to all inputs.
	//
	// With the relevant inputs and outputs selected and the aforementioned
	// substitions, the prefix hash consists of the hash of the
	// serialization of the following fields:
	//
	// 1) txversion|(SigHashSerializePrefix<<16) (as little-endian uint32)
	// 2) number of inputs (as varint)
	// 3) per input:
	//    a) prevout hash (as little-endian uint256)
	//    b) prevout index (as little-endian uint32)
	//    c) prevout tree (as single byte)
	//    d) sequence (as little-endian uint32)
	// 4) number of outputs (as varint)
	// 5) per output:
	//    a) output amount (as little-endian uint64)
	//    b) pkscript version (as little-endian uint16)
	//    c) pkscript length (as varint)
	//    d) pkscript (as unmodified bytes)
	// 6) transaction lock time (as little-endian uint32)
	// 7) transaction expiry (as little-endian uint32)
	//
	// In addition, an optimization for SigHashAll is provided when the
	// SigHashAnyOneCanPay flag is not set.  In that case, the prefix hash
	// can be reused because only the witness data has been modified, so
	// the wasteful extra O(N^2) hash can be avoided.
	var prefixHash hash.Hash
	if params.SigHashOptimization && cachedPrefix != nil &&
		hashType&sigHashMask == SigHashAll &&
		hashType&SigHashAnyOneCanPay == 0 {

		prefixHash = *cachedPrefix
	} else {
		// Choose the outputs to commit to based on the signature hash
		// type.
		//
		// As the names imply, SigHashNone commits to no outputs and
		// SigHashSingle commits to the single output that corresponds
		// to the input being signed.  However, SigHashSingle is also a
		// bit special in that it commits to cleared out variants of all
		// outputs prior to the one being signed.  This is required by
		// consensus due to legacy reasons.
		//
		// All other signature hash types, such as SighHashAll commit to
		// all outputs.  Note that this includes undefined hash types as well.
		txOuts := tx.TxOut
		switch hashType & sigHashMask {
		case SigHashNone:
			txOuts = nil
		case SigHashSingle:
			txOuts = tx.TxOut[:idx+1]
		default:
			fallthrough
		case SigHashOld:
			fallthrough
		case SigHashAll:
			// Nothing special here.
		}

		size := sigHashPrefixSerializeSize(hashType, txIns, txOuts, idx)
		prefixBuf := make([]byte, size)

		// Commit to the version and hash serialization type.
		version := uint32(tx.Version) | uint32(SigHashSerializePrefix)<<16
		offset := putUint32LE(prefixBuf, version)

		// Commit to the relevant transaction inputs.
		offset += putVarInt(prefixBuf[offset:], uint64(len(txIns)))
		for txInIdx, txIn := range txIns {
			// Commit to the outpoint being spent.
			prevOut := &txIn.PreviousOut
			offset += copy(prefixBuf[offset:], prevOut.Hash[:])
			offset += putUint32LE(prefixBuf[offset:], prevOut.OutIndex)

			// Commit to the sequence.  In the case of SigHashNone
			// and SigHashSingle, commit to 0 for everything that is
			// not the input being signed instead.
			sequence := txIn.Sequence
			if (hashType&sigHashMask == SigHashNone ||
				hashType&sigHashMask == SigHashSingle) &&
				txInIdx != signTxInIdx {

				sequence = 0
			}
			offset += putUint32LE(prefixBuf[offset:], sequence)
		}

		// Commit to the relevant transaction outputs.
		offset += putVarInt(prefixBuf[offset:], uint64(len(txOuts)))
		for txOutIdx, txOut := range txOuts {
			// Commit to the output amount, script version, and
			// public key script.  In the case of SigHashSingle,
			// commit to an output amount of -1 and a nil public
			// key script for everything that is not the output
			// corresponding to the input being signed instead.
			coinId := txOut.Amount.Id
			value := txOut.Amount.Value
			pkScript := txOut.PkScript
			if hashType&sigHashMask == SigHashSingle && txOutIdx != idx {
				value = 0
				pkScript = nil
			}
			offset += putUint16LE(prefixBuf[offset:], uint16(coinId))
			offset += putUint64LE(prefixBuf[offset:], uint64(value))
			offset += putVarInt(prefixBuf[offset:], uint64(len(pkScript)))
			offset += copy(prefixBuf[offset:], pkScript)
		}

		// Commit to the lock time and expiry.
		offset += putUint32LE(prefixBuf[offset:], tx.LockTime)
		putUint32LE(prefixBuf[offset:], tx.Expire)

		prefixHash = hash.HashH(prefixBuf)
	}

	// The witness hash commits to the input witness data depending on
	// whether or not the signature hash type has the SigHashAnyOneCanPay
	// flag set.  In particular the semantics are as follows:
	//
	// SigHashAnyOneCanPay:
	//   Commits to only the input being signed.  Without this flag set,
	//   commits to all inputs with the reference scripts cleared by setting
	//   them to nil.
	//
	// With the relevant inputs selected, and the aforementioned substitutions,
	// the witness hash consists of the hash of the serialization of the
	// following fields:
	//
	// 1) txversion|(SigHashSerializeWitness<<16) (as little-endian uint32)
	// 2) number of inputs (as varint)
	// 3) per input:
	//    a) length of prevout pkscript (as varint)
	//    b) prevout pkscript (as unmodified bytes)

	size := sigHashWitnessSerializeSize(hashType, txIns, signScript)
	witnessBuf := make([]byte, size)

	// Commit to the version and hash serialization type.
	version := uint32(tx.Version) | uint32(SigHashSerializeWitness)<<16
	offset := putUint32LE(witnessBuf, version)

	// Commit to the relevant transaction inputs.
	offset += putVarInt(witnessBuf[offset:], uint64(len(txIns)))
	for txInIdx := range txIns {
		// Commit to the input script at the index corresponding to the
		// input index being signed.  Otherwise, commit to a nil script
		// instead.
		commitScript := signScript
		if txInIdx != signTxInIdx {
			commitScript = nil
		}
		offset += putVarInt(witnessBuf[offset:], uint64(len(commitScript)))
		offset += copy(witnessBuf[offset:], commitScript)
	}

	witnessHash := hash.HashH(witnessBuf)

	// The final signature hash (message to sign) is the hash of the
	// serialization of the following fields:
	//
	// 1) the hash type (as little-endian uint32)
	// 2) prefix hash (as produced by hash function)
	// 3) witness hash (as produced by hash function)
	sigHashBuf := make([]byte, hash.HashSize*2+4)
	offset = putUint32LE(sigHashBuf, uint32(hashType))
	offset += copy(sigHashBuf[offset:], prefixHash[:])
	copy(sigHashBuf[offset:], witnessHash[:])
	return hash.HashB(sigHashBuf), nil
}

/*
func shallowCopyTx(tx types.ScriptTx) (types.Transaction,error){
	txCopy := types.Transaction{
		Version: tx.GetVersion(),

	}
	txIns := make([]types.TxInput, len(tx.GetInput()))
	for i, oldTxIn := range tx.GetInput() {
		in, ok := oldTxIn.(*types.TxInput);
		if !ok {
			return txCopy, fmt.Errorf("fail to convert %v to TxIN",oldTxIn)
		}
		txIns[i] = *in
		txCopy.TxIn[i] = &txIns[i]
	}
	txOuts := make([]types.TxOutput, len(tx.GetOutput()))
	for i, oldTxOut := range tx.GetOutput() {
		out, ok := oldTxOut.(*types.TxOutput)
		if !ok {
			return txCopy, fmt.Errorf("fail to convert %v to TxOut",oldTxOut)
		}
		txOuts[i] = *out
		txCopy.TxOut[i] = &txOuts[i]
	}
	return txCopy,nil
}
*/

// mergeScripts2 (refactor mergeScript)
func mergeScripts2(tx types.ScriptTx, idx int,
	pkScript []byte, class ScriptClass, addresses []types.Address,
	nRequired int, sigScript, prevScript []byte) []byte {

	// TODO(oga) the scripthash and multisig paths here are overly
	// inefficient in that they will recompute already known data.
	// some internal refactoring could probably make this avoid needless
	// extra calculations.
	switch class {
	case ScriptHashTy:
		// Remove the last push in the script and then recurse.
		// this could be a lot less inefficient.
		sigPops, err := parseScript(sigScript)
		if err != nil || len(sigPops) == 0 {
			return prevScript
		}
		prevPops, err := parseScript(prevScript)
		if err != nil || len(prevPops) == 0 {
			return sigScript
		}

		// assume that script in sigPops is the correct one, we just
		// made it.
		script := sigPops[len(sigPops)-1].data

		// We already know this information somewhere up the stack,
		// therefore the error is ignored.
		s, _ := ParsePkScript(script)
		class := s.GetClass()
		addresses := s.GetAddresses()
		nrequired := 0
		if s.RequiredSigs() {
			nrequired = 1
		}
		// regenerate scripts.
		sigScript, _ := unparseScript(sigPops)
		prevScript, _ := unparseScript(prevPops)

		// Merge
		mergedScript := mergeScripts2(tx, idx, script,
			class, addresses, nrequired, sigScript, prevScript)

		// Reappend the script and return the result.
		builder := NewScriptBuilder()
		builder.AddOps(mergedScript)
		builder.AddData(script)
		finalScript, _ := builder.Script()
		return finalScript
	case MultiSigTy:
		return mergeMultiSig2(tx, idx, addresses, nRequired, pkScript,
			sigScript, prevScript)

	// It doesn't actually make sense to merge anything other than multiig
	// and scripthash (because it could contain multisig). Everything else
	// has either zero signature, can't be spent, or has a single signature
	// which is either present or not. The other two cases are handled
	// above. In the conflict case here we just assume the longest is
	// correct (this matches behaviour of the reference implementation).
	default:
		if len(sigScript) > len(prevScript) {
			return sigScript
		}
		return prevScript
	}
}

// mergeMultiSig2 (refactor of mergeMultiSig)
func mergeMultiSig2(tx types.ScriptTx, idx int, addresses []types.Address,
	nRequired int, pkScript, sigScript, prevScript []byte) []byte {

	// This is an internal only function and we already parsed this script
	// as ok for multisig (this is how we got here), so if this fails then
	// all assumptions are broken and who knows which way is up?
	pkPops, _ := parseScript(pkScript)

	sigPops, err := parseScript(sigScript)
	if err != nil || len(sigPops) == 0 {
		return prevScript
	}

	prevPops, err := parseScript(prevScript)
	if err != nil || len(prevPops) == 0 {
		return sigScript
	}

	// Convenience function to avoid duplication.
	extractSigs := func(pops []ParsedOpcode, sigs [][]byte) [][]byte {
		for _, pop := range pops {
			if len(pop.data) != 0 {
				sigs = append(sigs, pop.data)
			}
		}
		return sigs
	}

	possibleSigs := make([][]byte, 0, len(sigPops)+len(prevPops))
	possibleSigs = extractSigs(sigPops, possibleSigs)
	possibleSigs = extractSigs(prevPops, possibleSigs)

	// Now we need to match the signatures to pubkeys, the only real way to
	// do that is to try to verify them all and match it to the pubkey
	// that verifies it. we then can go through the addresses in order
	// to build our script. Anything that doesn't parse or doesn't verify we
	// throw away.
	addrToSig := make(map[string][]byte)
sigLoop:
	for _, sig := range possibleSigs {

		// can't have a valid signature that doesn't at least have a
		// hashtype, in practise it is even longer than this. but
		// that'll be checked next.
		if len(sig) < 1 {
			continue
		}
		tSig := sig[:len(sig)-1]
		hashType := SigHashType(sig[len(sig)-1])

		pSig, err := ecc.Secp256k1.ParseDERSignature(tSig)
		if err != nil {
			continue
		}

		// We have to do this each round since hash types may vary
		// between signatures and so the hash will vary. We can,
		// however, assume no sigs etc are in the script since that
		// would make the transaction nonstandard and thus not
		// MultiSigTy, so we just need to hash the full thing.
		var h []byte
		// TODO, need to abstract SignatureHash calculator, instead of switch by type
		switch tx.GetType() {
		case types.QitmeerScriptTx:
			h, err = calcSignatureHash2(pkPops, hashType, tx, idx, nil)
		}
		if err != nil {
			// is this the right handling for SIGHASH_SINGLE error ?
			// make sure this doesn't break anything.
			// TODO revisit the SIGHASH_SINGLE design
			continue
		}

		for _, addr := range addresses {
			// All multisig addresses should be pubkey addreses
			// it is an error to call this internal function with
			// bad input.
			pkaddr := addr.(*address.SecpPubKeyAddress)

			pubKey := pkaddr.PubKey()

			// If it matches we put it in the map. We only
			// can take one signature per public key so if we
			// already have one, we can throw this away.
			r := pSig.GetR()
			s := pSig.GetS()
			if ecc.Secp256k1.Verify(pubKey, h, r, s) {
				aStr := addr.Encode()
				if _, ok := addrToSig[aStr]; !ok {
					addrToSig[aStr] = sig
				}
				continue sigLoop
			}
		}
	}

	// Extra opcode to handle the extra arg consumed (due to previous bugs
	// in the reference implementation).
	builder := NewScriptBuilder() //.AddOp(OP_FALSE)
	doneSigs := 0
	// This assumes that addresses are in the same order as in the script.
	for _, addr := range addresses {
		sig, ok := addrToSig[addr.Encode()]
		if !ok {
			continue
		}
		builder.AddData(sig)
		doneSigs++
		if doneSigs == nRequired {
			break
		}
	}

	// padding for missing ones.
	for i := doneSigs; i < nRequired; i++ {
		builder.AddOp(OP_0)
	}

	script, _ := builder.Script()
	return script
}
