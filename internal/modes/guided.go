package modes

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"chuchu/internal/agents"
	"chuchu/internal/config"
	"chuchu/internal/events"
	"chuchu/internal/llm"
	"chuchu/internal/ml"
)

type GuidedMode struct {
	events       *events.Emitter
	provider     llm.Provider
	baseProvider llm.Provider
	cwd          string
	model        string
	editorModel  string
}

func NewGuidedMode(provider llm.Provider, cwd string, model string) *GuidedMode {
	return &GuidedMode{
		events:       events.NewEmitter(os.Stderr),
		provider:     provider,
		baseProvider: provider,
		cwd:          cwd,
		model:        model,
		editorModel:  model,
	}
}

func NewGuidedModeWithCustomModel(provider llm.Provider, baseProvider llm.Provider, cwd string, model string, editorModel string) *GuidedMode {
	return &GuidedMode{
		events:       events.NewEmitter(os.Stderr),
		provider:     provider,
		baseProvider: baseProvider,
		cwd:          cwd,
		model:        model,
		editorModel:  editorModel,
	}
}

func (g *GuidedMode) Execute(ctx context.Context, userMessage string) error {
	_, err := g.ExecuteAndReturnPlan(ctx, userMessage)
	return err
}

func (g *GuidedMode) ExecuteAndReturnPlan(ctx context.Context, userMessage string) (string, error) {
	_ = g.events.Status("Analyzing task...")

	draftPlan, err := g.createDraftPlan(ctx, userMessage)
	if err != nil {
		return "", fmt.Errorf("failed to create draft: %w", err)
	}

	draftPath, err := g.saveDraft(draftPlan)
	if err != nil {
		return "", fmt.Errorf("failed to save draft: %w", err)
	}

	_ = g.events.OpenPlan(draftPath)
	_ = g.events.Message("Draft plan created.")

	_ = g.events.Status("Creating detailed plan...")

	fullPlan, err := g.createDetailedPlan(ctx, userMessage, draftPlan)
	if err != nil {
		return "", fmt.Errorf("failed to create plan: %w", err)
	}

	planPath, err := g.savePlan(fullPlan)
	if err != nil {
		return "", fmt.Errorf("failed to save plan: %w", err)
	}

	_ = g.events.OpenPlan(planPath)
	_ = g.events.Message("Detailed plan ready. Send 'implement' to start implementation.")

	return fullPlan, nil
}

func (g *GuidedMode) createDraftPlan(ctx context.Context, task string) (string, error) {
	prompt := fmt.Sprintf(`You are creating a DRAFT implementation plan.

Task: %s

Create a brief outline covering:
1. Overview (2-3 sentences)
2. Key steps (3-5 bullet points)
3. Files likely affected

Keep it concise - this is just a draft for user approval.`, task)

	resp, err := g.provider.Chat(ctx, llm.ChatRequest{
		SystemPrompt: "You create concise technical plans.",
		UserPrompt:   prompt,
		Model:        g.model,
	})
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}

func (g *GuidedMode) createDetailedPlan(ctx context.Context, task string, draft string) (string, error) {
	research := "No research available"

	if orchestrator, ok := g.provider.(*llm.OrchestratorProvider); ok {
		researchAgent := agents.NewResearch(orchestrator)

		statusCallback := func(status string) {
			_ = g.events.Status(status)
		}

		history := []llm.ChatMessage{
			{Role: "user", Content: fmt.Sprintf("Research this codebase for: %s", task)},
		}

		if result, err := researchAgent.Execute(ctx, history, statusCallback); err == nil {
			research = result
		}
	}

	prompt := fmt.Sprintf(`Create a SIMPLE, DIRECT implementation plan.

Task: %s

Draft outline:
%s

Codebase research:
%s

IMPORTANT CONSTRAINTS:
- Keep it MINIMAL - only what's strictly necessary
- NO extra features, NO tests unless explicitly requested
- NO scripts, NO automation unless asked
- ONLY modify files that already exist OR that are explicitly requested
- Focus on the EXACT task, nothing more

Create a brief plan:
# Plan

## What to do
[1-2 sentences]

## Files to modify
[List ONLY files that exist or are explicitly requested]

## Changes
[Specific, minimal changes to make]`, task, draft, research)

	resp, err := g.provider.Chat(ctx, llm.ChatRequest{
		SystemPrompt: "You create MINIMAL, DIRECT plans. Focus ONLY on what's asked. NO extra features.",
		UserPrompt:   prompt,
		Model:        g.model,
	})
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}

