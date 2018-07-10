package blkmgr

import "github.com/noxproject/nox/core/types"

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


