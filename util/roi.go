package util

// GetROI returns the product ROI by company category for MVP underwriting offers.
func GetROI(category string) float64 {
	switch category {
	case "CAT_SA", "CAT_A":
		return 11.49
	case "CAT_B":
		return 12.49
	case "CAT_C":
		return 13.49
	case "CAT_D":
		return 15.49
	default:
		return 13.49
	}
}
