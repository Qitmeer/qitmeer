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
	err = checkTransactionSanityForAllNet(&tx)
	if err != nil {
		t.Fatal(err)
	}

	// We create an attack transacton data
	attackerPkScript, err := hex.DecodeString("76a914c0f0b73c320e1fe38eb1166a57b953e509c8f93e88ac")
	if err != nil {
		panic(err)
	}
	tx.AddTxOut(&types.TxOutput{
		Amount:   999999999999,
		PkScript: attackerPkScript,
	})

	err = checkTransactionSanityForAllNet(&tx)
	if err == nil {
		t.Fatal("Successful attack")
	}
}

func checkTransactionSanityForAllNet(tx *types.Transaction) error {
	err := CheckTransactionSanity(tx, &params.TestNetParams)
	if err != nil {
		return err
	}

	err = CheckTransactionSanity(tx, &params.PrivNetParams)
	if err != nil {
		return err
	}

	err = CheckTransactionSanity(tx, &params.MixNetParams)
	if err != nil {
		return err
	}

	err = CheckTransactionSanity(tx, &params.MainNetParams)
	if err != nil {
		return err
	}
	return nil
}