func (g *GuidedMode) Implement(ctx context.Context, plan string) error {
	editorAgent := agents.NewEditor(g.baseProvider, g.cwd, g.editorModel)

	statusCallback := func(status string) {
		_ = g.events.Status(status)
	}

	implementPrompt := fmt.Sprintf(`Implement EXACTLY what this plan says - NOTHING MORE:

---
%s
---

RULES:
- ONLY modify files explicitly listed in the plan
- ONLY make the specific changes described
- Do NOT create extra files (scripts, tests, configs) unless the plan says to
- Do NOT add features not mentioned in the plan
- If a file doesn't exist and isn't explicitly requested, DON'T create it
- When done, stop - do NOT keep iterating

Execute the plan directly and minimally.`, plan)

	history := []llm.ChatMessage{
		{Role: "user", Content: implementPrompt},
	}

	result, err := editorAgent.Execute(ctx, history, statusCallback)
	if err != nil {
		return err
	}
	
	if os.Getenv("CHUCHU_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[IMPLEMENT] Editor result: %s\n", result)
	}
	
	// Check if editor just returned without doing anything meaningful
	if strings.Contains(result, "reached max iterations") {
		return fmt.Errorf("editor reached max iterations without completing task")
	}
	
	return nil
}

func (g *GuidedMode) saveDraft(content string) (string, error) {
	home, _ := os.UserHomeDir()
	dir := fmt.Sprintf("%s/.chuchu/plans", home)
	_ = os.MkdirAll(dir, 0755)

	path := fmt.Sprintf("%s/draft.md", dir)
	return path, os.WriteFile(path, []byte(content), 0644)
}

func (g *GuidedMode) savePlan(content string) (string, error) {
	home, _ := os.UserHomeDir()
	dir := fmt.Sprintf("%s/.chuchu/plans", home)
	_ = os.MkdirAll(dir, 0755)

	timestamp := time.Now().Format("2006-01-02-150405")
	path := fmt.Sprintf("%s/%s-plan.md", dir, timestamp)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	currentPath := fmt.Sprintf("%s/.chuchu/current_plan.txt", home)
	_ = os.WriteFile(currentPath, []byte(content), 0644)

	return path, nil
}

func (g *GuidedMode) waitForConfirmation(prompt string, id string) bool {
	if err := g.events.Confirm(prompt, id); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to emit confirm event: %v\n", err)
	}

	os.Stdout.Sync()
	time.Sleep(100 * time.Millisecond)

	responseChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			errorChan <- err
			return
		}
		responseChan <- response
	}()

	timeout := time.After(120 * time.Second)
	select {
	case response := <-responseChan:
		response = strings.TrimSpace(strings.ToLower(response))
		return response == "y" || response == "yes"
	case <-errorChan:
		return false
	case <-timeout:
		_ = g.events.Notify("Timeout waiting for confirmation", "warn")
		return false
	}
}

func IsComplexTask(message string) bool {
	// Try to load ML model
	p, err := ml.LoadEmbedded("complexity_detection")
	if err != nil {
		return false
	}
	label, probs := p.Predict(message)
	if label == "multistep" {
		return true
	}
	if label == "complex" {
		threshold := 0.55
		if setup, err2 := config.LoadSetup(); err2 == nil {
			if setup.Defaults.MLComplexThreshold > 0 {
				threshold = setup.Defaults.MLComplexThreshold
			}
		}
		if v, ok := probs["complex"]; ok && v >= threshold {
			return true
		}
	}
	return false
}
