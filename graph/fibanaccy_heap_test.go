package graph

import (
	"testing"
)

type testVal struct {
	int
}

func TestAddMin(t *testing.T) {
	fh := NewHeap()
	fh.Add(4.5, &testVal{123})
	fh.Add(5, &testVal{13})
	fh.Add(1, &testVal{12})
	fh.Add(10, &testVal{22})
	fh.Add(12, &testVal{16})
	fh.Add(14, &testVal{2})

	v, c := fh.Min()
	if v != 1 {
		t.Error("Min value returned:", v, "but 1 expected")
	}
	v, c = fh.Min()
	if v != 4.5 {
		t.Error("Min value returned:", v, "but 4.5 expected")
	}
	v, c = fh.Min()
	if v != 5 {
		t.Error("Min value returned:", v, "but 5 expected")
	}
	v, c = fh.Min()
	if v != 10 {
		t.Error("Min value returned:", v, "but 10 expected")
	}
	v, c = fh.Min()
	if v != 12 {
		t.Error("Min value returned:", v, "but 12 expected")
	}
	v, c = fh.Min()
	if v != 14 {
		t.Error("Min value returned:", v, "but 14 expected")
	}
	v, c = fh.Min()
	if c != nil {
		t.Error("Min value returned:", c, "but <nil> expected")
	}
}

func TestDecreaseScore(t *testing.T) {
	fh := NewHeap()
	fh.Add(0.5, &testVal{123})
	fh.Add(1, &testVal{123})
	fh.Add(5, &testVal{13})
	fh.Add(7, &testVal{12})
	fh.Add(9, &testVal{22})
	fh.Add(10, &testVal{16})
	twelve := &testVal{2}
	fh.Add(12, twelve)
	fifteen := &testVal{9}
	fh.Add(15, fifteen)
	fh.Add(17, &testVal{13})
	fh.Add(19, &testVal{442})
	fh.Add(22, &testVal{55})
	fh.Add(27, &testVal{26})
	fh.Add(28, &testVal{26})
	fh.Add(30, &testVal{26})
	fh.Add(31, &testVal{26})
	fh.Add(34, &testVal{26})
	fh.Add(36, &testVal{26})
	fh.Add(37, &testVal{26})
	fh.Add(38, &testVal{26})

	fh.Min()

	fh.DecreaseScore(2, twelve)
	fh.DecreaseScore(4, fifteen)
	fh.DecreaseScore(1, fifteen)

	expectedMins := []float64{1, 1, 2, 5, 7, 9, 10, 17, 19, 22, 27, 28, 30, 31, 34, 36, 37, 38}
	for _, v := range expectedMins {
		gv, _ := fh.Min()
		if v != gv {
			t.Error("Expected result after decrease keys:", v, " but:", gv, "obtained")
		}
	}
}
