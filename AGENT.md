# AGENT.md — Guidelines para Agentes de IA trabalhando no LIA

> Este documento define as regras e melhores práticas para agentes de IA (LLMs) que contribuem com o projeto LIA.

---

## 1. CONTEXTO E MISSÃO

O **LIA (Linguagem Intermediária Assistida)** é um toolchain model-first para geração de software por IA. Diferente de linguagens convencionais, o LIA foi desenhado para ser:

- **Gerado por modelos**: Sintaxe e semântica otimizadas para LLMs
- **Verificável**: Determinismo e reprodutibilidade por construção
- **Linkável**: Módulos gerados em paralelo podem ser unidos via linker

**Sua missão como agente**: Contribuir com código, testes e documentação mantendo os princípios de reprodutibilidade e qualidade do projeto.

---

## 2. PRINCÍPIOS OBRIGATÓRIOS

### 2.1 Reprodutibilidade First

- **Toda geração de código** deve incluir um bloco `@gen` com metadados:
  - `prompt_ref`: Referência para o prompt usado
  - `prompt_hash`: Hash SHA-256 do prompt
  - `model_id`: Identificação do modelo (ex: `gpt-4o-mini`, `qwen2.5-coder:7b`)
  - `model_params`: Parâmetros como `temperature`, `top_p`, `seed`
  
- **Exemplo de bloco @gen em LIA**:

  ```lia
  @gen {
    prompt_ref: "p-001",
    prompt_hash: "a3f2b1c...",
    model_id: "qwen2.5-coder:7b",
    model_params: "temperature=0.1"
  }
  module orders.core as domain {
    // ...
  }
  ```

### 2.2 Documentação First

- **Antes de mudar o código**, atualize a especificação em `docs/spec/`.
- **Antes de adicionar uma feature**, documente no plano de implementação.
- **Use ADRs** (Architecture Decision Records) para decisões importantes.

### 2.3 Testes são Inegociáveis

- **Todo código Go deve ter testes unitários** (`*_test.go`).
- **Coverage mínimo**: 75% para novos pacotes.
- **Linker e Parser**: Devem manter >85% de cobertura.

---

## 3. WORKFLOW DE CONTRIBUIÇÃO

### 3.1 Branch Naming

- `feat/nome-descritivo`: Novas funcionalidades
- `fix/issue-id`: Correções de bugs
- `spec/proposta`: Mudanças na especificação LIA
- `docs/topico`: Melhorias em documentação

### 3.2 Commits Convencionais

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat(parser): add support for hole declarations`
- `fix(linker): resolve duplicate symbols correctly`
- `docs(guide): add examples for pack usage`
- `test(codec): increase coverage to 85%`

### 3.3 Pull Request Checklist

Antes de criar um PR, **verifique**:

- [ ] A especificação em `docs/spec/` está atualizada (se aplicável)
- [ ] Testes unitários foram adicionados/atualizados
- [ ] `go test ./internal/... -cover` passa sem erros
- [ ] Se mudou Parser/Linker, atualizou exemplo em `examples/`
- [ ] Se for geração de IA, incluiu bloco `@gen` ou metadados no commit

---

## 4. ESTRUTURA DE CÓDIGO GO

### 4.1 Organização de Pacotes

```
internal/
  parser/     → Parsing de .lia para IR
  linker/     → Resolução de símbolos e linking
  codec/      → Serialização canônica e hashing
  llmgen/     → Integração com LLMs (Ollama, etc.)
  policy/     → Enforcement de constraints
  repro/      → Metadados de reprodutibilidade
```

### 4.2 Padrões de Código

- **Siga [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)**
- **Erros tipados**: Use `%w` para wrapping
- **Exports mínimos**: Mantenha interfaces pequenas
- **Canonicalização**: Use `codec.SortAll()` antes de serializar

### 4.3 Testes

```go
// Nomenclatura: Test<FunctionName>_<Scenario>
func TestParseFile_BasicProject(t *testing.T) {
    // Arrange
    content := `project TestProj { ... }`
    tmpfile := createTempFile(t, content)
    
    // Act
    prog, err := ParseFile(tmpfile)
    
    // Assert
    if err != nil {
        t.Fatalf("ParseFile failed: %v", err)
    }
    // ... demais assertions
}
```

---

## 5. GERAÇÃO DE CÓDIGO LIA

### 5.1 Uso do LLMGen

Se você for **usar o LIA para gerar LIA** (meta-geração):

```go
provider := llmgen.NewOllamaProvider("http://localhost:11434")
gen := llmgen.NewGenerator(provider, "./prompt-tape.json")

