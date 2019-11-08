package blockdag

import (
	"fmt"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
	"math"
)

const tol = 1e-10

// parameters
// N: just pick a value much greater than 1.
// alpha: the attacker’s relative computational power.
// lambda: blocks per second.
// delay: the upper bound on the recent delay diameter in the network.
// waitingTime: wait time.
// antiPast: min(|future(x')|), where x' is x or any block in anticone(x)
// and x is the block we want to confirm, ideally this should be
// about waitingTime * lambda.
func GetRisk(N int, alpha float64, lambda float64, delay float64, waitingTime uint, antiPast int) float64 {
	if N < 3 || antiPast <= 0 {
		return 0
	}
	delta := alpha * lambda * delay

	tMatData := make([]float64, N*N)

	for i := 1; i < N-1; i++ {
		tMatData[i*N+i-1] = 1 - alpha
		tMatData[i*N+i+1] = alpha
	}
	tMatData[(N-1)*N+N-2] = 1 - alpha
	tMatData[(N-1)*N+N-1] = alpha
	tMatData[0] = (1 - alpha) * math.Exp(-delta)
	tMatData[1] = math.Exp(-delta) * (alpha + delta)

	p := distuv.Poisson{Lambda: delta}
	for i := 2; i < N-1; i++ {
		tMatData[i] = p.Prob(float64(i))
	}
	tMatData[N-1] = 1 - p.CDF(float64(N-2))

	tMat := mat.NewDense(N, N, tMatData)

	var eig mat.Eigen
	ok := eig.Factorize(tMat, mat.EigenLeft)
	if !ok {
		fmt.Println("Eigendecomposition failed")
	}

	ceigenvalues := eig.Values(nil)
	cfeatures := complex(1, 0)
	featuresIndex := -1

	for i := 0; i < len(ceigenvalues); i++ {
		if !floats.EqualWithinAbs(imag(ceigenvalues[i]), imag(cfeatures), tol) {
			continue
		}
		if !floats.EqualWithinAbs(real(ceigenvalues[i]), real(cfeatures), tol) {
			continue
		}
		featuresIndex = i
		break
	}
	if featuresIndex == -1 {
		fmt.Println("eigen vector failed")
		return 0
	}
	ceigenvectors := eig.LeftVectorsTo(nil)
	r, _ := ceigenvectors.Dims()
	vecData := []float64{}
	var vecMod float64
	for i := 0; i < r; i++ {
		realData := real(ceigenvectors.At(i, featuresIndex))
		vecData = append(vecData, realData)
		vecMod += realData
	}
	vecRMod := 1 / vecMod
	vect := mat.NewVecDense(r, vecData)
	vect.ScaleVec(vecRMod, vect)

	a := (float64(waitingTime) + 2*delay) * alpha * lambda
	pa := distuv.Poisson{Lambda: a}
	qa := alpha / (1 - alpha)
	riskHidden := 0.0
	for i := 0; i < N; i++ {
		sum_m := 0.0
		mg := antiPast - i - 1
		mj := int(math.Max(float64(mg), 0))
		for j := 0; j <= int(mj); j++ {
			sum_m += pa.Prob(float64(j)) * math.Pow(qa, math.Max(float64(mg-j), 0))
		}
		sum_m += 1 - pa.CDF(float64(mj))
		riskHidden += vect.AtVec(i) * sum_m
	}
	return riskHidden
}
