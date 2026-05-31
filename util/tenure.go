package util

// DeriveTenure maps age to a recommended loan tenure in months.
func DeriveTenure(age int) int {
	switch {
	case age >= 21 && age <= 55:
		return 60
	case age == 56:
		return 48
	case age == 57:
		return 36
	case age == 58:
		return 24
	case age == 59:
		return 12
	default:
		return 0
	}
}
