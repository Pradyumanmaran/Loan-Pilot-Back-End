package util

import "math"

// RoundDownToThousand truncates a value to the nearest lower thousand.
func RoundDownToThousand(value float64) float64 {
	if value <= 0 {
		return 0
	}

	return math.Floor(value/1000) * 1000
}

// RoundUpToThousand increases a value to the nearest higher thousand.
func RoundUpToThousand(value float64) float64 {
	if value <= 0 {
		return 0
	}

	return math.Ceil(value/1000) * 1000
}
