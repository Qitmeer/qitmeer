/*
 * Copyright (c) 2017-2020 The qitmeer developers
 */

package synch

import (
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
)

var defaultReadDuration = TtfbTimeout
var defaultWriteDuration = 10 * time.Second

// SetRPCStreamDeadlines sets read and write deadlines for libp2p-based connection streams.
func SetRPCStreamDeadlines(stream network.Stream) {
	SetStreamReadDeadline(stream, defaultReadDuration)
	SetStreamWriteDeadline(stream, defaultWriteDuration)
}

// SetStreamReadDeadline for reading from libp2p connection streams, deciding when to close
// a connection based on a particular duration.
func SetStreamReadDeadline(stream network.Stream, duration time.Duration) {
	if err := stream.SetReadDeadline(time.Now().Add(duration)); err != nil {
		log.Debug(fmt.Sprintf("Failed to set stream deadline:%v peer:%s protocol:%s direction:%s",
			err, stream.Conn().RemotePeer(), stream.Protocol(), stream.Stat().Direction))
	}
}

// SetStreamWriteDeadline for writing to libp2p connection streams, deciding when to close
// a connection based on a particular duration.
func SetStreamWriteDeadline(stream network.Stream, duration time.Duration) {
	if err := stream.SetWriteDeadline(time.Now().Add(duration)); err != nil {
		log.Debug(fmt.Sprintf("Failed to set stream deadline:%v peer:%s protocol:%s direction:%s",
			err, stream.Conn().RemotePeer(), stream.Protocol(), stream.Stat().Direction))
	}
}
