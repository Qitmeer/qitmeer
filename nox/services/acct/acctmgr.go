package acct

import (
	"github.com/noxproject/nox/log"
	"github.com/noxproject/nox/p2p"
	"github.com/noxproject/nox/rpc"
)

// account manager communicate with various backends for signing transactions.
type AccountManager struct {

}

func (a *AccountManager) Start(server *p2p.PeerServer) error {
	log.Debug("Starting account manager")
	return nil
}

func (a *AccountManager) Stop() error {
	log.Debug("Stopping account manager")
	return nil
}

func (a AccountManager)	APIs() []rpc.API {
	return []rpc.API{
		{
			NameSpace: rpc.DefaultServiceNameSpace,
			Service:   NewPublicAccountManagerAPI(&a),
		},
	}
}

func New() (*AccountManager, error){
	a := AccountManager{}
	return &a,nil
}


