package core

import (
	"math"
	"testing"
)

func TestL2Normalize(t *testing.T) {
	values := []float32{3, 4}
	got := L2Normalize(values)

	if math.Abs(float64(got[0]-0.6)) > 1e-6 {
		t.Fatalf("expected first value to be 0.6, got %f", got[0])
	}
	if math.Abs(float64(got[1]-0.8)) > 1e-6 {
		t.Fatalf("expected second value to be 0.8, got %f", got[1])
	}
}
