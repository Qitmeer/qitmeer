package blkmgr

import (
	"github.com/Qitmeer/qng-core/core/types"
)

// getCurrentTemplateMsg handles a request for the current mining block template.
type getCurrentTemplateMsg struct {
	reply chan getCurrentTemplateResponse
}

// getCurrentTemplateResponse is a response sent to the reply channel of a
// getCurrentTemplateMsg.
type getCurrentTemplateResponse struct {
	Template *types.BlockTemplate
}

// setCurrentTemplateMsg handles a request to change the current mining block
// template.
type setCurrentTemplateMsg struct {
	Template *types.BlockTemplate
	reply    chan setCurrentTemplateResponse
}

// setCurrentTemplateResponse is a response sent to the reply channel of a
// setCurrentTemplateMsg.
type setCurrentTemplateResponse struct {
}

// getParentTemplateMsg handles a request for the current parent mining block
// template.
type getParentTemplateMsg struct {
	reply chan getParentTemplateResponse
}

// getParentTemplateResponse is a response sent to the reply channel of a
// getParentTemplateMsg.
type getParentTemplateResponse struct {
	Template *types.BlockTemplate
}

// setParentTemplateMsg handles a request to change the parent mining block
// template.
type setParentTemplateMsg struct {
	Template *types.BlockTemplate
	reply    chan setParentTemplateResponse
}

// setParentTemplateResponse is a response sent to the reply channel of a
// setParentTemplateMsg.
type setParentTemplateResponse struct {
}

// GetCurrentTemplate gets the current block template for mining.
func (b *BlockManager) GetCurrentTemplate() *types.BlockTemplate {
	reply := make(chan getCurrentTemplateResponse)
	b.msgChan <- getCurrentTemplateMsg{reply: reply}
	response := <-reply
	return response.Template
}

// SetCurrentTemplate sets the current block template for mining.
func (b *BlockManager) SetCurrentTemplate(bt *types.BlockTemplate) {
	reply := make(chan setCurrentTemplateResponse)
	b.msgChan <- setCurrentTemplateMsg{Template: bt, reply: reply}
	<-reply
}

// GetParentTemplate gets the current parent block template for mining.
func (b *BlockManager) GetParentTemplate() *types.BlockTemplate {
	reply := make(chan getParentTemplateResponse)
	b.msgChan <- getParentTemplateMsg{reply: reply}
	response := <-reply
	return response.Template
}

// SetParentTemplate sets the current parent block template for mining.
func (b *BlockManager) SetParentTemplate(bt *types.BlockTemplate) {
	reply := make(chan setParentTemplateResponse)
	b.msgChan <- setParentTemplateMsg{Template: bt, reply: reply}
	<-reply
}

// deepCopyBlockTemplate returns a deeply copied block template that copies all
// data except a block's references to transactions, which are kept as pointers
// in the block. This is considered safe because transaction data is generally
// immutable, with the exception of coinbases which we alternatively also
// deep copy.
func deepCopyBlockTemplate(blockTemplate *types.BlockTemplate) *types.BlockTemplate {
	if blockTemplate == nil {
		return nil
	}

	// Deep copy the header, which we hash on.
	headerCopy := blockTemplate.Block.Header

	// Copy transactions pointers. Duplicate the coinbase
	// transaction, because it might update it by modifying
	// the extra nonce.
	transactionsCopy := make([]*types.Transaction, len(blockTemplate.Block.Transactions))
	coinbaseCopy :=
		types.NewTxDeep(blockTemplate.Block.Transactions[0])
	for i, mtx := range blockTemplate.Block.Transactions {
		if i == 0 {
			transactionsCopy[i] = coinbaseCopy.Transaction()
		} else {
			transactionsCopy[i] = mtx
		}
	}

	msgBlockCopy := &types.Block{
		Header:       headerCopy,
		Transactions: transactionsCopy,
	}

	fees := make([]int64, len(blockTemplate.Fees))
	copy(fees, blockTemplate.Fees)

	sigOps := make([]int64, len(blockTemplate.SigOpCounts))
	copy(sigOps, blockTemplate.SigOpCounts)

	return &types.BlockTemplate{
		Block:           msgBlockCopy,
		Fees:            fees,
		SigOpCounts:     sigOps,
		Height:          blockTemplate.Height,
		Blues:           blockTemplate.Blues,
		ValidPayAddress: blockTemplate.ValidPayAddress,
	}
}
