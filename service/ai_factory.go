package service

import (
	"log"
	"strings"

	"loanpilot-backend/config"
)

func NewAIService() AIService {
	cfg := config.LoadAIConfig()
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "gemini":
		log.Printf("[AI] Provider=gemini")
		return NewGeminiAIService()
	case "openrouter":
		log.Printf("[AI] Provider=openrouter")
		return NewOpenRouterAIService()
	default:
		log.Printf("[AI] Provider=mock")
		return &MockAIService{}
	}
}
