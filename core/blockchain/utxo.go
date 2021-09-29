// Copyright (c) 2017-2018 The qitmeer developers
package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockdag"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/params"
	"sync"
)

// txoFlags is a bitmask defining additional information and state for a
// transaction output in a utxo view.
type txoFlags uint8

const (
	// tfCoinBase indicates that a txout was contained in a coinbase tx.
	tfCoinBase txoFlags = 1 << iota

	// tfSpent indicates that a txout is spent.
	tfSpent

	// tfModified indicates that a txout has been modified since it was
	// loaded.
	tfModified
)

// utxoOutput houses details about an individual unspent transaction output such
// as whether or not it is spent, its public key script, and how much it pays.
//
// Standard public key scripts are stored in the database using a compressed
// format. Since the vast majority of scripts are of the standard form, a fairly
// significant savings is achieved by discarding the portions of the standard
// scripts that can be reconstructed.
//
// Also, since it is common for only a specific output in a given utxo entry to
// be referenced from a redeeming transaction, the script and amount for a given
// output is not uncompressed until the first time it is accessed.  This
// provides a mechanism to avoid the overhead of needlessly uncompressing all
// outputs for a given utxo entry at the time of load.
//
// The struct is aligned for memory efficiency.
type UtxoEntry struct {
	amount      types.Amount // The amount of the output.
	pkScript    []byte       // The public key script for the output.
	blockHash   hash.Hash
	packedFlags txoFlags
}

// isModified returns whether or not the output has been modified since it was
// loaded.
func (entry *UtxoEntry) isModified() bool {
	return entry.packedFlags&tfModified == tfModified
}

// IsCoinBase returns whether or not the output was contained in a coinbase
// transaction.
func (entry *UtxoEntry) IsCoinBase() bool {
	return entry.packedFlags&tfCoinBase == tfCoinBase
}

// BlockHash returns the hash of the block containing the output.
func (entry *UtxoEntry) BlockHash() *hash.Hash {
	return &entry.blockHash
}

// IsSpent returns whether or not the output has been spent based upon the
// current state of the unspent transaction output view it was obtained from.
func (entry *UtxoEntry) IsSpent() bool {
	return entry.packedFlags&tfSpent == tfSpent
}

// Spend marks the output as spent.  Spending an output that is already spent
// has no effect.
func (entry *UtxoEntry) Spend() {
	// Nothing to do if the output is already spent.
	if entry.IsSpent() {
		return
	}

	// Mark the output as spent and modified.
	entry.packedFlags |= tfSpent | tfModified
}

// Amount returns the amount of the output.
func (entry *UtxoEntry) Amount() types.Amount {
	return entry.amount
}

// PkScript returns the public key script for the output.
func (entry *UtxoEntry) PkScript() []byte {
	return entry.pkScript
}

// Clone returns a shallow copy of the utxo entry.
func (entry *UtxoEntry) Clone() *UtxoEntry {
	if entry == nil {
		return nil
	}

	return &UtxoEntry{
		amount:      entry.amount,
		pkScript:    entry.pkScript,
		blockHash:   entry.blockHash,
		packedFlags: entry.packedFlags,
	}
}

// UtxoViewpoint represents a view into the set of unspent transaction outputs
// from a specific point of view in the chain.  For example, it could be for
// the end of the main chain, some point in the history of the main chain, or
// down a side chain.
//
// The unspent outputs are needed by other transactions for things such as
// script validation and double spend prevention.
type UtxoViewpoint struct {
	entries    map[types.TxOutPoint]*UtxoEntry
	viewpoints []*hash.Hash
}

// NewUtxoViewpoint returns a new empty unspent transaction output view.
func NewUtxoViewpoint() *UtxoViewpoint {
	return &UtxoViewpoint{
		entries: make(map[types.TxOutPoint]*UtxoEntry),
	}
}

func (view *UtxoViewpoint) RemoveEntry(outpoint types.TxOutPoint) {
	delete(view.entries, outpoint)
}

func (view *UtxoViewpoint) Clean() {
	view.entries = map[types.TxOutPoint]*UtxoEntry{}
}

