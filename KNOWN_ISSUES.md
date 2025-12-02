# Known Issues

## 1. Model capability limits with Go package rules

**Status**: FUNDAMENTAL - Model understanding issue

**Description**:
When the Editor adds/modifies Go files, it consistently makes these errors:
1. Infers package name from filename: `utils.go` → `package utils` (WRONG)
2. Creates complete programs with `main()` when asked to "add function to file"
3. Doesn't fix these issues even with explicit feedback in retry loops

**Root Cause**:
The model (llama-3.3-70b-versatile) doesn't have strong enough understanding of:
- Go's "all files in same dir = same package" rule
- Difference between "add function to existing file" vs "create standalone program"

**Attempted Solutions (ALL REJECTED)**:
- ✗ Automatic package name correction in tools → GAMBIARRA, não faz sentido
- ✗ More detailed prompts → System prompt já é grande demais
- ✓ Enhanced Editor prompt with Go package warning → Kept, but not sufficient

**Real Solution**:
Usar modelo mais capaz para tarefas de código. Testado:
- ✗ qwen3-coder:latest (Ollama): Timeout - muito lento
- ✗ deepseek-r1:32b (Ollama): Travou na fase de planning
- ✗ groq/compound: Não suporta file tools

Opções viáveis:
1. GPT-4o / GPT-4o-mini (OpenAI): Melhor para código, precisa configurar backend
2. Claude Sonnet (Anthropic): Excelente para código, precisa configurar backend  
3. Gemini 2.0 Flash (Google): Via OpenRouter free tier

**Feedback System**:
✓ Feedback negativo registrado via `chu feedback bad` para treinar ML

**Workaround Atual**:
Manual fix após execução ou usar comando mais explícito:
```bash
chu do "read utils.go, add Divide function that takes two floats and returns float64 or error for division by zero, keep existing package declaration and functions"
```

**Affected Models**:
- llama-3.3-70b-versatile (Groq): Confirmado inadequado para essa tarefa

**Test Case**:
```bash
chu do "add Divide function to utils.go"
```
Expected: Divide function added with correct `package main`
Actual: Divide function added with `package utils`, compilation fails

**Files Involved**:
- `/Users/jadercorrea/workspace/opensource/chuchu/internal/maestro/conductor.go` - Retry loop
- `/Users/jadercorrea/workspace/opensource/chuchu/internal/agents/editor.go` - Editor agent
- `/Users/jadercorrea/workspace/opensource/chuchu/internal/agents/reviewer.go` - Reviewer agent

**Next Steps**:
1. Add debug logging to see exactly what feedback Editor receives
2. Experiment with different prompt phrasing for package name corrections
3. Consider adding explicit package name validation in Editor before file write
4. Test with other models (qwen3-coder, deepseek-r1, gpt-4o)

---

## 2. Ollama gpt-oss:latest timeout issues

**Status**: LOW PRIORITY

**Description**:
The gpt-oss:latest (120B) model is extremely slow (~14 seconds for simple 3-word response). The chuchu client times out before the model can respond, causing "context deadline exceeded" errors.

**Workaround**:
Use Groq models (faster) or increase client timeout.

**Affected Models**:
- gpt-oss:latest (Ollama): 120B parameter model

---

## 3. Model capability detection needed

**Status**: COMPLETED ✓

**Description**:
Models have different tool support capabilities. Some models (like groq/compound) support tools but NOT file operation tools (read_file, write_file). This causes cryptic "No endpoints found that support tool use" errors.

**Solution**:
Added `capabilities` field to model catalog with:
- `supports_tools`: true/false
- `supports_file_operations`: true/false  
- `supports_code_execution`: true/false
- `notes`: Human-readable description

**Known Capabilities**:
- llama-3.3-70b-versatile (Groq): ✓ file tools
- groq/compound (Groq): ✓ tools, ✗ file tools (has web_search, wolfram, code_interpreter)
- moonshotai/kimi-k2-instruct (Groq): ✗ no tool support
- gpt-oss:latest (Ollama): ✓ file tools (but very slow)

Script: `/Users/jadercorrea/workspace/opensource/chuchu/ml/scripts/add_model_capabilities.py`
