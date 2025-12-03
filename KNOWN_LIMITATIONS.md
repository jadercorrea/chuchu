# Known Limitations & Issues

Documentação de limitações conhecidas e issues encontradas durante testes de capacidades.

## Status: Em Progresso
Última atualização: 2025-12-03

---

## 1. Validator Issues

### 1.1 Build Verification Too Strict
**Status**: ❌ Critical Issue  
**Impact**: Bloqueia tarefas simples

**Problema**:
O BuildVerifier tenta rodar `go build` em TODOS os projetos, mesmo quando:
- Projeto não é Go
- Tarefa é apenas ler/informar (git status, gh pr list)
- Arquivos modificados são apenas markdown/docs

**Exemplo**:
```bash
chu do "run git status and tell me if the repo is clean"
# Falha com: "go build failed with exit status 1"
```

**Root Cause**:
- `internal/maestro/verifier.go` detecta linguagem do diretório
- Sempre aplica build verification se detectar Go
- Não considera tipo da tarefa ou arquivos modificados

**Fix Proposta**:
1. Verificar se tarefa é read-only (git status, gh pr list, etc.)
2. Skip build se nenhum arquivo .go foi modificado
3. Tornar build optional para tarefas de análise/informação

### 1.2 Success Criteria Overly Specific
**Status**: ⚠️ Moderate Issue  
**Impact**: Retry loops desnecessários

**Problema**:
Success criteria gerados pelo Planner são muito específicos e literais:
- "git status must show information about uncommitted changes" → falha se repo está limpo
- "git log must include commit hashes" → formato esperado não bate
- "output must include remote repository status" → tarefa não pediu isso

**Fix Proposta**:
- Planner deve gerar critérios baseados apenas no objetivo da tarefa
- Evitar critérios sobre formato de output
- Focar em "tarefa concluída" vs "output específico"

---

## 2. Task Decomposition Issues

### 2.1 Simple Tasks Over-Complicated
**Status**: ⚠️ Moderate Issue  
**Impact**: Performance/custo

**Problema**:
Tarefas simples são decompostas em múltiplos movements desnecessariamente:
- "run git status" → 3 movements
- "show git log" → 2 movements
- Deveria ser single-shot

**Exemplo**:
```
Task: "run git status and tell me if the repo is clean"
Decomposed into:
1. Execute Git Status
2. Analyze Output
3. Report Status
```

**Fix Proposta**:
- ML classifier melhor para detectar tasks verdadeiramente simples
- Threshold de complexidade mais alto
- Bypass decomposition para read-only shell commands

---

## 3. Model Selection Issues

### 3.1 Catalog Backward Compatibility
**Status**: ✅ Fixed (commit 0957373)  
**Impact**: Bloqueador total

**Problema**:
Models sem campo `capabilities` eram rejeitados com:
```
no suitable model found for action=edit lang=go
```

**Fix Aplicado**:
- Default `supports_file_operations=true` para backward compatibility
- Models legacy agora funcionam

### 3.2 Language Detection False Positives
**Status**: ⚠️ Moderate Issue  
**Impact**: Seleção subótima de modelos

**Problema**:
- Detecta "go" por estar em cwd com projeto Go
- Aplica filtro de linguagem mesmo para tarefas agnósticas
- Exemplo: "gh pr list" não é específico de Go

**Fix Proposta**:
- Detectar linguagem do **alvo** da tarefa, não do cwd
- Tasks de CLI/shell não devem ter language filter
- Only apply language quando editando código

---

## 4. In-Memory Processing

### 4.1 Success Story ✅
**Status**: ✅ Working  
**Impact**: Positive

**O que funciona**:
- Editor usa `run_command` corretamente
- Não cria arquivos intermediários
- Processa output de comandos em memória

**Evidência**:
```bash
chu do "use gh pr list to get open PRs"
# Movement 1 completa SEM criar pull_requests.md
# Output processado diretamente em memória
```

