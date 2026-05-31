package util

// GetMultiplier returns the eligibility multiplier for a category, salary band, and tenure.
func GetMultiplier(category string, salary float64, tenure int) float64 {
	switch category {
	case "CAT_SA", "CAT_A":
		switch tenure {
		case 12:
			return 9
		case 24:
			switch {
			case salary < 50000:
				return 13
			case salary < 75000:
				return 15
			default:
				return 16
			}
		case 36:
			switch {
			case salary < 50000:
				return 20
			case salary < 75000:
				return 20
			default:
				return 23
			}
		case 48:
			switch {
			case salary < 50000:
				return 21
			case salary < 75000:
				return 23
			default:
				return 26
			}
		case 60:
			switch {
			case salary < 50000:
				return 24
			case salary < 75000:
				return 26
			default:
				return 28
			}
		default:
			return 0
		}
	case "CAT_B":
		switch tenure {
		case 12:
			return 7
		case 24:
			return 13
		case 36:
			return 18
		case 48:
			return 22
		case 60:
			return 24
		default:
			return 0
		}
	case "CAT_C", "CAT_D":
		switch tenure {
		case 12:
			return 7
		case 24:
			return 12
		case 36:
			return 15
		case 48:
			return 18
		case 60:
			return 18
		default:
			return 0
		}
	default:
		return 0
	}
}
