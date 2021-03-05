package rpc

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/blockchain"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/database"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
	"math"
	"time"
)

// MaxInvPerMsg is the maximum number of inventory vectors that can be in a
// single inv message.
const MaxInvPerMsg = 50000

// ErrRescanReorg defines the error that is returned when an unrecoverable
// reorganize is detected during a rescan.
var ErrRescanReorg = cmds.RPCError{
	Code:    cmds.ErrRPCDatabase,
	Message: "Reorganize",
}

// recoverFromReorg attempts to recover from a detected reorganize during a
// rescan.  It fetches a new range of block shas from the database and
// verifies that the new range of blocks is on the same fork as a previous
// range of blocks.  If this condition does not hold true, the JSON-RPC error
// for an unrecoverable reorganize is returned.
func recoverFromReorg(chain *blockchain.BlockChain, minBlock, maxBlock uint64,
	lastBlock *hash.Hash) ([]hash.Hash, error) {

	hashList, err := chain.OrderRange(minBlock, maxBlock)
	if err != nil {
		log.Error(fmt.Sprintf("Error looking up block range: %v", err))
		return nil, &cmds.RPCError{
			Code:    cmds.ErrRPCDatabase,
			Message: "Database error: " + err.Error(),
		}
	}
	if lastBlock == nil || len(hashList) == 0 {
		return hashList, nil
	}

	blk, err := chain.BlockByHash(&hashList[0])
	if err != nil {
		log.Error(fmt.Sprintf("Error looking up possibly reorged block: %v",
			err))
		return nil, &cmds.RPCError{
			Code:    cmds.ErrRPCDatabase,
			Message: "Database error: " + err.Error(),
		}
	}
	jsonErr := descendantBlock(lastBlock, blk)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return hashList, nil
}

// descendantBlock returns the appropriate JSON-RPC error if a current block
// fetched during a reorganize is not a direct child of the parent block hash.
func descendantBlock(prevHash *hash.Hash, curBlock *types.SerializedBlock) error {
	curHash := &curBlock.Block().Header.ParentRoot
	if !prevHash.IsEqual(curHash) {
		log.Error(fmt.Sprintf("Stopping rescan for reorged block %v "+
			"(replaced by block %v)", prevHash, curHash))
		return &ErrRescanReorg
	}
	return nil
}