// Entries returns the underlying map that stores of all the utxo entries.
func (view *UtxoViewpoint) Entries() map[types.TxOutPoint]*UtxoEntry {
	return view.entries
}

func (view *UtxoViewpoint) AddTxOut(tx *types.Tx, txOutIdx uint32, blockHash *hash.Hash) {
	// Can't add an output for an out of bounds index.
	if txOutIdx >= uint32(len(tx.Tx.TxOut)) {
		return
	}

	// Update existing entries.  All fields are updated because it's
	// possible (although extremely unlikely) that the existing entry is
	// being replaced by a different transaction with the same hash.  This
	// is allowed so long as the previous transaction is fully spent.
	prevOut := types.TxOutPoint{Hash: *tx.Hash(), OutIndex: txOutIdx}
	txOut := tx.Tx.TxOut[txOutIdx]
	view.addTxOut(prevOut, txOut, tx.Tx.IsCoinBase(), blockHash)
}

// AddTxOuts adds all outputs in the passed transaction which are not provably
// unspendable to the view.  When the view already has entries for any of the
// outputs, they are simply marked unspent.  All fields will be updated for
// existing entries since it's possible it has changed during a reorg.
func (view *UtxoViewpoint) AddTxOuts(tx *types.Tx, blockHash *hash.Hash) {
	// Loop all of the transaction outputs and add those which are not
	// provably unspendable.
	isCoinBase := tx.Tx.IsCoinBase()
	prevOut := types.TxOutPoint{Hash: *tx.Hash()}
	for txOutIdx, txOut := range tx.Tx.TxOut {
		// Update existing entries.  All fields are updated because it's
		// possible (although extremely unlikely) that the existing
		// entry is being replaced by a different transaction with the
		// same hash.  This is allowed so long as the previous
		// transaction is fully spent.
		prevOut.OutIndex = uint32(txOutIdx)
		view.addTxOut(prevOut, txOut, isCoinBase, blockHash)
	}
}

func (view *UtxoViewpoint) addTxOut(outpoint types.TxOutPoint, txOut *types.TxOutput, isCoinBase bool, blockHash *hash.Hash) {
	// Don't add provably unspendable outputs.
	if txscript.IsUnspendable(txOut.PkScript) {
		return
	}

	// Update existing entries.  All fields are updated because it's
	// possible (although extremely unlikely) that the existing entry is
	// being replaced by a different transaction with the same hash.  This
	// is allowed so long as the previous transaction is fully spent.
	entry := view.LookupEntry(outpoint)
	if entry == nil {
		entry = new(UtxoEntry)
		view.entries[outpoint] = entry
	}

	entry.amount = txOut.Amount
	entry.pkScript = txOut.PkScript
	entry.blockHash = *blockHash
	entry.packedFlags = tfModified
	if isCoinBase {
		entry.packedFlags |= tfCoinBase
	}
}

func (view *UtxoViewpoint) AddTokenTxOut(outpoint types.TxOutPoint, pkscript []byte) {
	entry := view.LookupEntry(outpoint)
	if entry == nil {
		entry = new(UtxoEntry)
		view.entries[outpoint] = entry
	}
	if len(pkscript) <= 0 {
		pkscript = params.ActiveNetParams.Params.TokenAdminPkScript
	}
	txOut := &types.TxOutput{PkScript: pkscript}
	entry.amount = txOut.Amount
	entry.pkScript = txOut.PkScript
	entry.packedFlags = tfModified
}

// Viewpoints returns the hash of the viewpoint block in the chain the view currently
// respresents.
func (view *UtxoViewpoint) Viewpoints() []*hash.Hash {
	return view.viewpoints
}

// SetViewpoints sets the hash of the viewpoint block in the chain the view currently
// respresents.
func (view *UtxoViewpoint) SetViewpoints(views []*hash.Hash) {
	view.viewpoints = views
}

