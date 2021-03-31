/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package client

type notificationState struct {
	notifyBlocks       bool
	notifyNewTx        bool
	notifyNewTxVerbose bool
	notifyReceived     map[string]struct{}
}

func (s *notificationState) Copy() *notificationState {
	var stateCopy notificationState
	stateCopy.notifyBlocks = s.notifyBlocks
	stateCopy.notifyNewTx = s.notifyNewTx
	stateCopy.notifyNewTxVerbose = s.notifyNewTxVerbose
	stateCopy.notifyReceived = make(map[string]struct{})
	for addr := range s.notifyReceived {
		stateCopy.notifyReceived[addr] = struct{}{}
	}
	return &stateCopy
}

func newNotificationState() *notificationState {
	return &notificationState{
		notifyReceived: make(map[string]struct{}),
	}
}
