/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package client

type notificationState struct {
	notifyBlocks   bool
	notifyReceived map[string]struct{}
}

func (s *notificationState) Copy() *notificationState {
	var stateCopy notificationState
	stateCopy.notifyBlocks = s.notifyBlocks
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