// fetchUtxosMain fetches unspent transaction output data about the provided
// set of transactions from the point of view of the end of the main chain at
// the time of the call.
//
// Upon completion of this function, the view will contain an entry for each
// requested transaction.  Fully spent transactions, or those which otherwise
// don't exist, will result in a nil entry in the view.
func (view *UtxoViewpoint) fetchUtxosMain(db database.DB, outpoints map[types.TxOutPoint]struct{}) error {
	// Nothing to do if there are no requested hashes.
	if len(outpoints) == 0 {
		return nil
	}

	// Load the unspent transaction output information for the requested set
	// of transactions from the point of view of the end of the main chain.
	//
	// NOTE: Missing entries are not considered an error here and instead
	// will result in nil entries in the view.  This is intentionally done
	// since other code uses the presence of an entry in the store as a way
	// to optimize spend and unspend updates to apply only to the specific
	// utxos that the caller needs access to.
	return db.View(func(dbTx database.Tx) error {
		for outpoint := range outpoints {
			entry, err := dbFetchUtxoEntry(dbTx, outpoint)
			if err != nil {
				return err
			}
			if entry == nil {
				continue
			}
			view.entries[outpoint] = entry
		}

		return nil
	})
}

func (view *UtxoViewpoint) FilterInvalidOut(bc *BlockChain) {
	for outpoint, entry := range view.entries {
		if !bc.IsInvalidOut(entry) {
			continue
		}
		delete(view.entries, outpoint)
	}
}

// LookupEntry returns information about a given transaction according to the
// current state of the view.  It will return nil if the passed transaction
// hash does not exist in the view or is otherwise not available such as when
// it has been disconnected during a reorg.
func (view *UtxoViewpoint) LookupEntry(outpoint types.TxOutPoint) *UtxoEntry {
	entry, ok := view.entries[outpoint]
	if !ok {
		return nil
	}

	return entry
}

func (view *UtxoViewpoint) FetchInputUtxos(db database.DB, block *types.SerializedBlock, bc *BlockChain) error {
	return view.fetchInputUtxos(db, block, bc)
}

// fetchInputUtxos loads utxo details about the input transactions referenced
// by the transactions in the given block into the view from the database as
// needed.  In particular, referenced entries that are earlier in the block are
// added to the view and entries that are already in the view are not modified.
// TODO, revisit the usage on the parent block
func (view *UtxoViewpoint) fetchInputUtxos(db database.DB, block *types.SerializedBlock, bc *BlockChain) error {
	// Build a map of in-flight transactions because some of the inputs in
	// this block could be referencing other transactions earlier in this
	// block which are not yet in the chain.
	txInFlight := map[hash.Hash]int{}
	transactions := block.Transactions()
	for i, tx := range transactions {
		if tx.IsDuplicate || types.IsTokenTx(tx.Tx) {
			continue
		}
		txInFlight[*tx.Hash()] = i
	}

	// Loop through all of the transaction inputs (except for the coinbase
	// which has no inputs) collecting them into sets of what is needed and
	// what is already known (in-flight).
	txNeededSet := make(map[types.TxOutPoint]struct{})
	for i, tx := range transactions[1:] {
		if tx.IsDuplicate {
			continue
		}
		if types.IsTokenTx(tx.Tx) && !types.IsTokenMintTx(tx.Tx) {
			continue
		}

		for txInIdx, txIn := range tx.Transaction().TxIn {
			if txInIdx == 0 && types.IsTokenMintTx(tx.Tx) {
				continue
			}
			// It is acceptable for a transaction input to reference
			// the output of another transaction in this block only
			// if the referenced transaction comes before the
			// current one in this block.  Add the outputs of the
			// referenced transaction as available utxos when this
			// is the case.  Otherwise, the utxo details are still
			// needed.
			//
			// NOTE: The >= is correct here because i is one less
			// than the actual position of the transaction within
			// the block due to skipping the coinbase.
			originHash := &txIn.PreviousOut.Hash
			if inFlightIndex, ok := txInFlight[*originHash]; ok &&
				i >= inFlightIndex {

				originTx := transactions[inFlightIndex]
				view.AddTxOuts(originTx, block.Hash())
				continue
			}

			// Don't request entries that are already in the view
			// from the database.
			if _, ok := view.entries[txIn.PreviousOut]; ok {
				continue
			}

			txNeededSet[txIn.PreviousOut] = struct{}{}
		}
	}
	err := view.fetchUtxosMain(db, txNeededSet)
	if err != nil {
		return err
	}
	view.FilterInvalidOut(bc)
	// Request the input utxos from the database.
	return nil

}

