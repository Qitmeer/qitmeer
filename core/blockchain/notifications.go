// Copyright (c) 2017-2018 The nox developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"

	"qitmeer/core/types"
	"qitmeer/common/hash"
)

// NotificationType represents the type of a notification message.
type NotificationType int

// NotificationCallback is used for a caller to provide a callback for
// notifications about various chain events.
type NotificationCallback func(*Notification)

// Constants for the type of a notification message.
const (
	// BlockAccepted indicates the associated block was accepted into
	// the block chain.  Note that this does not necessarily mean it was
	// added to the main chain.  For that, use NTBlockConnected.
	BlockAccepted NotificationType = iota

	// BlockConnected indicates the associated block was connected to the
	// main chain.
	BlockConnected

	// BlockDisconnected indicates the associated block was disconnected
	// from the main chain.
	BlockDisconnected

	// Reorganization indicates that a blockchain reorganization is in
	// progress.
	Reorganization

)

// notificationTypeStrings is a map of notification types back to their constant
// names for pretty printing.
var notificationTypeStrings = map[NotificationType]string{
	BlockAccepted:         "BlockAccepted",
	BlockConnected:        "BlockConnected",
	BlockDisconnected:     "BlockDisconnected",
	Reorganization:        "Reorganization",
}

// String returns the NotificationType in human-readable form.
func (n NotificationType) String() string {
	if s, ok := notificationTypeStrings[n]; ok {
		return s
	}
	return fmt.Sprintf("Unknown Notification Type (%d)", int(n))
}

// BlockAcceptedNotifyData is the structure for data indicating information
// about an accepted block.  Note that this does not necessarily mean the block
// that was accepted extended the best chain as it might have created or
// extended a side chain.
type BlockAcceptedNotifyData struct {
	// BestHeight is the height of the current best chain.  Since the accepted
	// block might be on a side chain, this is not necessarily the same as the
	// height of the accepted block.
	BestHeight uint64

	// ForkLen is the length of the side chain the block extended or zero in the
	// case the block extended the main chain.
	//
	// This can be used in conjunction with the height of the accepted block to
	// determine the height at which the side chain the block created or
	// extended forked from the best chain.
	ForkLen int64

	// Block is the block that was accepted into the chain.
	Block *types.SerializedBlock
}

// ReorganizationNotifyData is the structure for data indicating information
// about a reorganization.
type ReorganizationNotifyData struct {
	OldHash   hash.Hash
	OldHeight uint64
	NewHash   hash.Hash
	NewHeight uint64
}

// Notification defines notification that is sent to the caller via the callback
// function provided during the call to New and consists of a notification type
// as well as associated data that depends on the type as follows:
// 	- BlockAccepted:         *BlockAcceptedNotifyData
// 	- BlockConnected:        []*types.Block of len 2
// 	- BlockDisconnected:     []*types.Block of len 2
//  - Reorganization:        *ReorganizationNotifyData

type Notification struct {
	Type NotificationType
	Data interface{}
}

// sendNotification sends a notification with the passed type and data if the
// caller requested notifications by providing a callback function in the call
// to New.
func (b *BlockChain) sendNotification(typ NotificationType, data interface{}) {
	// Ignore it if the caller didn't request notifications.
	if b.notifications == nil {
		return
	}

	// Generate and send the notification.
	n := Notification{Type: typ, Data: data}
	log.Trace("send blkmgr notification", "type",n.Type, "data",n.Data)
	b.notifications(&n)
}
