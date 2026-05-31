package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type AIService interface {
	GenerateUnderwritingSummary(req AISummaryRequest) (string, error)
}

type AISummaryRequest struct {
	Status               int
	CompanyCategory      string
	RequestedAmount      float64
	FinalEligibility     float64
	RequiredSalary       float64
	RequiredEmiReduction float64
	SelectedMethod       string
	IsCapped             bool
	MaxLoanAmount        float64
}

type MockAIService struct{}

func (m *MockAIService) GenerateUnderwritingSummary(req AISummaryRequest) (string, error) {
	return fallbackUnderwritingSummary(req), nil
}

type GeminiAIService struct {
	apiKey     string
	httpClient *http.Client
}

func NewGeminiAIService() *GeminiAIService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	service := &GeminiAIService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	log.Printf(
		"Gemini service initialized: apiKeyConfigured=%t, apiKeyLength=%d, httpClientCreated=%t",
		strings.TrimSpace(apiKey) != "",
		len(strings.TrimSpace(apiKey)),
		service.httpClient != nil,
	)
	return service
}

func (g *GeminiAIService) GenerateUnderwritingSummary(req AISummaryRequest) (string, error) {
	fallback := fallbackUnderwritingSummary(req)
	if g == nil || strings.TrimSpace(g.apiKey) == "" {
		log.Printf("Gemini API error: %v", fmt.Errorf("gemini api key is not configured"))
		log.Printf("Gemini fallback summary returned: %s", fallback)
		return fallback, fmt.Errorf("gemini api key is not configured")
	}

	prompt := buildGeminiUnderwritingPrompt(req)
	log.Printf("Gemini API request started: model=gemini-2.0-flash, selectedMethod=%s, status=%d, requestedAmount=%.0f, finalEligibility=%.0f, isCapped=%t",
		req.SelectedMethod, req.Status, req.RequestedAmount, req.FinalEligibility, req.IsCapped)
	body := geminiGenerateContentRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:     floatPtr(0.2),
			MaxOutputTokens: intPtr(800),
			TopP:            floatPtr(0.95),
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		log.Printf("Gemini API error: request marshal failed: %v", err)
		log.Printf("Gemini fallback summary returned: %s", fallback)
		return fallback, err
	}

	endpoint := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		log.Printf("Gemini API error: request creation failed: %v", err)
		log.Printf("Gemini fallback summary returned: %s", fallback)
		return fallback, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Gemini API error: request failed: %v", err)
		log.Printf("Gemini fallback summary returned: %s", fallback)
		return fallback, err
	}
	defer resp.Body.Close()

	log.Printf("Gemini API response received: statusCode=%d", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Gemini API error: response read failed: %v", err)
		log.Printf("Gemini fallback summary returned: %s", fallback)
		return fallback, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := fmt.Errorf("gemini api error: %s", strings.TrimSpace(string(respBody)))
		log.Printf("Gemini API error: %v", apiErr)
		log.Printf("Gemini fallback summary returned: %s", fallback)
		return fallback, apiErr
	}

	log.Printf("Raw Gemini response body: %s", string(respBody))
	log.Printf("Raw Gemini response JSON: %q", string(respBody))

	var geminiResp geminiGenerateContentResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		log.Printf("Gemini API error: response unmarshal failed: %v", err)
		log.Printf("Gemini fallback summary returned: %s", fallback)
		return fallback, err
	}

	log.Printf("Parsed Gemini response: candidates=%d, firstCandidateParts=%d, finishReason=%s", len(geminiResp.Candidates), firstGeminiPartCount(geminiResp), firstGeminiFinishReason(geminiResp))

	summary := strings.TrimSpace(firstGeminiText(geminiResp))
	if summary == "" {
		log.Printf("Gemini API error: empty summary returned")
		log.Printf("Gemini fallback summary returned: %s", fallback)
		return fallback, fmt.Errorf("gemini returned an empty summary")
	}

	log.Printf("Gemini generated text length: %d", len(summary))
	log.Printf("Gemini generated text quoted: %q", summary)
	log.Printf("FINAL SUMMARY RETURNED TO API: %q", summary)
	return summary, nil
}