// fetchUtxos loads the unspent transaction outputs for the provided set of
// outputs into the view from the database as needed unless they already exist
// in the view in which case they are ignored.
func (view *UtxoViewpoint) fetchUtxos(db database.DB, outpoints map[types.TxOutPoint]struct{}) error {
	// Nothing to do if there are no requested outputs.
	if len(outpoints) == 0 {
		return nil
	}

	// Filter entries that are already in the view.
	neededSet := make(map[types.TxOutPoint]struct{})
	for outpoint := range outpoints {
		// Already loaded into the current view.
		if _, ok := view.entries[outpoint]; ok {
			continue
		}

		neededSet[outpoint] = struct{}{}
	}

	// Request the input utxos from the database.
	return view.fetchUtxosMain(db, neededSet)
}

// connectTransaction updates the view by adding all new utxos created by the
// passed transaction and marking all utxos that the transactions spend as
// spent.  In addition, when the 'stxos' argument is not nil, it will be updated
// to append an entry for each spent txout.  An error will be returned if the
// view does not contain the required utxos.
func (view *UtxoViewpoint) connectTransaction(tx *types.Tx, node *BlockNode, blockIndex uint32, stxos *[]SpentTxOut, bc *BlockChain) error {
	msgTx := tx.Transaction()
	// Coinbase transactions don't have any inputs to spend.
	if msgTx.IsCoinBase() {
		// Add the transaction's outputs as available utxos.
		view.AddTxOuts(tx, node.GetHash()) //TODO, remove type conversion
		return nil
	}

	// Spend the referenced utxos by marking them spent in the view and,
	// if a slice was provided for the spent txout details, append an entry
	// to it.
	for txInIndex, txIn := range msgTx.TxIn {
		if txInIndex == 0 && types.IsTokenMintTx(tx.Tx) {
			continue
		}
		entry := view.entries[txIn.PreviousOut]

		// Ensure the referenced utxo exists in the view.  This should
		// never happen unless there is a bug is introduced in the code.
		if entry == nil {
			return AssertError(fmt.Sprintf("view missing input %v",
				txIn.PreviousOut))
		}
		entry.Spend()

		// Don't create the stxo details if not requested.
		if stxos == nil {
			continue
		}

		// Populate the stxo details using the utxo entry.  When the
		// transaction is fully spent, set the additional stxo fields
		// accordingly since those details will no longer be available
		// in the utxo set.
		var stxo = SpentTxOut{
			Amount:     entry.Amount(),
			Fees:       types.Amount{Value: 0, Id: entry.Amount().Id},
			PkScript:   entry.PkScript(),
			BlockHash:  entry.blockHash,
			IsCoinBase: entry.IsCoinBase(),
			TxIndex:    uint32(tx.Index()),
			TxInIndex:  uint32(txInIndex),
		}
		if stxo.IsCoinBase && !entry.BlockHash().IsEqual(bc.params.GenesisHash) {
			if txIn.PreviousOut.OutIndex == CoinbaseOutput_subsidy ||
				entry.Amount().Id != types.MEERID {
				stxo.Fees.Value = bc.GetFeeByCoinID(&stxo.BlockHash, stxo.Fees.Id)
			}
		}
		// Append the entry to the provided spent txouts slice.
		*stxos = append(*stxos, stxo)
	}

	// Add the transaction's outputs as available utxos.
	view.AddTxOuts(tx, node.GetHash()) //TODO, remove type conversion

	return nil
}

