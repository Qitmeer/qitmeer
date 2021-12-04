// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"github.com/Qitmeer/qng-core/common/hash"
	"github.com/Qitmeer/qng-core/core/event"
	"github.com/Qitmeer/qng-core/core/types"
)

// NotificationType represents the type of a notification message.
type NotificationType int

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
	BlockAccepted:     "BlockAccepted",
	BlockConnected:    "BlockConnected",
	BlockDisconnected: "BlockDisconnected",
	Reorganization:    "Reorganization",
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
	IsMainChainTipChange bool

	// Block is the block that was accepted into the chain.
	Block *types.SerializedBlock

	Flags BehaviorFlags
}

// ReorganizationNotifyData is the structure for data indicating information
// about a reorganization.
type ReorganizationNotifyData struct {
	OldBlocks []*hash.Hash
	NewBlock  *hash.Hash
	NewOrder  uint64
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
	if b.events == nil {
		return
	}

	// Generate and send the notification.
	n := &Notification{Type: typ, Data: data}
	b.CacheNotifications = append(b.CacheNotifications, n)
}

func (b *BlockChain) flushNotifications() {
	if len(b.CacheNotifications) <= 0 {
		return
	}
	for _, n := range b.CacheNotifications {
		log.Trace("send blkmgr notification", "type", n.Type, "data", n.Data)
		b.events.Send(event.New(n))
	}
	b.CacheNotifications = []*Notification{}
}
