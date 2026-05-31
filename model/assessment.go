package model

const (
	StatusRejected          = 1
	StatusPartiallyEligible = 2
	StatusEligible          = 3
)

type Offer struct {
	Method         string  `json:"method"`
	EligibleAmount float64 `json:"eligibleAmount"`
	ROI            float64 `json:"roi"`
	Tenure         int     `json:"tenure"`
	Selected       bool    `json:"selected"`
	FOIR           float64 `json:"foir,omitempty"`
	Multiplier     float64 `json:"multiplier,omitempty"`
}

type AssessmentRequest struct {
	FullName            string `json:"fullName"`
	Age                 int    `json:"age"`
	CompanyName         string `json:"companyName"`
	MonthlySalary       int    `json:"monthlySalary"`
	ExistingEmi         int    `json:"existingEmi"`
	RequestedLoanAmount int    `json:"requestedLoanAmount"`
}

type AssessmentResponse struct {
	Status                int                    `json:"status"`
	CompanyCategory       string                 `json:"companyCategory"`
	FOIR                  float64                `json:"foir"`
	Multiplier            float64                `json:"multiplier"`
	RecommendedTenure     int                    `json:"recommendedTenure"`
	RequestedAmount       float64                `json:"requestedAmount"`
	FinalEligibility      float64                `json:"finalEligibility"`
	MaxLoanAmount         float64                `json:"maxLoanAmount"`
	IsCapped              bool                   `json:"isCapped"`
	Offers                []Offer                `json:"offers"`
	RequiredSalary        float64                `json:"requiredSalary"`
	RequiredEmiReduction  float64                `json:"requiredEmiReduction"`
	AIReadySummary        map[string]interface{} `json:"aiReadySummary"`
	AIUnderwritingSummary string                 `json:"aiUnderwritingSummary"`
	Message               string                 `json:"message"`
}
