# GPTCode CLI

[![CI](https://img.shields.io/github/actions/workflow/status/gptcode-cloud/cli/ci.yml?label=CI)](https://github.com/gptcode-cloud/cli/actions/workflows/ci.yml)
[![CD](https://img.shields.io/github/actions/workflow/status/gptcode-cloud/cli/cd.yml?label=CD)](https://github.com/gptcode-cloud/cli/actions/workflows/cd.yml)
[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](#)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](#license)

Open-source AI coding assistant with specialized agents, validation, and an ML-driven model selector. Fast, affordable, and transparent. Works from your terminal and Neovim.

- Autonomous mode with Analyzer → Planner → Editor → Validator
- Smart context via dependency graph + PageRank
- ML intent/complexity classifiers and model recommender
- Cost-aware profiles and multiple backends (Groq, OpenRouter, Ollama)

## Quick Start

### Install (Go)
```bash
go install github.com/gptcode-cloud/cli/cmd/gptcode@latest
```

### Install (prebuilt binaries)
```bash
curl -fsSL https://raw.githubusercontent.com/gptcode-cloud/cli/main/install.sh | bash
```

### Setup and first run
```bash
gptcode setup
# Chat
gptcode chat "add user authentication with JWT"
# Research / Plan / Implement
gptcode research "error handling best practices"
gptcode plan "implement rate limiting"
gptcode implement <plan-id>
```

## Profiles and Backends
```bash
# Show backends
gptcode backend list
# Use a backend profile
gptcode profiles use openrouter.free
# Set models per agent
gptcode profiles set-agent openrouter free router google/gemini-2.0-flash-exp:free
```

## Neovim Integration
Minimal example (lazy.nvim):
```lua
{
  dir = "~/workspace/gptcode/neovim",
  config = function()
    require("gptcode").setup()
  end,
  keys = {
    { "<C-d>", "<cmd>GPTCodeChat<cr>", desc = "Toggle Chat" },
    { "<C-m>", "<cmd>GPTCodeModels<cr>", desc = "Profiles" },
  }
}
```

## Feedback and Model Learning
- Record feedback: `gptcode feedback good|bad --backend=<b> --model=<m> --agent=<a> --task "..."`
- Stats: `gptcode feedback stats`
- Export anonymized data (for community dataset): `gptcode feedback export --dry-run`

## Documentation
- Commands: `docs/` and the website pages under `docs/`
- Concepts: agent orchestration, dependency graph, ML features
- Blog posts for positioning and guides

## Contributing
We welcome contributions! Please read [CONTRIBUTING.md](CONTRIBUTING.md).

- Go 1.24+
- Tests: `go test ./...`
- Linting and CI are enforced

## License
MIT — see CONTRIBUTING for details.