module, meta, err := gen.GenerateLIAModule(ctx, llmgen.ModuleSpec{
    Name:        "user.repository",
    Role:        "adapter",
    Model:       "qwen2.5-coder:7b",
    Temperature: 0.1,
    Context:     "Database access layer for User entity",
})
```

### 5.2 Prompts Estruturados

Ao criar prompts para gerar LIA:

- **Seja explícito sobre roles**: `"Create a domain module (no IO effects)"`
- **Inclua constraints**: `"Must not depend on adapters"`
- **Forneça contexto**: Packs usados, estilos de arquitetura

### 5.3 Validação Pós-Geração

Após gerar código LIA:

1. **Parse**: `lia parse output.lia -o output.liao`
2. **Check**: `lia check output.liao`
3. **Link**: `lia link output.liao -o linked.lial`
4. **Verificar hash**: O hash do `.lial` deve ser determinístico

---

## 6. CONSTRAINTS E BOAS PRÁTICAS

### 6.1 Domain-Driven Design (via Packs)

O LIA não impõe Clean/Hexagonal, mas os **packs** podem:

- `HexCore@1.0.0`: Impõe regras hexagonais via policies
- `LayeredArch@1.0.0`: Impõe camadas tradicionais

**Como agente, respeite as policies do projeto**:

```bash
lia check mycode.liao --pack-dir ./docs/spec/packs
```

### 6.2 Evite Over-Engineering

- **Budgets**: Alguns projetos têm limites de camadas/módulos.
- **Simplicidade**: Se o projeto é marcado como "simple", não gere arquitetura complexa.

### 6.3 Nomeação Canônica

- **Módulos**: `snake_case` ou `dotted.hierarchy` (ex: `orders.core`)
- **Tipos**: `PascalCase` (ex: `OrderId`, `PaymentStatus`)
- **Holes**: `hole_<description>` (ex: `hole_user_repository`)

---

## 7. INTEGRAÇÃO COM OLLAMA

Se você for um agente **executando localmente** e precisar testar geração:

### 7.1 Setup

```bash
# Instalar Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Baixar modelo de código
ollama pull qwen2.5-coder:7b
```

### 7.2 Configuração

Crie `.env.local`:

```bash
LIA_LLM_PROVIDER=ollama
LIA_LLM_BASE_URL=http://localhost:11434
LIA_LLM_MODEL=qwen2.5-coder:7b
LIA_LLM_TEMPERATURE=0.1
```

### 7.3 Testes

```bash
# Rode testes com mock (rápido, sem LLM)
go test ./internal/llmgen -v

# Rode testes com LLM real (requer Ollama rodando)
go test ./internal/llmgen -v -tags=llm_integration
```

---

## 8. CHECKLIST PRÉ-COMMIT (AGENTES)

Antes de submeter qualquer mudança, **verifique**:

- [ ] O código compila: `go build ./...`
- [ ] Testes passam: `go test ./internal/... -cover`
- [ ] Cobertura não diminuiu (use `go tool cover`)
- [ ] Documentação atualizada (se aplicável)
- [ ] Bloco `@gen` presente em artefatos gerados por IA
- [ ] Commit message segue Conventional Commits
- [ ] Nenhuma informação sensível (API keys, secrets) no código

---

## 9. RECURSOS ADICIONAIS

- **Especificação LIA**: `docs/spec/lia-v0.1.md`
- **Manifesto**: `docs/LIA_MANIFESTO.md`
- **Guia de Testes**: `docs/TESTING_GUIDE.md`
- **Estratégia de Testes**: `docs/TESTING_STRATEGY.md`
- **Contribuição**: `docs/CONTRIBUTING.md`

---

## 10. FILOSOFIA MODEL-FIRST

**Lembre-se**: O LIA não é uma linguagem para humanos escreverem à mão. Ele é desenhado para:

1. **Modelos gerarem** (você!)
2. **Máquinas validarem** (linker, checker)
3. **Humanos auditarem** (via decision logs e prompt tapes)

Sua contribuição como agente é **gerar artefatos verificáveis e reprodutíveis**, não código "bonito" para leitura humana.

---

**Última Atualização**: 2026-01-31  
**Versão do LIA**: 0.1 (pré-alpha)
