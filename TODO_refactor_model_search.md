# TODO: Refatorar busca de modelos

## Problema atual
A lógica de busca de modelos está no plugin Lua do Neovim (`neovim/lua/chuchu/init.lua`).
Isso é ruim porque:
- Frontend (Lua) não deveria ter lógica de negócio
- Código duplicado (mesma lógica em Lua e potencialmente em outros frontends)
- Difícil de testar
- Viola separação de responsabilidades

## Solução proposta

### 1. Criar comando CLI `chu models search`
```bash
chu models search --backend openrouter --query "gemini" --agent router
```

Output JSON:
```json
{
  "models": [
    {
      "id": "google/gemini-2.0-flash-exp:free",
      "name": "Google: Gemini 2.0 Flash Experimental (free)",
      "pricing_prompt": 0.0,
      "pricing_completion": 0.0,
      "recommended": true,
      "context_window": 1050000
    },
    ...
  ]
}
```

### 2. Plugin Lua chama o CLI
```lua
function M.search_models(backend, query, agent, callback)
  local cmd = string.format("chu models search --backend %s --query '%s' --agent %s", 
    backend, query, agent or "")
  
  vim.fn.jobstart(cmd, {
    on_stdout = function(_, data)
      local models = vim.fn.json_decode(table.concat(data, "\n"))
      callback(models)
    end
  })
end
```

### 3. Benefícios
- ✅ Lua fica "burro como uma pedra" (só UI)
- ✅ Lógica centralizada no Go (testável, reutilizável)
- ✅ Fácil adicionar outros frontends (VSCode, CLI interativo, etc)
- ✅ Busca, ordenação, filtragem tudo no Go
- ✅ Cache e otimizações no lugar certo

### 4. Arquivos a modificar
- `cmd/chu/main.go` - Adicionar comando `models search`
- `internal/catalog/catalog.go` - Adicionar função `SearchModels(backend, query, agent)`
- `neovim/lua/chuchu/init.lua` - Remover lógica de busca, chamar CLI

### 5. Testes
Mover `test_catalog_test.go` para `internal/catalog/search_test.go`

## Prioridade
Média - funciona, mas deveria ser refatorado para manutenibilidade
