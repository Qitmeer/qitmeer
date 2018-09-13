package addmgr

import "github.com/noxproject/nox/log"

type AddrManager struct {}
func (* AddrManager) Start() {
	log.Info("AddrManager started")
}
func (* AddrManager) Stop() {
	log.Info("AddrManager stopped")
}
