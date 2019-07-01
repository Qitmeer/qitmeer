package blockdag

import (
	"fmt"
	"testing"
	"gonum.org/v1/gonum/floats"
)

func TestOnlineRiskInSpectre(t *testing.T) {
	if floats.EqualWithinAbs(GetRisk(300,0.1,10,5,10,30),0.1509544, tol) {
		t.FailNow()
	}
}

func TestOnlineRiskInSpectreData(t *testing.T) {
	maxL:=1
	for i:=0;i<maxL ;i++  {
		risk:=GetRisk(500,0.1,666,0.05,600,666)

		fmt.Println(risk)
	}

}