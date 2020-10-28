/*
 * Copyright (c) 2020.
 * Project:qitmeer
 * File:addrfactory.go
 * Date:7/15/20 9:34 AM
 * Author:Jin
 * Email:lochjin@gmail.com
 */

package p2p

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/config"
	ma "github.com/multiformats/go-multiaddr"
)

// withRelayAddrs returns an AddrFactory which will return Multiaddr via
// specified relay string in addition to existing MultiAddr.
func withRelayAddrs(relay string) config.AddrsFactory {
	return func(addrs []ma.Multiaddr) []ma.Multiaddr {
		if relay == "" {
			return addrs
		}

		var relayAddrs []ma.Multiaddr

		for _, a := range addrs {
			if strings.Contains(a.String(), "/p2p-circuit") {
				continue
			}
			relayAddr, err := ma.NewMultiaddr(relay + "/p2p-circuit" + a.String())
			if err != nil {
				log.Error(fmt.Sprintf("Failed to create multiaddress for relay node: %v", err))
			} else {
				relayAddrs = append(relayAddrs, relayAddr)
			}
		}

		if len(relayAddrs) == 0 {
			log.Warn("Addresses via relay node are zero - using non-relay addresses")
			return addrs
		}
		return append(addrs, relayAddrs...)
	}
}
