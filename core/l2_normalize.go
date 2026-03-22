package core

import "math"

func L2Normalize(values []float32) []float32 {
	var sum float64
	for _, value := range values {
		sum += float64(value * value)
	}

	if sum == 0 {
		out := make([]float32, len(values))
		copy(out, values)
		return out
	}

	norm := float32(math.Sqrt(sum))
	out := make([]float32, len(values))
	for i, value := range values {
		out[i] = value / norm
	}

	return out
}
