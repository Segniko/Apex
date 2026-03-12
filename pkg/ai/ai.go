package ai

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// TacticalAI represents the core engine for chat intelligence powered by Gemini.
type TacticalAI struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewTacticalAI initializes a new AI engine using the Gemini API.
func NewTacticalAI(apiKey string) *TacticalAI {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("❌ APEX_AI: Failed to initialize Gemini client: %v", err)
		return nil
	}

	model := client.GenerativeModel("gemini-3-flash-preview")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text("You are APEX_AI, a tactical forensics unit for the Apex Monitoring System. " +
				"Your prime directive is real-time root-cause reconstruction and system audit. " +
				"You assist developers in understanding crash reports, telemetry, and system architecture. " +
				"Be concise, technical, and use tactical terminology. " +
				"IMPORTANT: Always use Markdown for structure (lists, code blocks). " +
				"DO NOT use '**' for bolding normal text or headers; use plain uppercase for headers instead. " +
				"Ensure perfect spacing, punctuation, and logical flow. Avoid dense blocks of text."),
		},
	}

	return &TacticalAI{
		client: client,
		model:  model,
	}
}

// Chat generates a tactical response based on the message and optional report context.
func (ai *TacticalAI) Chat(message string, reportContext string) string {
	if ai == nil || ai.client == nil {
		return "IDENTITY_ERROR: Gemini client not initialized. Ensure GEMINI_API_KEY is set in the environment."
	}

	ctx := context.Background()
	prompt := message
	if reportContext != "" {
		prompt = fmt.Sprintf("Context: [Project Crash ID: %s]\n\nUser Question: %s", reportContext, message)
	}

	resp, err := ai.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Sprintf("SIGNAL_LOSS: Failed to generate response: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "SIGNAL_NOISE: Data inconclusive. Gemini returned an empty response."
	}

	part := resp.Candidates[0].Content.Parts[0]
	return fmt.Sprintf("%v", part)
}

// AnalyzeReport performs a deep-trace forensic analysis of a crash report.
func (ai *TacticalAI) AnalyzeReport(message string, stackTrace string) string {
	if ai == nil || ai.client == nil {
		return "FORENSIC_LEVEL_1: Gemini not initialized. Basic pattern analysis suggested."
	}

	ctx := context.Background()
	prompt := fmt.Sprintf("Perform a tactical forensic analysis on this crash:\n\nError: %s\nStack Trace:\n%s\n\nProvide a structured breakdown including 'ROOT_CAUSE:', 'IMPACT_ASSESSMENT:', and a concise 'TACTICAL_FIX:'. Use plain text headers (uppercase) and lists for clarity. Do not use '**' for bolding.", message, stackTrace)

	resp, err := ai.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Sprintf("FORENSIC_SIG_LOSS: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "FORENSIC_DATA_GAP: Empty response from AI node."
	}

	part := resp.Candidates[0].Content.Parts[0]
	return fmt.Sprintf("%v", part)
}

// Close releases Gemini resources.
func (ai *TacticalAI) Close() {
	if ai != nil && ai.client != nil {
		ai.client.Close()
	}
}
