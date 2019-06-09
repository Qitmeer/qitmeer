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
// level: Security level, the probability of an honest block being marked red
func GetSize(_delay, _rate, _security float64) int {

	factor := 2 * _delay * _rate
	if factor > 10000 {
		myLog.Fatalf("keep factor:%v = 2 * _delay:%v * _rate:%v under 1000!", factor, _delay, _rate)
	}

	coef := math.Pow(math.E, factor)

	//myLog.Printf("_delay:%v _rate:%v  level:%v factor:%v coef:%v\n\n", _delay, _rate, _security, factor, coef)

	sum := 0.0

	outLen := 10
	kQueue := make([]float64, 0)

	end := 1000

	k := -1

	for kk := 1; kk < end; kk++ {
		sum = coef

		sigma := 1.0
		for j := 1; j <= kk; j++ {
			n := 1.0
			for jj := 1; jj <= j; jj++ {
				n *= factor / float64(jj)
			}
			sigma += n
		}
		sum -= sigma
		sum /= coef

		if k < 0 {
			if sum < _security {
				for i := 0; i < len(kQueue); i++ {
					leftBound := outLen
					if kk <= leftBound {
						leftBound = kk - 1
					}
					//myLog.Printf("kk=%v sum=%v", kk-(leftBound-i), kQueue[i])
				}
				//myLog.Printf("\n[MIN]kk=%v sum=%v\n\n", kk, sum)
				k = kk
				end = kk + outLen + 1
			}

			kQueue = append(kQueue, sum)
			if kk > outLen {
				kQueue = kQueue[1:]
			}

		} else {
			//myLog.Printf("kk=%v sum=%v", kk, sum)
		}
	}

	return k
}
