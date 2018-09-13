package er

import (
	"fmt"
	"errors"
	"github.com/noxproject/nox/common/hash"
)

// rpcNoTxInfoError is a convenience function for returning a nicely formatted
// RPC error which indicates there is no information available for the provided
// transaction hash.
func RpcNoTxInfoError(txHash *hash.Hash) error {
	return errors.New(
		fmt.Sprintf("No information available about transaction %v",
			txHash))
}

func RpcInternalError(err, context string) error{
	return errors.New(
		fmt.Sprintf("%s : %s",context,err))
}
