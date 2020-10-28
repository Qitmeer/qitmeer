/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package event

type Event struct {
	Data interface{}
	Ack  chan<- struct{}
}

func New(data interface{}) *Event {
	return &Event{Data: data, Ack: nil}
}
