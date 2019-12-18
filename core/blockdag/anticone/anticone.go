package anticone

import (
	"log"
	"math"
	"os"
)

var myLog = log.New(os.Stdout, "", 0)

// Calculate  AntiCone size, which means when some miner has just created a block,
// how many blocks at most are there created by other miners.
// Simply understanding, it is the block creating concurrency
//
// delay: Max propagation delay, unit is second
// rate: Block rate, unit is blocks/second
// security: Security level, the probability of an honest block being marked red
func GetSize(_delay, _rate, _security float64) int {
	expect := 2 * _delay * _rate
	if expect > 999 {
		myLog.Fatalf("keep expect:%v = 2 * _delay:%v * _rate:%v under 1000!", expect, _delay, _rate)
	}

	coef := 1 / math.Pow(math.E, expect)

	myLog.Printf("_delay:%v _rate:%v  _security:%v expect:%v coef:%v\n\n", _delay, _rate, _security, expect, coef)

	end := 1000

	sum := 1.0

	for k := 0; k < end; k++ {
		part := 1.0
		for j := 1; j <= k; j++ {
			part *= expect / float64(j)
		}
		sum -= part * coef
		myLog.Printf("k=%v sum=%v", k, sum)
		if sum < _security {
			return k
		}
	}
	return 0

}
