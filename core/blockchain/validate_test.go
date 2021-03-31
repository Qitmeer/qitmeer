package blockchain

import (
	"bytes"
	"encoding/hex"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/params"
	"testing"
)

func Test_CheckTransactionSanity(t *testing.T) {
	txStr := "0100000001df44f16415ecc836e1c8e7f8cd3af896dc1c81f0f4e159535e744a0f75c76b7501000000ffffffff0101000093459e0b0000001976a9149a2aff5fcbfc29f9a30262bd6ba976f08f9fcf9188ac00000000000000004c790a60016a4730440220336757a0c9bdcb2510fe627520ad79694fc36d1880bf24ecc7496a65fc65a991022073853aba8fa9f3f29a8f1135f3b570aa3f804887cdbfa260d1544a0ea8c9aa530121036f2b25ef58a673b255d63802d1a9f25ffeb6f2a78c0eab33ee17831e07e5e487"
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
		Amount:   types.Amount{Value: 99999999999999999, Id: types.MEERID},
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
