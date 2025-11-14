package modes

import (
	"chuchu/internal/config"
	"chuchu/internal/llm"
	"chuchu/internal/memory"
	"chuchu/internal/prompt"
	"context"
	"fmt"
	"strings"
)

func Chat(input string, args []string) {
	setup, _ := config.LoadSetup()
	store, _ := memory.LoadStore()

	lang := setup.Defaults.Lang
	if len(args) >= 1 && args[0] != "" {
		lang = args[0]
	}

	backendName := setup.Defaults.Backend
	modelAlias := setup.Defaults.Model

	if len(args) >= 2 && args[1] != "" {
		backendName = args[1]
	}
	if len(args) >= 3 && args[2] != "" {
		modelAlias = args[2]
	}

	backendCfg := setup.Backend[backendName]
	model := backendCfg.DefaultModel
	if alias, ok := backendCfg.Models[modelAlias]; ok {
		model = alias
	} else if modelAlias != "" {
		model = modelAlias
	}

	pb := prompt.NewDefaultBuilder(store)
	hint := input
	if len(hint) > 200 {
		hint = hint[:200]
	}
	sys := pb.BuildSystemPrompt(prompt.BuildOptions{
		Lang: lang,
		Mode: "chat",
		Hint: hint,
	})

	sys += "\n\n## Conversational Flow for Code Generation\n\n" +
		"When the user asks you to create code, implement features, or solve technical problems:\n\n" +
		"1. **Clarify first** - Ask specific questions about:\n" +
		"   - Environment and deployment preferences\n" +
		"   - Data formats and persistence\n" +
		"   - Frequency/scheduling if applicable\n" +
		"   - Performance/cost constraints\n" +
		"   - Preferred languages/tools\n\n" +
		"2. **After clarification**, propose a plan:\n" +
		"   - List files/modules to create or modify\n" +
		"   - Describe the approach\n" +
		"   - Ask for confirmation\n\n" +
		"3. **Generate code ONLY after confirmation**:\n" +
		"   - Use ```tests and ```impl blocks for each module\n" +
		"   - Each test/impl pair becomes a separate tab in the editor\n" +
		"   - Format: ```tests\\n# file: path/to/test.ext\\ntest code\\n```\n" +
		"   - Format: ```impl\\n# file: path/to/impl.ext\\nimplementation code\\n```\n\n" +
		"**IMPORTANT**: Do NOT generate code blocks immediately. Always clarify and plan first.\n"

	user := input

	var provider llm.Provider
	switch backendCfg.Type {
	case "openai":
		provider = llm.NewOpenAI(backendCfg.BaseURL, backendName)
	default:
		provider = llm.NewOllama(backendCfg.BaseURL)
	}

	resp, err := provider.Chat(context.Background(), llm.ChatRequest{
		SystemPrompt: sys,
		UserPrompt:   user,
		Model:        model,
	})
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	fmt.Println(strings.TrimSpace(resp.Text))
}

func RunChat(builder *prompt.Builder, provider llm.Provider, model string) error {
	input := ""
	args := []string{}
	Chat(input, args)
	return nil
}
