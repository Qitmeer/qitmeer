package anticone

import (
	"log"
	"testing"
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

	GetSize(15, 1.0, 0.1)

}
