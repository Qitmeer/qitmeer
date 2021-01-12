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

func (c *Client) NotifyBlocks() error {
	return c.NotifyBlocksAsync().Receive()
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
