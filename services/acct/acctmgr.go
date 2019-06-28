package acct

import (
	"github.com/HalalChain/qitmeer-lib/log"
	"github.com/HalalChain/qitmeer-lib/rpc"
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



