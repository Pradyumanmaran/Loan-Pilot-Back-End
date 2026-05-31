package service

import (
	"math"

	"loanpilot-backend/model"
	"loanpilot-backend/util"
)

func NewAssessmentService() *AssessmentService {
	return &AssessmentService{ai: NewAIService()}
}

func NewAssessmentServiceWithAI(ai AIService) *AssessmentService {
	if ai == nil {
		ai = NewAIService()
	}

	return &AssessmentService{ai: ai}
}

type AssessmentService struct {
	ai AIService
}

func (s *AssessmentService) Assess(req model.AssessmentRequest) model.AssessmentResponse {
	requestedAmount := float64(req.RequestedLoanAmount)

	if req.Age < 21 {
		return s.rejectedResponse(requestedAmount, "Rejected: age must be at least 21")
	}

	if req.Age >= 60 {
		return s.rejectedResponse(requestedAmount, "Rejected: age must be below 60")
	}

	if req.MonthlySalary < 18000 {
		return s.rejectedResponse(requestedAmount, "Rejected: monthly salary must be at least 18000")
	}

	companyCategory := util.GetCompanyCategory(req.CompanyName)
	recommendedTenure := util.DeriveTenure(req.Age)
	foir := util.GetFOIR(companyCategory, float64(req.MonthlySalary))
	multiplier := util.GetMultiplier(companyCategory, float64(req.MonthlySalary), recommendedTenure)
	roi := util.GetROI(companyCategory)

	foirOfferAmount, multiplierOfferAmount, rawFinalEligibility := s.calculateEligibilityForSalaryAndEmi(
		float64(req.MonthlySalary),
		float64(req.ExistingEmi),
		foir,
		multiplier,
	)

	offers := s.buildOffers(foirOfferAmount, multiplierOfferAmount, roi, recommendedTenure, foir, multiplier)
	selectedMethod := s.selectedOfferMethod(offers)
	isCapped := rawFinalEligibility > util.MaxLoanAmount
	finalEligibility := rawFinalEligibility
	if isCapped {
		finalEligibility = util.MaxLoanAmount
	}

	requiredSalary := 0.0
	requiredEmiReduction := 0.0
	status := model.StatusEligible
	message := "Passed underwriting assessment"

	if finalEligibility < util.MinLoanAmount {
		status = model.StatusRejected
		message = "Rejected: eligibility below minimum loan amount"
	} else if requestedAmount > finalEligibility {
		status = model.StatusPartiallyEligible
		message = "Requested amount exceeds final eligibility"
		requiredSalary = s.calculateRequiredSalary(
			float64(req.MonthlySalary),
			float64(req.ExistingEmi),
			requestedAmount,
			companyCategory,
			recommendedTenure,
		)
		requiredEmiReduction = s.calculateRequiredEmiReduction(
			float64(req.MonthlySalary),
			float64(req.ExistingEmi),
			requestedAmount,
			companyCategory,
			recommendedTenure,
		)
	}

	if isCapped && status != model.StatusRejected {
		message = "Passed underwriting assessment. Offer capped at maximum product limit."
	}

	aiSummaryRequest := AISummaryRequest{
		Status:               status,
		CompanyCategory:      companyCategory,
		RequestedAmount:      requestedAmount,
		FinalEligibility:     finalEligibility,
		RequiredSalary:       requiredSalary,
		RequiredEmiReduction: requiredEmiReduction,
		SelectedMethod:       selectedMethod,
		IsCapped:             isCapped,
		MaxLoanAmount:        util.MaxLoanAmount,
	}

	aiUnderwritingSummary, aiErr := s.ai.GenerateUnderwritingSummary(aiSummaryRequest)
	if aiErr != nil && aiUnderwritingSummary == "" {
		aiUnderwritingSummary = fallbackUnderwritingSummary(aiSummaryRequest)
	}

	aiSummary := map[string]interface{}{
		"status":               status,
		"companyCategory":      companyCategory,
		"requestedAmount":      requestedAmount,
		"finalEligibility":     finalEligibility,
		"maxLoanAmount":        util.MaxLoanAmount,
		"isCapped":             isCapped,
		"requiredSalary":       requiredSalary,
		"requiredEmiReduction": requiredEmiReduction,
		"selectedMethod":       selectedMethod,
	}

	return model.AssessmentResponse{
		Status:                status,
		CompanyCategory:       companyCategory,
		FOIR:                  foir,
		Multiplier:            multiplier,
		RecommendedTenure:     recommendedTenure,
		RequestedAmount:       requestedAmount,
		FinalEligibility:      finalEligibility,
		MaxLoanAmount:         util.MaxLoanAmount,
		IsCapped:              isCapped,
		Offers:                offers,
		RequiredSalary:        requiredSalary,
		RequiredEmiReduction:  requiredEmiReduction,
		AIReadySummary:        aiSummary,
		AIUnderwritingSummary: aiUnderwritingSummary,
		Message:               message,
	}
}

