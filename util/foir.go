package util

// GetFOIR returns the FOIR percentage for a company category and salary band.
func GetFOIR(category string, salary float64) float64 {
	switch category {
	case "CAT_SA", "CAT_A", "CAT_B":
		switch {
		case salary < 50000:
			return 60
		case salary < 75000:
			return 65
		default:
			return 70
		}
	case "CAT_C", "CAT_D":
		switch {
		case salary < 30000:
			return 50
		case salary < 50000:
			return 60
		case salary < 75000:
			return 65
		default:
			return 70
		}
	default:
		switch {
		case salary < 30000:
			return 50
		case salary < 50000:
			return 60
		case salary < 75000:
			return 65
		default:
			return 70
		}
	}
}
