package blockchain

import (
	"bytes"
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
	"testing"
)

func Test_CheckTransactionSanity(t *testing.T) {
	txStr := "0100000001091caf9a96b172c4151dbdbf387b37dc4f9cb4ff892ed25a39319eccef23d6fdffffffffffffffff01007841cb020000001976a914c1777151516afe2b9f59bbd1479231aa2f250d2888ac000000000000000001145108e5a95f9f26f1371a092f7169746d6565722f"
	if len(txStr)%2 != 0 {
		txStr = "0" + txStr
	}
	serializedTx, err := hex.DecodeString(txStr)
	if err != nil {
		t.Fatal(err)
	}
	var tx types.Transaction
	err = tx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		t.Fatal(err)
	}

	err = CheckTransactionSanity(&tx, &params.TestNetParams)
	if err != nil {
		t.Fatal(err)
	}

	err = CheckTransactionSanity(&tx, &params.PrivNetParams)
	if err != nil {
		t.Fatal(err)
	}

	err = CheckTransactionSanity(&tx, &params.MixNetParams)
	if err != nil {
		t.Fatal(err)
	}

	err = CheckTransactionSanity(&tx, &params.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}
}
