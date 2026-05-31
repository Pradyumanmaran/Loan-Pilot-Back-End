package handler

import (
	"github.com/gofiber/fiber/v2"
	"loanpilot-backend/model"
	"loanpilot-backend/service"
)

type AssessmentHandler struct {
	service *service.AssessmentService
}

func NewAssessmentHandler(svc *service.AssessmentService) *AssessmentHandler {
	return &AssessmentHandler{service: svc}
}

func (h *AssessmentHandler) CreateAssessment(c *fiber.Ctx) error {
	var req model.AssessmentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request payload",
		})
	}

	resp := h.service.Assess(req)
	return c.JSON(resp)
}
