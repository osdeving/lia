# Guia de Execução de Testes LIA

## Início Rápido

### Executar Todos os Testes Unitários

```bash
go test ./internal/... -v -cover
```

### Gerar Relatório de Cobertura

```bash
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Executar Testes de Pacotes Específicos

```bash
# Testes do Parser
go test ./internal/parser -v

# Testes do Linker
go test ./internal/linker -v

# Testes do Codec
go test ./internal/codec -v

# Testes do Gerador LLM (apenas mock, sem LLM real)
go test ./internal/llmgen -v
```

## Testes de Integração com LLM

### Pré-requisitos

1. Instalar Ollama: <https://ollama.ai>
2. Baixar um modelo de código:

```bash
ollama pull qwen2.5-coder:7b
# ou
ollama pull deepseek-coder:6.7b
```

### Executar Testes de Integração LLM

```bash
# Definir variáveis de ambiente
export LIA_LLM_PROVIDER=ollama
export LIA_LLM_BASE_URL=http://localhost:11434
export LIA_LLM_MODEL=qwen2.5-coder:7b

# Executar com tag de integração (quando implementado)
go test ./internal/llmgen -v -tags=llm_integration
```

### Testes com Docker

```bash
# Construir imagem de teste
docker build -t lia-test -f Dockerfile .

# Executar testes no container
docker run --rm lia-test go test ./internal/... -v -cover

# Executar com acesso ao Ollama (modo host network)
docker run --rm --network=host lia-test go test ./internal/llmgen -v
```

## Status Atual de Cobertura

| Pacote | Cobertura | Status |
|---------|----------|--------|
| parser | 96.4% | ✅ Excelente |
| linker | 78.4% | ✅ Bom |
| codec | 43.9% | ⚠️ Precisa melhorias |
| llmgen | 16.5% | ⚠️ Cobertura básica |
| check | 0.0% | ❌ Sem testes |
| cli | 0.0% | ❌ Sem testes |
| policy | 0.0% | ❌ Sem testes |
| repro | 0.0% | ❌ Sem testes |

## Próximos Passos

1. Adicionar testes para o pacote `policy` (enforcement de constraints)
2. Adicionar testes para o pacote `repro` (metadados de reprodutibilidade)
3. Adicionar testes de integração end-to-end
4. Implementar testes reais de integração LLM com Ollama
