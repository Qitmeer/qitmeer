/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package client

import (
	"encoding/json"
	"github.com/Qitmeer/qitmeer/common/hash"
	"time"
)

type NotificationHandlers struct {
	OnClientConnected     func()
	OnBlockConnected      func(hash *hash.Hash, height int32, t time.Time)
	OnBlockDisconnected   func(hash *hash.Hash, height int32, t time.Time)
	OnUnknownNotification func(method string, params []json.RawMessage)
}
