package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"loanpilot-backend/routes"
	"log"
	"os"
)

func main() {

	err := godotenv.Load()
	log.Println("Gemini Key Loaded:", os.Getenv("GEMINI_API_KEY") != "")
	if err != nil {
		log.Println("No .env file found")
	}
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	routes.Register(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "LoanPilot AI Backend Running",
		})
	})

	log.Fatal(app.Listen(":3000"))
}
