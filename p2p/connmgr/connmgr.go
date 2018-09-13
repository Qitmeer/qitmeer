package connmgr

import "github.com/noxproject/nox/log"

type ConnManager struct {}
func (* ConnManager) Start() {
	log.Info("ConnManager started")
}
func (* ConnManager) Stop() {
	log.Info("ConnManager stopped")
}


