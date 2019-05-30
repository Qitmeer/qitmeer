package blockchain

import (
	"qitmeer/common/hash"
	"qitmeer/core/blockdag"
)

// BlockLocator is used to help locate a specific block.  The algorithm for
// building the block locator is to add the hashes in reverse order until
// the genesis block is reached.  In order to keep the list of locator hashes
// to a reasonable number of entries, first the most recent previous 12 block
// hashes are added, then the step is doubled each loop iteration to
// exponentially decrease the number of hashes as a function of the distance
// from the block being located.
//
// For example, assume a block chain with a side chain as depicted below:
// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
// 	                              \-> 16a -> 17a
//
// The block locator for block 17a would be the hashes of blocks:
// [17a 16a 15 14 13 12 11 10 9 8 7 6 4 genesis]
type BlockLocator []*hash.Hash

// log2FloorMasks defines the masks to use when quickly calculating
// floor(log2(x)) in a constant log2(32) = 5 steps, where x is a uint32, using
// shifts.  They are derived from (2^(2^x) - 1) * (2^(2^x)), for x in 4..0.
var log2FloorMasks = []uint32{0xffff0000, 0xff00, 0xf0, 0xc, 0x2}

// fastLog2Floor calculates and returns floor(log2(x)) in a constant 5 steps.
func fastLog2Floor(n uint32) uint8 {
	rv := uint8(0)
	exponent := uint8(16)
	for i := 0; i < 5; i++ {
		if n&log2FloorMasks[i] != 0 {
			rv += exponent
			n >>= exponent
		}
		exponent >>= 1
	}
	return rv
}

// LatestBlockLocator returns a block locator for the latest DAG state.
//
// This function is safe for concurrent access.
func (b *BlockChain) LatestBlockLocator() (BlockLocator, error) {
	b.chainLock.RLock()
	locator := b.blockLocator(nil)
	b.chainLock.RUnlock()
	return locator, nil
}

// LocateBlocks returns the hashes of the blocks after the first known block in
// the locator until the provided stop hash is reached, or up to the provided
// max number of block hashes.
//
// In addition, there are two special cases:
//
// - When no locators are provided, the stop hash is treated as a request for
//   that block, so it will either return the stop hash itself if it is known,
//   or nil if it is unknown
// - When locators are provided, but none of them are known, hashes starting
//   after the genesis block will be returned
//
// This function is safe for concurrent access.
func (b *BlockChain) LocateBlocks(gs *blockdag.GraphState, maxHashes uint32) []*hash.Hash {
	b.chainLock.RLock()
	hashes := b.bd.LocateBlocks(gs, uint(maxHashes))
	b.chainLock.RUnlock()
	return hashes
}

// locateBlocks returns the hashes of the blocks after the first known block in
// the locator until the provided stop hash is nearby, or up to the provided
// max number of block hashes.
//
// See the comment on the exported function for more details on special cases.
//
// This function MUST be called with the chain state lock held (for reads).
func (b *BlockChain) locateBlocks(locator BlockLocator, hashStop *hash.Hash, maxHashes uint32) []hash.Hash {
	// It must be not empty
	loLen:=len(locator)
	if loLen==0 {
		return nil
	}
	hashes:=[]hash.Hash{}
	endHash:=hashStop
	if hashStop.IsEqual(&hash.ZeroHash) {
		// If the stop block is zero, that means it doesn't end until last tip.
		endHash=b.bd.GetLastBlock().GetHash()
	}else if hashStop.IsEqual(locator[0]) {
		// In this case, we're going back to what block we need.
		for _,v:=range locator{
			if !b.index.HaveBlock(v) {
				continue
			}
			hashes=append(hashes,hash.Hash(*v))
		}
		return hashes
	}
	if !b.bd.HasBlock(endHash) {
		return nil
	}
	endBlock:=b.bd.GetBlock(endHash)
	hashesSet:=blockdag.NewHashSet()

	// First of all, we need to make sure we have the parents of block.
	hashesSet.AddSet(endBlock.GetParents())
	curNum:=uint32(hashesSet.Len())

	// Because of chain forking, a common forking point must be found.
	// It's the real starting point.
	var curBlock *blockdag.Block
	for i:=0;i<loLen;i++{
		if b.bd.HasBlock(locator[i]) {
			curBlock=b.bd.GetBlock(locator[0])
			break
		}
	}

	for curBlock!=nil {
		curBlockH:=b.bd.GetBlockByOrder(curBlock.GetOrder()+1)
		if curBlockH==nil {
			break
		}
		curBlock=b.bd.GetBlock(curBlockH)
		hashesSet.Add(curBlock.GetHash())
		curNum++

		if curNum>=maxHashes||
			curBlock==endBlock||
			curBlock.GetOrder()>=endBlock.GetOrder() {
			break
		}
	}

	for k:=range hashesSet.GetMap(){
		hashes=append(hashes,hash.Hash(k))
	}
	return hashes
}

// BlockLocatorFromHash returns a block locator for the passed block hash.
// See BlockLocator for details on the algorithm used to create a block locator.
//
// In addition to the general algorithm referenced above, this function will
// return the block locator for the latest DAG.
//
// This function is safe for concurrent access.
func (b *BlockChain) BlockLocatorFromHash(hash *hash.Hash) BlockLocator {
	b.chainLock.RLock()
	node := b.index.LookupNode(hash)
	locator := b.blockLocator(node)
	b.chainLock.RUnlock()
	return locator
}

// blockLocator returns a block locator for the passed block node.  The passed
// node can be nil in which case the block locator for the current DAG
// associated with the view will be returned.
// This function MUST be called with the view mutex locked (for reads).
func (b *BlockChain) blockLocator(node *blockNode) BlockLocator {
	// Use the current tip if requested.
	if node == nil {
		lb:=b.bd.GetLastBlock()
		node = b.index.lookupNode(lb.GetHash())
		if node == nil {
			return nil
		}
	}

	// Calculate the max number of entries that will ultimately be in the
	// block locator.  See the description of the algorithm for how these
	// numbers are derived.
	var maxEntries uint8
	if node.order <= 12 {
		maxEntries = uint8(node.order) + 1
	} else {
		// Requested hash itself + previous 10 entries + genesis block.
		// Then floor(log2(height-10)) entries for the skip portion.
		adjustedHeight := uint32(node.order) - 10
		maxEntries = 12 + fastLog2Floor(adjustedHeight)
	}
	locator := make(BlockLocator, 0, maxEntries)

	step := uint64(1)
	for node != nil {
		locator = append(locator, &node.hash)

		// Nothing more to add once the genesis block has been added.
		if node.order == 0 {
			break
		}

		// Calculate height of previous node to include ensuring the
		// final node is the genesis block.
		// height := node.height - step
		// if height < 0 {
		//	 height = 0
		// }
		height := uint64(0)
		if node.order > step {
			height = node.order - step
		}

		nodeH := b.bd.GetBlockByOrder(uint(height))
		node = b.index.lookupNode(nodeH)
		// Once 11 entries have been included, start doubling the
		// distance between included hashes.
		if len(locator) > 10 {
			step *= 2
		}
	}

	return locator
}

