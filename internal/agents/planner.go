package agents

import (
	"context"
	"fmt"

	"chuchu/internal/llm"
)

type PlannerAgent struct {
	provider llm.Provider
	model    string
}

func NewPlanner(provider llm.Provider, model string) *PlannerAgent {
	return &PlannerAgent{
		provider: provider,
		model:    model,
	}
}

const plannerPrompt = `You are a minimal planner. Your ONLY job is to create focused, minimal plans.

WORKFLOW:
1. Read the task and analysis
2. List ONLY files that need changes
3. Describe ONLY necessary changes

CRITICAL RULES:
- Be EXTREMELY minimal
- NO extra features
- NO tests unless requested
- NO scripts unless requested
- ONLY modify existing files OR files explicitly requested
- Focus on the EXACT task

Create minimal, direct plans.`

func (p *PlannerAgent) CreatePlan(ctx context.Context, task string, analysis string, statusCallback StatusCallback) (string, error) {
	if statusCallback != nil {
		statusCallback("Planner: Creating minimal plan...")
	}

	planPrompt := fmt.Sprintf(`Create a MINIMAL implementation plan.

Task: %s

Codebase Analysis:
---
%s
---

Create a brief plan:
# Plan

## Files to modify
[List ONLY files that exist OR are explicitly requested]

## Changes
[Describe ONLY the minimal changes needed]

## Success Criteria
[How to verify it worked]

Keep it MINIMAL. NO extra features.`, task, analysis)

	resp, err := p.provider.Chat(ctx, llm.ChatRequest{
		SystemPrompt: plannerPrompt,
		UserPrompt:   planPrompt,
		Model:        p.model,
	})
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}