// disconnectTransactions updates the view by removing all of the transactions
// created by the passed block, restoring all utxos the transactions spent by
// using the provided spent txo information, and setting the best hash for the
// view to the block before the passed block.
//
// This function will ONLY work correctly for a single transaction tree at a
// time because of index tracking.
func (view *UtxoViewpoint) disconnectTransactions(block *types.SerializedBlock, stxos []SpentTxOut, bc *BlockChain) error {
	// Sanity check the correct number of stxos are provided.
	if len(stxos) != bc.countSpentOutputs(block) {
		return AssertError("disconnectTransactions called with bad " +
			"spent transaction out information")
	}

	stxoIdx := len(stxos) - 1
	transactions := block.Transactions()
	for txIdx := len(transactions) - 1; txIdx > -1; txIdx-- {
		tx := transactions[txIdx]
		if tx.IsDuplicate {
			continue
		}
		if types.IsTokenTx(tx.Tx) {
			if !types.IsTokenMintTx(tx.Tx) {
				continue
			}
		}
		var packedFlags txoFlags
		isCoinBase := txIdx == 0
		if isCoinBase {
			packedFlags |= tfCoinBase
		}

		txHash := tx.Hash()
		prevOut := types.TxOutPoint{Hash: *txHash}
		for txOutIdx, txOut := range tx.Tx.TxOut {
			if txscript.IsUnspendable(txOut.PkScript) {
				continue
			}

			prevOut.OutIndex = uint32(txOutIdx)
			entry := view.entries[prevOut]
			if entry == nil {
				entry = &UtxoEntry{
					amount:      txOut.Amount,
					pkScript:    txOut.PkScript,
					blockHash:   *block.Hash(),
					packedFlags: packedFlags,
				}

				view.entries[prevOut] = entry
			}

			entry.Spend()
		}

		if isCoinBase {
			continue
		}
		for txInIdx := len(tx.Tx.TxIn) - 1; txInIdx > -1; txInIdx-- {
			if types.IsTokenMintTx(tx.Tx) && txInIdx == 0 {
				continue
			}
			stxo := &stxos[stxoIdx]
			stxoIdx--

			originOut := &tx.Tx.TxIn[txInIdx].PreviousOut
			entry := view.entries[*originOut]
			if entry == nil {
				entry = new(UtxoEntry)
				view.entries[*originOut] = entry
			}

			entry.amount = stxo.Amount
			entry.pkScript = stxo.PkScript
			entry.blockHash = stxo.BlockHash
			entry.packedFlags = tfModified
			if stxo.IsCoinBase {
				entry.packedFlags |= tfCoinBase
			}
		}
	}

	view.SetViewpoints(nil)
	return nil
}

// commit prunes all entries marked modified that are now fully spent and marks
// all entries as unmodified.
func (view *UtxoViewpoint) commit() {
	for outpoint, entry := range view.entries {
		if entry == nil || (entry.isModified() && entry.IsSpent()) {
			delete(view.entries, outpoint)
			continue
		}

		entry.packedFlags ^= tfModified
	}
}

func (bc *BlockChain) IsInvalidOut(entry *UtxoEntry) bool {
	if entry == nil {
		return true
	}
	if entry.blockHash.IsEqual(&hash.ZeroHash) {
		return false
	}
	node := bc.BlockDAG().GetBlock(&entry.blockHash)
	if node != nil {
		if !node.GetStatus().KnownInvalid() {
			return false
		}
	}
	return true
}

// FetchUtxoView loads utxo details about the input transactions referenced by
// the passed transaction from the point of view of the end of the main chain.
// It also attempts to fetch the utxo details for the transaction itself so the
// returned view can be examined for duplicate unspent transaction outputs.
//
// This function is safe for concurrent access however the returned view is NOT.
func (b *BlockChain) FetchUtxoView(tx *types.Tx) (*UtxoViewpoint, error) {
	// Create a set of needed transactions based on those referenced by the
	// inputs of the passed transaction.  Also, add the passed transaction
	// itself as a way for the caller to detect duplicates that are not
	// fully spent.
	neededSet := make(map[types.TxOutPoint]struct{})
	prevOut := types.TxOutPoint{Hash: *tx.Hash()}
	for txOutIdx := range tx.Tx.TxOut {
		prevOut.OutIndex = uint32(txOutIdx)
		neededSet[prevOut] = struct{}{}
	}
	if !tx.Tx.IsCoinBase() {
		for _, txIn := range tx.Tx.TxIn {
			neededSet[txIn.PreviousOut] = struct{}{}
		}
	}

	// Request the utxos from the point of view of the end of the main
	// chain.
	view := NewUtxoViewpoint()
	view.SetViewpoints(b.GetMiningTips(blockdag.MaxPriority))
	b.ChainRLock()
	err := view.fetchUtxosMain(b.db, neededSet)
	b.ChainRUnlock()
	if err != nil {
		return view, err
	}
	view.FilterInvalidOut(b)
	return view, err
}

