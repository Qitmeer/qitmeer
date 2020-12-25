/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package client

type notificationState struct {
	notifyBlocks   bool
	notifyReceived map[string]struct{}
}
