package service

import (
	"testing"

	"loanpilot-backend/model"
)

func TestAssessEligibleCase(t *testing.T) {
	svc := NewAssessmentService()

	resp := svc.Assess(model.AssessmentRequest{
		FullName:            "Eligible User",
		Age:                 30,
		CompanyName:         "TCS",
		MonthlySalary:       80000,
		ExistingEmi:         5000,
		RequestedLoanAmount: 1000000,
	})

	if resp.Status != model.StatusEligible {
		t.Fatalf("expected status %d, got %d", model.StatusEligible, resp.Status)
	}
	if resp.RequiredSalary != 0 {
		t.Fatalf("expected required salary 0, got %.0f", resp.RequiredSalary)
	}
	if resp.RequiredEmiReduction != 0 {
		t.Fatalf("expected required EMI reduction 0, got %.0f", resp.RequiredEmiReduction)
	}
	if resp.FinalEligibility != 2100000 {
		t.Fatalf("expected final eligibility 2100000, got %.0f", resp.FinalEligibility)
	}
	if len(resp.Offers) != 2 {
		t.Fatalf("expected 2 offers, got %d", len(resp.Offers))
	}
	if !resp.Offers[1].Selected {
		t.Fatalf("expected multiplier offer to be selected")
	}
}

func TestAssessPartiallyEligibleCase(t *testing.T) {
	svc := NewAssessmentService()

	resp := svc.Assess(model.AssessmentRequest{
		FullName:            "Partial User",
		Age:                 26,
		CompanyName:         "TCS",
		MonthlySalary:       60000,
		ExistingEmi:         5000,
		RequestedLoanAmount: 1500000,
	})

	if resp.Status != model.StatusPartiallyEligible {
		t.Fatalf("expected status %d, got %d", model.StatusPartiallyEligible, resp.Status)
	}
	if resp.FinalEligibility != 1430000 {
		t.Fatalf("expected final eligibility 1430000, got %.0f", resp.FinalEligibility)
	}
	if resp.RequiredSalary != 63000 {
		t.Fatalf("expected required salary 63000, got %.0f", resp.RequiredSalary)
	}
	if resp.RequiredEmiReduction != 2700 {
		t.Fatalf("expected required EMI reduction 2700, got %.0f", resp.RequiredEmiReduction)
	}
	if resp.IsCapped {
		t.Fatalf("expected offer to be uncapped")
	}
}

func TestAssessRejectedCase(t *testing.T) {
	svc := NewAssessmentService()

	resp := svc.Assess(model.AssessmentRequest{
		FullName:            "Rejected User",
		Age:                 20,
		CompanyName:         "TCS",
		MonthlySalary:       60000,
		ExistingEmi:         5000,
		RequestedLoanAmount: 1000000,
	})

	if resp.Status != model.StatusRejected {
		t.Fatalf("expected status %d, got %d", model.StatusRejected, resp.Status)
	}
	if resp.RequiredSalary != 0 {
		t.Fatalf("expected required salary 0, got %.0f", resp.RequiredSalary)
	}
	if resp.RequiredEmiReduction != 0 {
		t.Fatalf("expected required EMI reduction 0, got %.0f", resp.RequiredEmiReduction)
	}
	if len(resp.Offers) != 0 {
		t.Fatalf("expected no offers for rejected case, got %d", len(resp.Offers))
	}
}
