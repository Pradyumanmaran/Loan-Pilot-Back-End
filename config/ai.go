package config

import "os"

type AIConfig struct {
	Provider         string
	GeminiAPIKey     string
	OpenRouterAPIKey string
}

func LoadAIConfig() AIConfig {
	return AIConfig{
		Provider:         os.Getenv("AI_PROVIDER"),
		GeminiAPIKey:     os.Getenv("GEMINI_API_KEY"),
		OpenRouterAPIKey: os.Getenv("OPENROUTER_API_KEY"),
	}
}

// TODO:
// Implement Gemini provider
// TODO:
// Implement OpenRouter provider
// TODO:
// Allow switching providers via AI_PROVIDER
