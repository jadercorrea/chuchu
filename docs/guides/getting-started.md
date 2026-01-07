# Getting Started

This guide helps you install GPTCode, configure providers, and start using the core workflows. It also includes a 10‑second quick start for universal feedback capture.

## Install

- Build from source:
```bash
# from repository root
go build -o ./bin/gptcode ./cmd/gptcode
```
- Or use your preferred package manager (coming soon).

## Initial setup
```bash
gt setup             # creates ~/.gptcode/setup.yaml
gptcode key openrouter    # add API key(s) as needed
gt backend           # check current backend
gt backend list      # list all backends
gt profile           # check current profile
```

## Quick start: two‑keystroke feedback (Ctrl+g)
Capture corrections from any CLI as training signals.
```bash
# zsh
gt feedback hook install --with-diff --and-source

# bash
gt feedback hook install --shell=bash --with-diff --and-source

# fish
gt feedback hook install --shell=fish --with-diff
```
Usage:
1) Type/paste the suggested command
2) Press Ctrl+g to mark the suggestion
3) Edit (if needed) and press Enter

GPTCode records good/bad outcomes and saves changed files and optional git patch.

Check stats:
```bash
gt feedback stats
```

## Core commands

- Chat (code‑focused Q&A):
```bash
gt chat "how does auth middleware work?"
```

- Orchestrated execution (Analyzer → Planner → Editor → Validator):
```bash
gt do "add feature"
gt do --supervised "refactor module"
```

- Model management:
```bash
gt model list                           # see all available models
gt model recommend editor               # get recommendation
gt model set groq/compound              # set default model
```

## Recommended Models by Backend

> **Important**: Not all models support tool calling. Use these recommended models for best results:

| Backend | Model | Context | Best For |
|---------|-------|---------|----------|
| **Groq** | `groq/compound` | 128k | General use, tool calling ✅ |
| **Groq** | `llama-3.3-70b-versatile` | 128k | Fast, large context |
| **OpenRouter** | `openrouter/auto` | varies | Auto-routing, best model per task |
| **Ollama** | `qwen2.5-coder:32b` | 32k | Local, tool calling ✅ |
| **Ollama** | `llama3.3:70b` | 128k | Local, large context |

```bash
# Set your default model
gt model set groq/compound

# Or for local
gt model set qwen2.5-coder:32b
```

## Troubleshooting
- Missing API keys: `gptcode key <backend>`
- Hook not active: `source ~/.zshrc` (or your shell rc) and try Ctrl+g again
- Files not captured: ensure you are inside a git repo
