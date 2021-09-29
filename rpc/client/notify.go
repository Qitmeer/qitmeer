package client

import (
	"errors"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
)

var (
	// ErrWebsocketsRequired is an error to describe the condition where the
	// caller is trying to use a websocket-only feature, such as requesting
	// notifications or other websocket requests when the client is
	// configured to run in HTTP POST mode.
	ErrWebsocketsRequired = errors.New("a websocket connection is required " +
		"to use this feature")
)

type FutureNotifyBlocksResult chan *response

func (r FutureNotifyBlocksResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

func (c *Client) NotifyBlocksAsync() FutureNotifyBlocksResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := cmds.NewNotifyBlocksCmd()
	return c.sendCmd(cmd)
}

func (c *Client) StopNotifyBlocksAsync() FutureNotifyBlocksResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := cmds.NewStopNotifyBlocksCmd()
	return c.sendCmd(cmd)
}

func (c *Client) NotifyBlocks() error {
	return c.NotifyBlocksAsync().Receive()
}

func (c *Client) StopNotifyBlocks() error {
	return c.StopNotifyBlocksAsync().Receive()
}

func (c *Client) NotifyTxsByAddrAsync(reload bool, addr []string, outpoint []cmds.OutPoint) FutureNotifyBlocksResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := cmds.NewNotifyTxsByAddrCmd(reload, addr, outpoint)
	return c.sendCmd(cmd)
}

func (c *Client) NotifyTxsByAddr(reload bool, addr []string, outpoint []cmds.OutPoint) error {
	return c.NotifyTxsByAddrAsync(reload, addr, outpoint).Receive()
}

func (c *Client) StopNotifyTxsByAddrAsync(addr []string) FutureNotifyBlocksResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := cmds.StopNotifyTxsByAddrCmd(addr)
	return c.sendCmd(cmd)
}

func (c *Client) StopNotifyTxsByAddr(addr []string) error {
	return c.StopNotifyTxsByAddrAsync(addr).Receive()
}

func (c *Client) RescanAsync(beginBlock, endBlock uint64, addrs []string, op []cmds.OutPoint) FutureNotifyBlocksResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := cmds.NotifyRescanCmd(beginBlock, endBlock, addrs, op)
	return c.sendCmd(cmd)
}

func (c *Client) Rescan(beginBlock, endBlock uint64, addrs []string, op []cmds.OutPoint) error {
	return c.RescanAsync(beginBlock, endBlock, addrs, op).Receive()
}

func (c *Client) NotifyTxsConfirmedAsync(txs []cmds.TxConfirm) FutureNotifyBlocksResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := cmds.NewNotifyTxsConfirmed(txs)
	return c.sendCmd(cmd)
}

func (c *Client) RemoveTxsConfirmedAsync(txs []cmds.TxConfirm) FutureNotifyBlocksResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := cmds.NewRemoveTxsConfirmed(txs)
	return c.sendCmd(cmd)
}

func (c *Client) NotifyTxsConfirmed(txs []cmds.TxConfirm) error {
	return c.NotifyTxsConfirmedAsync(txs).Receive()
}

func (c *Client) RemoveTxsConfirmed(txs []cmds.TxConfirm) error {
	return c.RemoveTxsConfirmedAsync(txs).Receive()
}

type FutureNotifyReceivedResult chan *response

func (r FutureNotifyReceivedResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

func (c *Client) notifyReceivedInternal(addresses []string) FutureNotifyReceivedResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	// Convert addresses to strings.
	cmd := cmds.NewNotifyReceivedCmd(addresses)
	return c.sendCmd(cmd)
}

func (c *Client) NotifyReceivedAsync(addresses []types.Address) FutureNotifyReceivedResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	// Convert addresses to strings.
	addrs := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		addrs = append(addrs, addr.String())
	}
	cmd := cmds.NewNotifyReceivedCmd(addrs)
	return c.sendCmd(cmd)
}

func (c *Client) NotifyReceived(addresses []types.Address) error {
	return c.NotifyReceivedAsync(addresses).Receive()
}

type FutureNotifyNewTransactionsResult chan *response

func (r FutureNotifyNewTransactionsResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

func (c *Client) NotifyNewTransactionsAsync(verbose bool) FutureNotifyNewTransactionsResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := cmds.NewNotifyNewTransactionsCmd(verbose)
	return c.sendCmd(cmd)
}

func (c *Client) NotifyNewTransactions(verbose bool) error {
	return c.NotifyNewTransactionsAsync(verbose).Receive()
}

type FutureStopNotifyNewTransactionsResult chan *response

func (r FutureStopNotifyNewTransactionsResult) Receive() error {
	_, err := receiveFuture(r)
	return err
}

func (c *Client) StopNotifyNewTransactionsAsync() FutureStopNotifyNewTransactionsResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	// Ignore the notification if the client is not interested in
	// notifications.
	if c.ntfnHandlers == nil {
		return newNilFutureResult()
	}

	cmd := cmds.NewStopNotifyNewTransactionsCmd()
	return c.sendCmd(cmd)
}

func (c *Client) StopNotifyNewTransactions() error {
	return c.StopNotifyNewTransactionsAsync().Receive()
}
