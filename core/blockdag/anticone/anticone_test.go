package anticone

import (
	"fmt"
	"github.com/Qitmeer/qng-core/params"
	"log"
	"testing"
	"time"
)

func TestSigma(t *testing.T) {
	k := 1
	factor := 0.9
	sum := 0.0
	for j := k + 1; j < k+100; j++ {

		x := 1.0
		for jj := 1; jj <= j; jj++ {
			x *= factor / float64(j)
		}

		sum += x
	}
	log.Println(sum)
}

func TestAntiCone(t *testing.T) {
	result := []int{12, 8, 6, 5, 4, 4, 4, 3, 3, 3, 3, 3, 3, 2, 2, 2, 2, 2, 2}
	index := 0
	for i := 5; i < 100; i += 5 {
		rate := 1.0 / float64(i)
		if GetSize(BlockDelay, rate, SecurityLevel) != result[index] {
			t.Fatal()
		}
		index++
	}
}

func TestShowParamsAntiCone(t *testing.T) {
	rate := 1.0 / float64(params.TestNetParams.TargetTimePerBlock/time.Second)
	fmt.Printf("testnet:%d\n", GetSize(BlockDelay, rate, SecurityLevel))

	rate = 1.0 / float64(params.MainNetParams.TargetTimePerBlock/time.Second)
	fmt.Printf("mainnet:%d\n", GetSize(BlockDelay, rate, SecurityLevel))

	rate = 1.0 / float64(params.MixNetParams.TargetTimePerBlock/time.Second)
	fmt.Printf("mixnet:%d\n", GetSize(BlockDelay, rate, SecurityLevel))

	rate = 1.0 / float64(params.PrivNetParams.TargetTimePerBlock/time.Second)
	fmt.Printf("privnet:%d\n", GetSize(BlockDelay, rate, SecurityLevel))
}
