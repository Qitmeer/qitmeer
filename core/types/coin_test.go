package types

import "testing"

func Test_CoinConfigs(t *testing.T) {
	coinsCfg := &CoinConfigs{
		&CoinConfig{
			Id:    METID,
			Type:  FloorFeeType,
			Value: AtomsPerCoin,
		},
		&CoinConfig{
			Id:    TERID,
			Type:  EqualFeeType,
			Value: AtomsPerCoin,
		},
	}

	tests := []struct {
		txFees AmountMap
		expect bool
	}{
		{
			txFees: AmountMap{
				MEERID: AtomsPerCoin,
				QITID:  AtomsPerCoin,
				METID:  AtomsPerCoin,
				TERID:  AtomsPerCoin,
			},
			expect: true,
		},
		{
			txFees: AmountMap{
				MEERID: AtomsPerCoin * 2,
				QITID:  AtomsPerCoin * 3,
				METID:  AtomsPerCoin,
				TERID:  AtomsPerCoin,
			},
			expect: true,
		},
		{
			txFees: AmountMap{
				MEERID: AtomsPerCoin,
				QITID:  AtomsPerCoin,
				METID:  AtomsPerCoin - 1,
				TERID:  AtomsPerCoin,
			},
			expect: false,
		},
		{
			txFees: AmountMap{
				MEERID: AtomsPerCoin,
				QITID:  AtomsPerCoin,
				METID:  AtomsPerCoin,
				TERID:  AtomsPerCoin * 2,
			},
			expect: false,
		},
	}

	for _, test := range tests {
		err := coinsCfg.CheckFees(test.txFees)
		if (err == nil && test.expect) ||
			(err != nil && !test.expect) {
			continue
		}
		t.Fatalf("txFees:%v Expect:%v", test.txFees, test.expect)
	}
}
