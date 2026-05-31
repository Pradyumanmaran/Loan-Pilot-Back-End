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
	if err != nil {
		log.Println("No .env file found")
	}

	log.Println("Gemini Key Loaded:", os.Getenv("GEMINI_API_KEY") != "")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	routes.Register(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "LoanPilot AI Backend Running",
		})
	})

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)

	log.Fatal(app.Listen(":" + port))
}