// FetchUtxoEntry loads and returns the unspent transaction output entry for the
// passed hash from the point of view of the end of the main chain.
//
// NOTE: Requesting a hash for which there is no data will NOT return an error.
// Instead both the entry and the error will be nil.  This is done to allow
// pruning of fully spent transactions.  In practice this means the caller must
// check if the returned entry is nil before invoking methods on it.
//
// This function is safe for concurrent access however the returned entry (if
// any) is NOT.
func (b *BlockChain) FetchUtxoEntry(outpoint types.TxOutPoint) (*UtxoEntry, error) {
	b.ChainRLock()
	defer b.ChainRUnlock()

	var entry *UtxoEntry
	err := b.db.View(func(dbTx database.Tx) error {
		var err error
		entry, err = dbFetchUtxoEntry(dbTx, outpoint)
		return err
	})
	if err != nil {
		return nil, err
	}
	if b.IsInvalidOut(entry) {
		entry = nil
	}
	return entry, nil
}

// dbFetchUtxoEntry uses an existing database transaction to fetch all unspent
// outputs for the provided Bitcoin transaction hash from the utxo set.
//
// When there is no entry for the provided hash, nil will be returned for the
// both the entry and the error.
func dbFetchUtxoEntry(dbTx database.Tx, outpoint types.TxOutPoint) (*UtxoEntry, error) {
	// Fetch the unspent transaction output information for the passed
	// transaction output.  Return now when there is no entry.
	key := outpointKey(outpoint)
	utxoBucket := dbTx.Metadata().Bucket(dbnamespace.UtxoSetBucketName)
	serializedUtxo := utxoBucket.Get(*key)
	recycleOutpointKey(key)
	if serializedUtxo == nil {
		return nil, nil
	}

	// A non-nil zero-length entry means there is an entry in the database
	// for a spent transaction output which should never be the case.
	if len(serializedUtxo) == 0 {
		return nil, AssertError(fmt.Sprintf("database contains entry "+
			"for spent tx output %v", outpoint))
	}

	// Deserialize the utxo entry and return it.
	entry, err := DeserializeUtxoEntry(serializedUtxo)
	if err != nil {
		// Ensure any deserialization errors are returned as database
		// corruption errors.
		if isDeserializeErr(err) {
			return nil, database.Error{
				ErrorCode: database.ErrCorruption,
				Description: fmt.Sprintf("corrupt utxo entry "+
					"for %v: %v", outpoint, err),
			}
		}

		return nil, err
	}

	return entry, nil
}

func dbPutUtxoView(dbTx database.Tx, view *UtxoViewpoint) error {
	utxoBucket := dbTx.Metadata().Bucket(dbnamespace.UtxoSetBucketName)
	for outpoint, entry := range view.entries {
		// No need to update the database if the entry was not modified.
		if entry == nil || !entry.isModified() {
			continue
		}

		// Remove the utxo entry if it is spent.
		if entry.IsSpent() {
			key := outpointKey(outpoint)
			err := utxoBucket.Delete(*key)
			recycleOutpointKey(key)
			if err != nil {
				return err
			}

			continue
		}

		// Serialize and store the utxo entry.
		serialized, err := serializeUtxoEntry(entry)
		if err != nil {
			return err
		}
		key := outpointKey(outpoint)
		err = utxoBucket.Put(*key, serialized)
		// NOTE: The key is intentionally not recycled here since the
		// database interface contract prohibits modifications.  It will
		// be garbage collected normally when the database is done with
		// it.
		if err != nil {
			return err
		}
	}

	return nil
}

const UtxoEntryAmountCoinIDSize = 2

