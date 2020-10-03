// Copyright (c) 2017-2018 The qitmeer developers
// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/common/roughtime"
	"github.com/Qitmeer/qitmeer/core/protocol"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"
	mrand "math/rand"
	"net"
	"strconv"
	"time"
)

const (
	// These constants are used by the DNS seed code to pick a random last
	// seen time.
	secondsIn3Days int32 = 24 * 60 * 60 * 3
	secondsIn4Days int32 = 24 * 60 * 60 * 4
)

// OnSeed is the signature of the callback function which is invoked when DNS
// seeding is succesfull.
type OnSeed func(addrs []*types.NetAddress)

// LookupFunc is the signature of the DNS lookup function.
type LookupFunc func(string) ([]net.IP, error)

// SeedFromDNS uses DNS seeding to populate the address manager with peers.
func SeedFromDNS(chainParams *params.Params, reqServices protocol.ServiceFlag, lookupFn LookupFunc, seedFn OnSeed) {
	for _, dnsseed := range chainParams.DNSSeeds {
		var host string
		if !dnsseed.HasFiltering || reqServices == protocol.Full {
			host = dnsseed.Host
		} else {
			host = fmt.Sprintf("x%x.%s", uint64(reqServices), dnsseed.Host)
		}

		go func(host string) {
			randSource := mrand.New(mrand.NewSource(roughtime.Now().UnixNano()))

			seedpeers, err := lookupFn(host)
			if err != nil {
				log.Warn("DNS discovery failed", "seed", host, "error", err)
				return
			}
			numPeers := len(seedpeers)

			log.Info(fmt.Sprintf("%d addresses found from DNS seed %s", numPeers, host))

			if numPeers == 0 {
				return
			}
			addresses := make([]*types.NetAddress, len(seedpeers))
			// if this errors then we have *real* problems
			intPort, _ := strconv.ParseUint(chainParams.DefaultPort,10,16)
			for i, peer := range seedpeers {
				addresses[i] = types.NewNetAddressTimestamp(
					// bitcoind seeds with addresses from
					// a time randomly selected between 3
					// and 7 days ago.
					roughtime.Now().Add(-1*time.Second*time.Duration(secondsIn3Days+
						randSource.Int31n(secondsIn4Days))),
					0, peer, uint16(intPort))
			}

			seedFn(addresses)
		}(host)
	}
}
