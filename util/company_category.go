package util

import "strings"

// GetCompanyCategory resolves a company name to a hardcoded category for the MVP.
func GetCompanyCategory(companyName string) string {
	normalized := strings.ToUpper(strings.TrimSpace(companyName))

	switch normalized {
	case "INDIAN OIL":
		return "CAT_SA"
	case "TCS", "INFOSYS", "HDFC BANK", "ICICI BANK":
		return "CAT_A"
	case "WIPRO", "TECH MAHINDRA":
		return "CAT_B"
	default:
		return "CAT_C"
	}
}
