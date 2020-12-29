package acct

import (
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/rpc"
	"github.com/Qitmeer/qitmeer/rpc/client/cmds"
)

// account manager communicate with various backends for signing transactions.
type AccountManager struct {
}

func (a *AccountManager) Start() error {
	log.Debug("Starting account manager")
	return nil
}

func (a *AccountManager) Stop() error {
	log.Debug("Stopping account manager")
	return nil
}

func (a AccountManager) APIs() []rpc.API {
	return []rpc.API{
		{
			NameSpace: cmds.DefaultServiceNameSpace,
			Service:   NewPublicAccountManagerAPI(&a),
			Public:    true,
		},
	}
}

func New() (*AccountManager, error) {
	a := AccountManager{}
	return &a, nil
}