type OpenRouterAIService struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenRouterAIService() *OpenRouterAIService {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	model := strings.TrimSpace(os.Getenv("OPENROUTER_MODEL"))
	if model == "" {
		model = "openrouter/free"
	}
	service := &OpenRouterAIService{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	log.Printf(
		"OpenRouter service initialized: apiKeyConfigured=%t, apiKeyLength=%d, model=%s, httpClientCreated=%t",
		strings.TrimSpace(apiKey) != "",
		len(strings.TrimSpace(apiKey)),
		service.model,
		service.httpClient != nil,
	)
	return service
}

func (o *OpenRouterAIService) GenerateUnderwritingSummary(req AISummaryRequest) (string, error) {
	fallback := fallbackUnderwritingSummary(req)
	log.Printf("[AI] Provider=openrouter")
	log.Printf("[AI] Model=%s", o.selectedModel())
	log.Printf("[AI] Request started")

	if o == nil || strings.TrimSpace(o.apiKey) == "" {
		err := fmt.Errorf("openrouter api key is not configured")
		log.Printf("[AI] Error: %v", err)
		log.Printf("[AI] Fallback summary returned: %s", fallback)
		return fallback, err
	}

	body := openRouterChatCompletionsRequest{
		Model: o.selectedModel(),
		Messages: []openRouterMessage{
			{
				Role:    "user",
				Content: buildGeminiUnderwritingPrompt(req),
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		log.Printf("[AI] Error: request marshal failed: %v", err)
		log.Printf("[AI] Fallback summary returned: %s", fallback)
		return fallback, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		log.Printf("[AI] Error: request creation failed: %v", err)
		log.Printf("[AI] Fallback summary returned: %s", fallback)
		return fallback, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[AI] Error: request failed: %v", err)
		log.Printf("[AI] Fallback summary returned: %s", fallback)
		return fallback, err
	}
	defer resp.Body.Close()

	log.Printf("[AI] Response received: statusCode=%d", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[AI] Error: response read failed: %v", err)
		log.Printf("[AI] Fallback summary returned: %s", fallback)
		return fallback, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := fmt.Errorf("openrouter api error: %s", strings.TrimSpace(string(respBody)))
		log.Printf("[AI] Error: %v", apiErr)
		log.Printf("[AI] Fallback summary returned: %s", fallback)
		return fallback, apiErr
	}

	var openRouterResp openRouterChatCompletionsResponse
	if err := json.Unmarshal(respBody, &openRouterResp); err != nil {
		log.Printf("[AI] Error: response unmarshal failed: %v", err)
		log.Printf("[AI] Fallback summary returned: %s", fallback)
		return fallback, err
	}

	summary := strings.TrimSpace(openRouterResp.FirstMessageContent())
	if summary == "" {
		err := fmt.Errorf("openrouter returned an empty summary")
		log.Printf("[AI] Error: %v", err)
		log.Printf("[AI] Fallback summary returned: %s", fallback)
		return fallback, err
	}

	log.Printf("[AI] Final summary returned to API: %q", summary)
	return summary, nil
}

func (o *OpenRouterAIService) selectedModel() string {
	if o == nil || strings.TrimSpace(o.model) == "" {
		return "openrouter/free"
	}

	return o.model
}

type openRouterChatCompletionsRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterChatCompletionsResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (r openRouterChatCompletionsResponse) FirstMessageContent() string {
	if len(r.Choices) == 0 {
		return ""
	}

	return r.Choices[0].Message.Content
}

type geminiGenerateContentRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		FinishReason string `json:"finishReason"`
		Content      struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func buildGeminiUnderwritingPrompt(req AISummaryRequest) string {
	return strings.TrimSpace(fmt.Sprintf(
		`Generate a personal loan underwriting explanation using only the provided values.

Status: %d
Company Category: %s
Requested Amount: %.0f
Final Eligibility: %.0f
Required Salary: %.0f
Required EMI Reduction: %.0f
Selected Method: %s
Is Capped: %t
MaxLoanAmount: %.0f

Rules:
- Use only the values above.
- Never invent reasons.
- Never mention credit score, credit profile, bureau score, alternative products, or repayment history unless explicitly provided in the input.
- If Final Eligibility = 0, explain that current underwriting rules did not produce the minimum eligible loan amount.
- Keep the response between 60 and 120 words.
- Use a professional banking tone.
- Use Indian Rupee format.
- Mention only facts from the underwriting payload.
- Return only explanation text.`,
		req.Status,
		req.CompanyCategory,
		req.RequestedAmount,
		req.FinalEligibility,
		req.RequiredSalary,
		req.RequiredEmiReduction,
		req.SelectedMethod,
		req.IsCapped,
		req.MaxLoanAmount,
	))
}

func fallbackUnderwritingSummary(req AISummaryRequest) string {
	var summary string

	switch req.Status {
	case 3:
		summary = "The applicant satisfies current underwriting requirements and qualifies for the requested loan amount."
	case 2:
		summary = "The applicant qualifies for a lower amount than requested. Increasing income or reducing existing obligations may improve eligibility."
	case 1:
		summary = "The applicant does not currently satisfy the minimum underwriting requirements."
	default:
		summary = "The underwriting outcome could not be determined."
	}

	if req.IsCapped {
		summary += " The calculated eligibility exceeds the product limit and has been capped at the maximum personal loan amount."
	}

	return summary
}

func firstGeminiText(resp geminiGenerateContentResponse) string {
	if len(resp.Candidates) == 0 {
		return ""
	}

	parts := resp.Candidates[0].Content.Parts
	var builder strings.Builder
	for _, part := range parts {
		if strings.TrimSpace(part.Text) == "" {
			continue
		}

		if builder.Len() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(strings.TrimSpace(part.Text))
	}

	return builder.String()
}

func firstGeminiPartCount(resp geminiGenerateContentResponse) int {
	if len(resp.Candidates) == 0 {
		return 0
	}

	return len(resp.Candidates[0].Content.Parts)
}

func firstGeminiFinishReason(resp geminiGenerateContentResponse) string {
	if len(resp.Candidates) == 0 {
		return ""
	}

	return resp.Candidates[0].FinishReason
}

func limitWords(text string, maxWords int) string {
	fields := strings.Fields(text)
	if len(fields) <= maxWords {
		return strings.TrimSpace(text)
	}

	return strings.Join(fields[:maxWords], " ")
}

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}
