package util

import "math"

const EmiPerLakh = 2200

// CalculateEligibility derives the final loan eligibility and rounds down to the nearest thousand.
func CalculateEligibility(salary float64, existingEmi float64, foir float64, multiplier float64) float64 {
	_, _, finalEligibility := CalculateOfferEligibility(salary, existingEmi, foir, multiplier)
	return finalEligibility
}

// CalculateOfferEligibility returns the FOIR offer, multiplier offer, and final eligibility.
func CalculateOfferEligibility(salary float64, existingEmi float64, foir float64, multiplier float64) (float64, float64, float64) {
	multiplierEligibility := RoundDownToThousand((salary - existingEmi) * multiplier)

	eligibleEmi := salary*(foir/100) - existingEmi
	foirEligibility := RoundDownToThousand((eligibleEmi / EmiPerLakh) * 100000)

	finalEligibility := math.Min(foirEligibility, multiplierEligibility)
	if finalEligibility < 0 {
		finalEligibility = 0
	}

	return foirEligibility, multiplierEligibility, finalEligibility
}
