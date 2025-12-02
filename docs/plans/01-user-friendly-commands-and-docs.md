# Objetivo
Atualizar CLI para comandos user-friendly e sincronizar toda documentação com o estado real do projeto, seguindo as diretrizes dos notebooks compartilhados.

# Contexto Atual
- ✅ Já implementamos `chu profile` (singular) com subcomandos friendly
- ⚠️ Ainda temos comandos verbosos como `chu config get defaults.backend`
- ⚠️ Documentação menciona comandos inexistentes ou desatualizados
- ⚠️ Foco em TDD quando deveria destacar agents + validation

# Mudanças no CLI

## 1. Backend Commands (Singular)
Adicionar `chu backend` (singular) com subcomandos friendly:

### Comandos Novos
```bash
chu backend                    # show current backend
chu backend list               # list all backends
chu backend show [name]        # show backend config
chu backend use <name>         # switch to backend
chu backend create <name> <type> <url>  # já existe
chu backend delete <name>      # já existe
```

### Implementação
- Criar `backendShowCmd` para mostrar backend atual
- Criar `backendUseCmd` para trocar backend (atualiza defaults.backend)
- Modificar `backendCmd` para ter RunE que mostra backend atual
- Arquivos: `cmd/chu/main.go`

## 2. Model Commands (já existe, verificar completude)
Verificar se `chu model` já tem todos os comandos friendly:
```bash
chu model list
chu model recommend
chu model install
chu model update
```

## 3. Atualizar Help Principal
Atualizar `rootCmd.Long` para:
- Remover menção a `chu config get/set` do help principal
- Adicionar `chu profile` e `chu backend` nos exemplos
- Manter categorização já feita

# Atualização de Documentação

## 4. README.md
Atualizar seções:

### Hero Section
- Manter mensagem atual sobre agents
- Adicionar diagrama de orchestração (se não existir)

### Backend Management (linha ~262)
Substituir:
```bash
chu config set defaults.backend mygroq
```
Por:
```bash
chu backend use mygroq
```

### Profile Management (linha ~280)
Adicionar exemplos com `chu profile` (singular):
```bash
chu profile                  # show current
chu profile list            # list all
chu profile use groq.speed  # switch
```

## 5. docs/commands.md
Atualizar seção de configuração:

### Backend Management
```bash
chu backend                 # Show current backend
chu backend list           # List all backends  
chu backend use groq       # Switch backend
chu backend create <name> <type> <url>
chu backend delete <name>
```

### Profile Management
```bash
chu profile                      # Show current profile
chu profile list [backend]       # List all profiles
chu profile show [backend.profile]
chu profile use backend.profile  # Switch profile
```

### Configuration (Advanced)
Mover `chu config get/set` para seção "Advanced" e marcar como "para uso avançado".

## 6. docs/index.md (Homepage)
Baseado no notebook "Repensando o Site":

### Hero Section
Verificar se já tem:
- Mensagem: "AI Coding Assistant with Specialized Agents"
- Subtítulo sobre Analyzer → Planner → Editor → Validator
- Menção a $0-5/month

### Features Cards
Ordem proposta:
1. 🚀 Agent Orchestration
2. ✅ File Validation & Success Criteria
3. 🧠 Intelligent Context (dependency graph)
4. 💰 Radically Affordable
5. 🔧 Supervised vs Autonomous
6. 🎯 Deep Neovim Integration

### Diagrama
Adicionar diagrama mermaid de agent orchestration (conforme notebook).

## 7. docs/features.md
Reestruturar seções:

### Nova ordem
1. Agent-Based Architecture
   - Analyzer, Planner, Editor, Validator
2. Validation & Safety
   - File validation, success criteria, over-engineering protection
3. Intelligence Features
   - ML routing, dependency graph, context optimization
4. Cost Optimization
   - Mix models per agent, profile management
5. Developer Experience
   - Neovim integration, workflow
6. TDD Features (mover para final)

## 8. Posts do Blog
Baseado no notebook "Revisão completa dos posts":

### Alta Prioridade (correções críticas)
- `2025-11-17-ollama-local-setup.md`
  - Remover seção de "hybrid setup" com múltiplos backends
  - Clarificar que só um backend por vez
  
- `2025-11-21-profile-management.md`
  - Atualizar exemplos para usar `chu profile` (singular)
  - Adicionar `chu profile use`

### Média Prioridade
- `2025-11-15-groq-optimal-configs.md`
  - Atualizar comandos de config
  
- `2025-11-16-openrouter-multi-provider.md`
  - Atualizar comandos de config

### Padrão de Busca e Substituição
Em todos os posts:
- `chu config set defaults.backend X` → `chu backend use X`
- `chu config set defaults.profile Y` → `chu profile use X.Y`
- `chu profiles list X` → `chu profile list X`
- `chu profiles show X Y` → `chu profile show X.Y`

## 9. Guias
- `docs/guides/getting-started.md`
  - Atualizar comandos de configuração
  - Usar `chu backend` e `chu profile` nos exemplos

# Ordem de Implementação

## Fase 1: CLI Core (Alta Prioridade)
1. ✅ Implementar `chu backend` (singular) commands
2. ✅ Atualizar help principal (remover config get/set)
3. ✅ Testar todos os comandos novos
4. ✅ Build e install

## Fase 2: Docs Core (Alta Prioridade)
5. ✅ Atualizar README.md (comandos)
6. ✅ Atualizar docs/commands.md
7. ✅ Atualizar docs/index.md (hero + features)

## Fase 3: Docs Features (Média Prioridade)
8. ✅ Reestruturar docs/features.md
9. ✅ Atualizar docs/guides/getting-started.md

## Fase 4: Blog Posts (Média Prioridade)
10. ✅ Atualizar profile-management.md
11. ✅ Atualizar ollama-local-setup.md
12. ✅ Buscar e substituir em outros posts

## Fase 5: Validação (Baixa Prioridade)
13. ⏳ Grep por `chu config set defaults` em todos os arquivos
14. ⏳ Verificar se algum doc menciona comandos inexistentes
15. ⏳ Criar checklist de comandos vs docs

# Success Criteria

## CLI
- ✅ `chu backend` mostra backend atual
- ✅ `chu backend use <name>` troca backend
- ✅ `chu profile use <backend>.<profile>` troca ambos
- ✅ Help principal não menciona config get/set
- ✅ Todos os comandos friendly funcionam

## Documentação
- ✅ README usa comandos friendly
- ✅ docs/commands.md atualizado
- ✅ docs/index.md destaca agents
- ✅ docs/features.md reorganizado
- ✅ Posts do blog atualizados
- ✅ Nenhuma menção a comandos inexistentes
- ✅ Mensagem consistente: "Agents + Validation"

## Consistência
- ✅ CLI help alinhado com docs
- ✅ Todos os exemplos usam comandos friendly
- ✅ TDD mencionado mas não dominante
- ✅ Foco em agents, orchestration, validation