func (s *AssessmentService) rejectedResponse(requestedAmount float64, message string) model.AssessmentResponse {
	return model.AssessmentResponse{
		Status:               model.StatusRejected,
		CompanyCategory:      "",
		FOIR:                 0,
		Multiplier:           0,
		RecommendedTenure:    0,
		RequestedAmount:      requestedAmount,
		FinalEligibility:     0,
		MaxLoanAmount:        util.MaxLoanAmount,
		IsCapped:             false,
		Offers:               []model.Offer{},
		RequiredSalary:       0,
		RequiredEmiReduction: 0,
		AIReadySummary: map[string]interface{}{
			"status":               model.StatusRejected,
			"companyCategory":      "",
			"requestedAmount":      requestedAmount,
			"finalEligibility":     0,
			"maxLoanAmount":        util.MaxLoanAmount,
			"isCapped":             false,
			"requiredSalary":       0,
			"requiredEmiReduction": 0,
			"selectedMethod":       "",
		},
		AIUnderwritingSummary: "",
		Message:               message,
	}
}

func (s *AssessmentService) buildOffers(
	foirAmount float64,
	multiplierAmount float64,
	roi float64,
	tenure int,
	foir float64,
	multiplier float64,
) []model.Offer {
	foirOffer := model.Offer{
		Method:         "FOIR",
		EligibleAmount: foirAmount,
		ROI:            roi,
		Tenure:         tenure,
		Selected:       foirAmount <= multiplierAmount,
		FOIR:           foir,
	}

	multiplierOffer := model.Offer{
		Method:         "MULTIPLIER",
		EligibleAmount: multiplierAmount,
		ROI:            roi,
		Tenure:         tenure,
		Selected:       multiplierAmount < foirAmount,
		Multiplier:     multiplier,
	}

	return []model.Offer{foirOffer, multiplierOffer}
}

func (s *AssessmentService) selectedOfferMethod(offers []model.Offer) string {
	for _, offer := range offers {
		if offer.Selected {
			return offer.Method
		}
	}

	return ""
}

func (s *AssessmentService) calculateEligibilityForSalaryAndEmi(
	salary float64,
	existingEmi float64,
	foir float64,
	multiplier float64,
) (float64, float64, float64) {
	foirOfferAmount, multiplierOfferAmount, finalEligibility := util.CalculateOfferEligibility(salary, existingEmi, foir, multiplier)
	return foirOfferAmount, multiplierOfferAmount, finalEligibility
}

func (s *AssessmentService) calculateRequiredSalary(
	currentSalary float64,
	existingEmi float64,
	targetAmount float64,
	category string,
	tenure int,
) float64 {
	target := math.Min(targetAmount, util.MaxLoanAmount)
	if s.finalEligibilityAtSalary(currentSalary, existingEmi, category, tenure) >= target {
		return 0
	}

	low := int(math.Ceil(currentSalary / 1000))
	if low < 18 {
		low = 18
	}
	high := low
	for s.finalEligibilityAtSalary(float64(high*1000), existingEmi, category, tenure) < target {
		low = high
		high *= 2
		if high > 1000000 {
			break
		}
	}

	if s.finalEligibilityAtSalary(float64(high*1000), existingEmi, category, tenure) < target {
		return float64(high * 1000)
	}

	for high-low > 1 {
		mid := low + (high-low)/2
		if s.finalEligibilityAtSalary(float64(mid*1000), existingEmi, category, tenure) >= target {
			high = mid
		} else {
			low = mid
		}
	}

	return float64(high * 1000)
}

func (s *AssessmentService) calculateRequiredEmiReduction(
	salary float64,
	currentEmi float64,
	targetAmount float64,
	category string,
	tenure int,
) float64 {
	target := math.Min(targetAmount, util.MaxLoanAmount)
	if s.finalEligibilityAtEmi(salary, currentEmi, 0, category, tenure) >= target {
		return 0
	}

	maxReduction := int(math.Floor(currentEmi / 100))
	if maxReduction < 0 {
		maxReduction = 0
	}

	if s.finalEligibilityAtEmi(salary, currentEmi, float64(maxReduction*100), category, tenure) < target {
		return float64(maxReduction * 100)
	}

	low := 0
	high := maxReduction
	for high-low > 1 {
		mid := low + (high-low)/2
		if s.finalEligibilityAtEmi(salary, currentEmi, float64(mid*100), category, tenure) >= target {
			high = mid
		} else {
			low = mid
		}
	}

	return float64(high * 100)
}

func (s *AssessmentService) finalEligibilityAtSalary(
	salary float64,
	existingEmi float64,
	category string,
	tenure int,
) float64 {
	foir := util.GetFOIR(category, salary)
	multiplier := util.GetMultiplier(category, salary, tenure)
	_, _, finalEligibility := util.CalculateOfferEligibility(salary, existingEmi, foir, multiplier)
	if finalEligibility > util.MaxLoanAmount {
		finalEligibility = util.MaxLoanAmount
	}
	return finalEligibility
}

func (s *AssessmentService) finalEligibilityAtEmi(
	salary float64,
	currentEmi float64,
	reduction float64,
	category string,
	tenure int,
) float64 {
	foir := util.GetFOIR(category, salary)
	multiplier := util.GetMultiplier(category, salary, tenure)
	_, _, finalEligibility := util.CalculateOfferEligibility(salary, currentEmi-reduction, foir, multiplier)
	if finalEligibility > util.MaxLoanAmount {
		finalEligibility = util.MaxLoanAmount
	}
	return finalEligibility
}
