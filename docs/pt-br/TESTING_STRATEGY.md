# Estratégia de Testes LIA & Integração com LLM

## Visão Geral

Este documento descreve a estratégia abrangente de testes para o toolchain LIA e a integração com LLMs locais para testes de geração.

## Metas de Cobertura

Use relatórios de cobertura para acompanhar progresso e evitar regressões:

```bash
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Metas (ver `AGENT.md`):

- Parser: >85%
- Linker: >85%
- Novos pacotes: >75%

## Arquitetura de Testes

### 1. Testes Unitários

Localizados em cada pacote `internal/*` como arquivos `*_test.go`.

**Áreas de Cobertura**:

- Parser: Parsing de arquivos, referências de packs, casos de erro
- Linker: Merging de programas, logging de decisões, determinismo de hash
- Codec: Operações de leitura/escrita, serialização canônica
- LLMGen: Requisições de geração, construção de prompts, rastreamento de metadados

### 2. Testes de Integração

Localizados em `tests/integration/` (a ser criado).

**Cenários de Teste**:

- End-to-end: `.lia` → parse → check → link → `.lial`
- Enforcement de políticas através de múltiplos módulos
- Reprodutibilidade: Mesma entrada + mesma config = mesmo hash

### 3. Testes de Geração LLM

Localizados em `tests/llm_integration/` (a ser criado).

**Tipos de Teste**:

- **Testes de Smoke**: Verificar conectividade LLM e geração básica
- **Testes de Reprodutibilidade**: Mesmo prompt → comparar hashes
- **Testes de Qualidade**: Validar sintaxe .lia gerada

## Arquitetura de Integração LLM

### Provedores Suportados

#### 1. Ollama (Principal - Local)

- **Endpoint**: `http://localhost:11434`
- **Modelos**: qwen2.5-coder, codellama, deepseek-coder
- **Uso**: Principal provedor de testes para desenvolvimento offline

#### 2. APIs Compatíveis com OpenAI

- **Endpoint**: Configurável (LMStudio, LocalAI, etc.)
- **Uso**: Provedores locais alternativos

### Configuração

Variáveis de ambiente para configuração de testes:

```bash
export LIA_LLM_PROVIDER=ollama           # ou "openai-compatible"
export LIA_LLM_BASE_URL=http://localhost:11434
export LIA_LLM_MODEL=qwen2.5-coder:7b
export LIA_LLM_TEMPERATURE=0.1
```

## Executando Testes

### Executando Todos os Testes

```bash
# Executar todos os testes unitários com cobertura
go test ./internal/... -v -cover

# Gerar relatório de cobertura
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Executando Testes de Integração LLM

```bash
# Requer Ollama rodando localmente
ollama serve &

# Executar testes de geração LLM
go test ./internal/llmgen -v -tags=llm_integration

# Executar com modelo específico
LIA_LLM_MODEL=deepseek-coder:6.7b go test ./internal/llmgen -v
```

### Testes Baseados em Docker

```bash
# Executar testes dentro do container Docker
docker build -t lia-test -f Dockerfile.test .
docker run --rm lia-test

# Executar com acesso ao Ollama (modo network host)
docker run --rm --network=host lia-test go test ./internal/llmgen -v
```

## Próximos Passos

1. **Fuzz Testing**: Usar `go-fuzz` para robustez do parser
2. **Property-Based Testing**: Usar `gopter` para invariantes do linker
3. **Benchmark Tests**: Rastrear regressão de performance
4. **Multi-Model Testing**: Testar através de Ollama, LMStudio e LLMs em nuvem