**Melhorias Aplicadas**:
1. Editor prompt com exemplos de `run_command` (commit 8501d6b)
2. Decomposer evita `output_files` intermediários (commit ba5cbcf)
3. Planner instruído a não criar arquivos de dados (commit e0ae29d)

---

## 5. GitHub Integration

### 5.1 PR Review Comments
**Status**: 🔲 Not Implemented  
**Impact**: Feature gap

**Capacidade Ausente**:
- Criar review comments em linhas específicas
- Responder a comments existentes
- Marcar threads como resolved

**Needs**:
- `gh pr review` com `--comment-body` e `--line`
- Tool definitions para GitHub review API
- Integration com code diff

---

## 6. Multi-Tool Orchestration

### 6.1 Complex Pipelines
**Status**: 🔲 Not Tested  
**Impact**: Unknown

**Não Validado**:
- Chains de 3+ tools (curl | jq | grep)
- Error handling em pipelines
- Partial results em caso de falha

**Test Needed**:
```bash
chu do "fetch github API, extract stargazers, and save top 10 to file"
```

---

## 7. Documentation Tasks

### 7.1 Blog Post Generation
**Status**: ⚠️ Partial  
**Impact**: Original task que motivou melhorias

**O que funciona**:
- Obter PRs com `gh pr list`
- Processar content em memória

**O que não funciona**:
- Success criteria incorretos (go build em markdown project)
- Validator espera comandos git que não fazem sentido

**Workaround**:
Usar `chu do` com tasks mais específicas:
```bash
chu do "read file content from PR #7 and create blog post in docs/_posts/"
```

---

## 8. Performance & Cost

### 8.1 Token Usage
**Status**: ⚠️ Needs Optimization  
**Impact**: Cost

**Observado**:
- Simple tasks usando 10-20K tokens
- Múltiplos retry loops aumentam custo
- Validation failures custam 3x o esperado

**Otimizações Possíveis**:
- Cache de plans similares
- Streaming responses ao invés de full context
- Parallel validation quando possível

---

## Test Matrix Summary

| Category | Tested | Working | Partial | Failing |
|----------|--------|---------|---------|---------|
| Git Ops | 3 | 0 | 0 | 3 |
| GitHub CLI | 0 | 0 | 0 | 0 |
| Code Gen | 2 | ? | ? | ? |
| Docs | 2 | ? | ? | ? |
| Modification | 2 | ? | ? | ? |
| Multi-Tool | 2 | ? | ? | ? |
| Packages | 1 | ? | ? | ? |
| Testing | 1 | ? | ? | ? |
| Docker | 1 | ? | ? | ? |
| CI/CD | 1 | ? | ? | ? |

**Total Capabilities**: 210 documented  
**Total Tested**: 15 (~7%)  
**Estimated Working**: 60-70% (based on architecture)

---

## Priority Fixes

### P0 - Blocker
1. ✅ Model selection backward compatibility (FIXED)
2. ❌ **Build verification too strict** → blocks all non-Go tasks

### P1 - High Impact
3. ⚠️ Success criteria overly specific → waste retries
4. ⚠️ Task decomposition over-complicated → waste tokens/cost

### P2 - Medium Impact  
5. Language detection false positives
6. PR review comments missing

### P3 - Nice to Have
7. Multi-tool complex pipelines
8. Performance optimizations

---

## Next Steps

1. **Fix BuildVerifier** (P0)
   - Add task type detection
   - Skip build for read-only tasks
   - Check modified files before running

2. **Test Suite Completion**
   - Run full test suite after P0 fix
   - Document results in CAPABILITIES_TEST_RESULTS.md
   - Update roadmap with status

3. **Success Criteria Improvement** (P1)
   - Refine Planner prompt
   - Add examples of good vs bad criteria
   - Test with variety of tasks

4. **Documentation**
   - Create guides for each working category
   - Document workarounds for known issues
   - Add troubleshooting section

---

## Contributing

Found a limitation not listed here?
1. Test with latest chu build
2. Document: `chu do "your task"` + error output
3. Add to this file with PR
4. Tag with priority (P0/P1/P2/P3)
