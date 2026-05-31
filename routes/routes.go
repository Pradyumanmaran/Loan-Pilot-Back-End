package routes

import (
	"github.com/gofiber/fiber/v2"
	"loanpilot-backend/handler"
	"loanpilot-backend/service"
)

func Register(app *fiber.App) {
	assessmentService := service.NewAssessmentService()
	assessmentHandler := handler.NewAssessmentHandler(assessmentService)

	app.Post("/assessment", assessmentHandler.CreateAssessment)
}
