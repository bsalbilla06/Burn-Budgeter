package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/genai"
)

// DetectedService represents a service found by the AI
type DetectedService struct {
	Provider string  `json:"provider"`
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
}

// ParseArchitecture uses Gemini to extract services from an ARCHITECTURE.md content string.
func ParseArchitecture(ctx context.Context, content string) ([]DetectedService, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %v", err)
	}

	// In the new SDK, SystemInstruction is part of GenerateContentConfig
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: "You are an expert cloud and AI architect. Your task is to analyze ARCHITECTURE.md files and extract a list of cloud (AWS, GCP) and AI (OpenAI, Anthropic, Gemini) services mentioned. For each service, provide the provider, a concise name, and a numerical monthly quantity. IMPORTANT: The quantity MUST be a number ONLY (e.g. 730, 10, 5). For AI models, use 'OpenAI', 'Anthropic', or 'Gemini' as provider. Prefer the latest version available in our database: GPT-5.2 for OpenAI, Claude 4.6 for Anthropic, and Gemini 3.1 for Google. Return ONLY a JSON array of objects with fields: provider, name, quantity."},
			},
		},
		ResponseMIMEType: "application/json",
	}


	// Use a reliable model for the demo
	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(content), config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response candidates received from AI")
	}

	text := resp.Candidates[0].Content.Parts[0].Text
	if text == "" {
		return nil, fmt.Errorf("empty response received from AI")
	}

	var services []DetectedService
	if err := json.Unmarshal([]byte(text), &services); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON response: %v\nResponse text: %s", err, text)
	}

	return services, nil
}

// GenerateArchitectureMarkdown uses Gemini to create a readable ARCHITECTURE.md from a list of services.
func GenerateArchitectureMarkdown(ctx context.Context, services []string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create genai client: %v", err)
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: "You are an expert cloud architect. Your task is to write a professional and clear ARCHITECTURE.md file based on a list of infrastructure services provided. Use appropriate headings and bullet points. Describe the stack logically (Compute, Storage, AI, etc.). Return ONLY the markdown content."},
			},
		},
	}

	prompt := "Create an ARCHITECTURE.md for a project using these services:\n"
	for _, s := range services {
		prompt += fmt.Sprintf("- %s\n", s)
	}

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), config)
	if err != nil {
		return "", fmt.Errorf("failed to generate markdown: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response candidates received from AI")
	}

	return resp.Candidates[0].Content.Parts[0].Text, nil
}
