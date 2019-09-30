package peerserver

import (
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/p2p/addmgr"
	"strconv"
	"time"
)

func (s *PeerServer) upnpUpdateThread() {
	// Go off immediately to prevent code duplication, thereafter we renew
	// lease every 15 minutes.
	timer := time.NewTimer(0 * time.Second)
	lport, _ := strconv.ParseInt(s.chainParams.DefaultPort, 10, 16)
	first := true
out:
	for {
		select {
		case <-timer.C:
			// TODO: pick external port  more cleverly
			// TODO: know which ports we are listening to on an external net.
			// TODO: if specific listen port doesn't work then ask for wildcard
			// listen port?
			// XXX this assumes timeout is in seconds.
			listenPort, err := s.nat.AddPortMapping("tcp", int(lport), int(lport),
				"dcrd listen port", 20*60)
			if err != nil {
				log.Warn("can't add UPnP port mapping", "error",err)
			}
			if first && err == nil {
				// TODO: look this up periodically to see if upnp domain changed
				// and so did ip.
				externalip, err := s.nat.GetExternalAddress()
				if err != nil {
					log.Warn("UPnP can't get external address","error", err)
					continue out
				}
				na := types.NewNetAddressIPPort(externalip, uint16(listenPort),
					s.services)
				err = s.addrManager.AddLocalAddress(na, addmgr.UpnpPrio)
				if err != nil {
					log.Warn("Failed to add UPnP local address %s: %v",
						na.IP.String(), err)
				} else {
					log.Warn("Successfully bound via UPnP","addr",addmgr.NetAddressKey(na))
					first = false
				}
			}
			timer.Reset(time.Minute * 15)
		case <-s.quit:
			break out
		}
	}

	timer.Stop()

	if err := s.nat.DeletePortMapping("tcp", int(lport), int(lport)); err != nil {
		log.Warn("unable to remove UPnP port mapping","error", err)
	} else {
		log.Debug("successfully disestablished UPnP port mapping")
	}

	s.wg.Done()
}