// scanBlockChunks executes a rescan in chunked stages. We do this to limit the
// amount of memory that we'll allocate to a given rescan. Every so often,
// we'll send back a rescan progress notification to the websockets client. The
// final block and block hash that we've scanned will be returned.
func scanBlockChunks(wsc *wsClient, cmd *cmds.RescanCmd, lookups *rescanKeys, minBlock,
	maxBlock uint64, chain *blockchain.BlockChain) (
	*types.SerializedBlock, *hash.Hash, error) {

	// lastBlock and lastBlockHash track the previously-rescanned block.
	// They equal nil when no previous blocks have been rescanned.
	var (
		lastBlock     *types.SerializedBlock
		lastBlockHash *hash.Hash
	)

	// A ticker is created to wait at least 10 seconds before notifying the
	// websocket client of the current progress completed by the rescan.
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Instead of fetching all block shas at once, fetch in smaller chunks
	// to ensure large rescans consume a limited amount of memory.
fetchRange:
	for minBlock < maxBlock {
		// Limit the max number of hashes to fetch at once to the
		// maximum number of items allowed in a single inventory.
		// This value could be higher since it's not creating inventory
		// messages, but this mirrors the limiting logic used in the
		// peer-to-peer protocol.
		maxLoopBlock := maxBlock
		if maxLoopBlock-minBlock > MaxInvPerMsg {
			maxLoopBlock = minBlock + MaxInvPerMsg
		}
		hashList, err := chain.OrderRange(minBlock, maxLoopBlock)
		if err != nil {
			log.Error(fmt.Sprintf("Error looking up block range: %v", err))
			return nil, nil, &cmds.RPCError{
				Code:    cmds.ErrRPCDatabase,
				Message: "Database error: " + err.Error(),
			}
		}
		if len(hashList) == 0 {
			// The rescan is finished if no blocks hashes for this
			// range were successfully fetched and a stop block
			// was provided.
			if maxBlock != math.MaxInt64 {
				break
			}

			// If the rescan is through the current block, set up
			// the client to continue to receive notifications
			// regarding all rescanned addresses and the current set
			// of unspent outputs.
			//
			// This is done safely by temporarily grabbing exclusive
			// access of the block manager.  If no more blocks have
			// been attached between this pause and the fetch above,
			// then it is safe to register the websocket client for
			// continuous notifications if necessary.  Otherwise,
			// continue the fetch loop again to rescan the new
			// blocks (or error due to an irrecoverable reorganize).
			best := wsc.server.BC.BestSnapshot()
			curHash := &best.Hash
			again := true
			if lastBlockHash == nil || *lastBlockHash == *curHash {
				again = false
			}
			if err != nil {
				log.Error(fmt.Sprintf("Error fetching best block "+
					"hash: %v", err))
				return nil, nil, &cmds.RPCError{
					Code: cmds.ErrRPCDatabase,
					Message: "Database error: " +
						err.Error(),
				}
			}
			if again {
				continue
			}
			break
		}

	loopHashList:
		for i := range hashList {
			blk, err := chain.BlockByHash(&hashList[i])
			if err != nil {
				// Only handle reorgs if a block could not be
				// found for the hash.
				if dbErr, ok := err.(database.Error); !ok ||
					dbErr.ErrorCode != database.ErrBlockNotFound {

					log.Error(fmt.Sprintf("Error looking up "+
						"block: %v", err))
					return nil, nil, &cmds.RPCError{
						Code: cmds.ErrRPCDatabase,
						Message: "Database error: " +
							err.Error(),
					}
				}
				h := blk.Block().BlockHash()
				node := wsc.server.BC.BlockIndex().LookupNode(&h)
				if node == nil {
					return nil, nil, &cmds.RPCError{
						Code:    cmds.ErrInvalidNode,
						Message: "no node ",
					}
				}
				// Update the source block order
				blk.SetOrder(node.GetOrder())
				blk.SetHeight(node.GetHeight())
				// If an absolute max block was specified, don't
				// attempt to handle the reorg.
				if maxBlock != math.MaxInt64 {
					log.Error(fmt.Sprintf("Stopping rescan for "+
						"reorged block %v",
						cmd.EndBlock))
					return nil, nil, &ErrRescanReorg
				}

				// If the lookup for the previously valid block
				// hash failed, there may have been a reorg.
				// Fetch a new range of block hashes and verify
				// that the previously processed block (if there
				// was any) still exists in the database.  If it
				// doesn't, we error.
				//
				// A goto is used to branch executation back to
				// before the range was evaluated, as it must be
				// reevaluated for the new hashList.
				minBlock += uint64(i)
				hashList, err = recoverFromReorg(
					chain, minBlock, maxBlock, lastBlockHash,
				)
				if err != nil {
					return nil, nil, err
				}
				if len(hashList) == 0 {
					break fetchRange
				}
				goto loopHashList
			}
			if i == 0 && lastBlockHash != nil {
				// Ensure the new hashList is on the same fork
				// as the last block from the old hashList.
				jsonErr := descendantBlock(lastBlockHash, blk)
				if jsonErr != nil {
					return nil, nil, jsonErr
				}
			}

			// A select statement is used to stop rescans if the
			// client requesting the rescan has disconnected.
			select {
			case <-wsc.quit:
				log.Debug(fmt.Sprintf("Stopped rescan at order %v "+
					"for disconnected client", blk.Order()))
				return nil, nil, nil
			default:
				rescanBlock(wsc, lookups, blk, wsc.server.BC)
				lastBlock = blk
				lastBlockHash = blk.Hash()
			}

			// Periodically notify the client of the progress
			// completed.  Continue with next block if no progress
			// notification is needed yet.
			select {
			case <-ticker.C: // fallthrough
			default:
				continue
			}

			n := cmds.NewRescanProgressNtfn(
				hashList[i].String(), blk.Order(),
				blk.Block().Header.Timestamp.Unix(),
			)
			mn, err := cmds.MarshalCmd(nil, n)
			if err != nil {
				log.Error(fmt.Sprintf("Failed to marshal rescan "+
					"progress notification: %v", err))
				continue
			}

			if err = wsc.QueueNotification(mn); err == ErrClientQuit {
				// Finished if the client disconnected.
				log.Debug(fmt.Sprintf("Stopped rescan at order %v "+
					"for disconnected client", blk.Order()))
				return nil, nil, nil
			}
		}

		minBlock += uint64(len(hashList))
	}

	return lastBlock, lastBlockHash, nil
}

// rescanBlock rescans all transactions in a single block.  This is a helper
// function for handleRescan.
func rescanBlock(wsc *wsClient, lookups *rescanKeys, blk *types.SerializedBlock, bc *blockchain.BlockChain) {
	for _, tx := range blk.Transactions() {
		// notifySpend is a closure we'll use when we first detect that
		// a transactions spends an outpoint/script in our filter list.
		notifyTx := func() error {
			wsc.server.ntfnMgr.NotifyBlockTx(tx, blk)
			return nil
		}
		needNotifyTx := false
		// We'll start by iterating over the transaction's inputs to
		// determine if it spends an outpoint/script in our filter list.
		for _, txin := range tx.Tx.TxIn {
			// We'll also recompute the pkScript the input is
			// attempting to spend to determine whether it is
			// relevant to us.
			pkScript, err := txscript.ComputePkScript(
				txin.SignScript, nil,
			)
			if err != nil {
				continue
			}
			addr, err := pkScript.Address(wsc.server.params)
			if err != nil {
				continue
			}

			// If it is, we'll also dispatch a spend notification
			// for this transaction if we haven't already.
			if _, ok := lookups.addrs[addr.String()]; ok {
				needNotifyTx = true
			}
		}

		for _, txout := range tx.Tx.TxOut {
			_, addrs, _, _ := txscript.ExtractPkScriptAddrs(
				txout.PkScript, wsc.server.params)

			for _, addr := range addrs {
				if _, ok := lookups.addrs[addr.String()]; !ok {
					continue
				}
				needNotifyTx = true
			}
		}
		if needNotifyTx {
			if err := notifyTx(); err != nil {
				// Stop the rescan early if the websocket client
				// disconnected.
				if err == ErrClientQuit {
					return
				} else {
					log.Error(fmt.Sprintf("Unable to notify "+
						"redeeming transaction %v: %v",
						tx.Hash(), err))
					continue
				}
			}
		}
	}
}