// deserializeUtxoEntry decodes a utxo entry from the passed serialized byte
// slice into a new UtxoEntry using a format that is suitable for long-term
// storage.  The format is described in detail above.
func DeserializeUtxoEntry(serialized []byte) (*UtxoEntry, error) {
	// Deserialize the header code.
	code, offset := serialization.DeserializeVLQ(serialized)
	if offset >= len(serialized) {
		return nil, errDeserialize("unexpected end of data after header")
	}

	// Decode the header code.
	//
	// Bit 0 indicates whether the containing transaction is a coinbase.
	// Bits 1-x encode id of containing transaction.
	isCoinBase := code&0x01 != 0

	blockHash, err := hash.NewHash(serialized[offset : offset+hash.HashSize])
	if err != nil {
		return nil, errDeserialize(fmt.Sprintf("unable to decode "+
			"utxo: %v", err))
	}
	offset += hash.HashSize
	// Decode amount coinId
	// Decode amount coinId
	amountCoinId := byteOrder.Uint16(serialized[offset : offset+SpentTxoutAmountCoinIDSize])
	offset += SpentTxoutAmountCoinIDSize
	// Decode the compressed unspent transaction output.
	amount, pkScript, _, err := decodeCompressedTxOut(serialized[offset:])
	if err != nil {
		return nil, errDeserialize(fmt.Sprintf("unable to decode "+
			"utxo: %v", err))
	}

	entry := &UtxoEntry{
		amount:      types.Amount{Value: int64(amount), Id: types.CoinID(amountCoinId)},
		pkScript:    pkScript,
		blockHash:   *blockHash,
		packedFlags: 0,
	}
	if isCoinBase {
		entry.packedFlags |= tfCoinBase
	}

	return entry, nil
}

func serializeUtxoEntry(entry *UtxoEntry) ([]byte, error) {
	// Spent outputs have no serialization.
	if entry.IsSpent() {
		return nil, nil
	}

	// Encode the header code.
	headerCode, err := utxoEntryHeaderCode(entry)
	if err != nil {
		return nil, err
	}

	// Calculate the size needed to serialize the entry.
	size := serialization.SerializeSizeVLQ(headerCode) + hash.HashSize + UtxoEntryAmountCoinIDSize +
		compressedTxOutSize(uint64(entry.Amount().Value), entry.PkScript())

	// Serialize the header code followed by the compressed unspent
	// transaction output.
	serialized := make([]byte, size)
	offset := serialization.PutVLQ(serialized, headerCode)
	copy(serialized[offset:offset+hash.HashSize], entry.blockHash.Bytes())
	offset += hash.HashSize
	// add Amount coinId
	byteOrder.PutUint16(serialized[offset:], uint16(entry.Amount().Id))
	offset += SpentTxoutAmountCoinIDSize
	putCompressedTxOut(serialized[offset:], uint64(entry.Amount().Value),
		entry.PkScript())

	return serialized, nil
}

func outpointKey(outpoint types.TxOutPoint) *[]byte {
	// A VLQ employs an MSB encoding, so they are useful not only to reduce
	// the amount of storage space, but also so iteration of utxos when
	// doing byte-wise comparisons will produce them in order.
	key := outpointKeyPool.Get().(*[]byte)
	idx := uint64(outpoint.OutIndex)
	*key = (*key)[:hash.HashSize+serialization.SerializeSizeVLQ(idx)]
	copy(*key, outpoint.Hash[:])
	serialization.PutVLQ((*key)[hash.HashSize:], idx)
	return key
}

var outpointKeyPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, hash.HashSize+serialization.MaxUint32VLQSerializeSize)
		return &b // Pointer to slice to avoid boxing alloc.
	},
}

func recycleOutpointKey(key *[]byte) {
	outpointKeyPool.Put(key)
}

func utxoEntryHeaderCode(entry *UtxoEntry) (uint64, error) {
	if entry.IsSpent() {
		return 0, AssertError("attempt to serialize spent utxo header")
	}

	// As described in the serialization format comments, the header code
	// encodes the height shifted over one bit and the coinbase flag in the
	// lowest bit.
	headerCode := uint64(0)
	if entry.IsCoinBase() {
		headerCode |= 0x01
	}

	return headerCode, nil
}
