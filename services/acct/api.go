package acct

// PublicEthereumAPI provides an API to access Ethereum full node-related
// information.
type PublicAccountManagerAPI struct {
	a *AccountManager
}

// NewPublicEthereumAPI creates a new Ethereum protocol API for full nodes.
func NewPublicAccountManagerAPI(a *AccountManager) *PublicAccountManagerAPI {
	return &PublicAccountManagerAPI{a}
}

func (api *PublicAccountManagerAPI) GetBalance() int32 {
	return 0
}
