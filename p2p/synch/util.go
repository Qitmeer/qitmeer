package synch

import (
	"fmt"
	libp2pcore "github.com/libp2p/go-libp2p-core"
)

func closeSteam(stream libp2pcore.Stream) {
	if err := stream.Close(); err != nil {
		log.Error(fmt.Sprintf("Failed to close stream:%v", err))
	}
}
