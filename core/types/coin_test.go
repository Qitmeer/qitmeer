package types

import "testing"

func Test_CoinConfigs(t *testing.T) {
	coinsCfg := &CoinConfigs{}

	tests := []struct {
		txFees AmountMap
		expect bool
	}{
		{
			txFees: AmountMap{
				MEERID: AtomsPerCoin,
				QITID:  AtomsPerCoin,
			},
			expect: true,
		},
		{
			txFees: AmountMap{
				MEERID: AtomsPerCoin * 2,
				QITID:  AtomsPerCoin * 3,
			},
			expect: true,
		},
		{
			txFees: AmountMap{
				MEERID: AtomsPerCoin,
				QITID:  AtomsPerCoin,
			},
			expect: true,
		},
		{
			txFees: AmountMap{
				MEERID: AtomsPerCoin,
				QITID:  AtomsPerCoin,
			},
			expect: true,
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
