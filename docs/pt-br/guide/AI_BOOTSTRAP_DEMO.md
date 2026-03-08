# Demo: Prompt Direto vs LIA

Este guia existe para mostrar o ponto central da LIA: o ganho não é “usar IA”, e sim obrigar a IA a gerar através de uma estrutura intermediária validável.

## Antes: pedir Java direto

Exemplo de prompt:

```text
Crie um microsserviço Java para pedidos com casos de uso, repositório, adapter HTTP e persistência.
```

Problemas:

- o output já nasce acoplado ao target
- a arquitetura fica implícita no prompt
- não existe artefato intermediário para validar
- não existe `decision-log`
- não existe `prompt-tape` ligado ao código final

## Depois: pedir um projeto LIA

Você fornece um `project-spec.json` e deixa a IA gerar módulos LIA.

Exemplo de spec:

- [project-spec.json](/home/willams/LIA/lia/examples/ai-bootstrap/project-spec.json)

Com Ollama:

```bash
go run ./cmd/lia gen project \
  --spec ./examples/ai-bootstrap/project-spec.json \
  --provider ollama \
  --model qwen2.5-coder:7b \
  --base-url http://localhost:11434 \
  --out-dir /tmp/ai-bootstrap \
  --pack-dir ./docs/pt-br/spec/packs
```

Com API OpenAI-compatible:

```bash
go run ./cmd/lia gen project \
  --spec ./examples/ai-bootstrap/project-spec.json \
  --provider openai-compatible \
  --base-url http://localhost:8000 \
  --api-key-env OPENAI_API_KEY \
  --model gpt-4o-mini \
  --out-dir /tmp/ai-bootstrap \
  --pack-dir ./docs/pt-br/spec/packs
```

## O que o comando gera

- `/tmp/ai-bootstrap/project.lia`
- `/tmp/ai-bootstrap/project.liao`
- `/tmp/ai-bootstrap/project.lial`
- `/tmp/ai-bootstrap/project.lial.decision-log.json`
- `/tmp/ai-bootstrap/prompt-tape.json`

Além disso, o comando já executa:

- parse
- check
- link
- replay

## O que fica concretamente diferente

### Saída direta em Java

- uma resposta só
- pouca separação entre domínio, contrato e implementação
- difícil comparar duas gerações semanticamente
- pouca auditabilidade

### Saída em LIA

- módulos explícitos por papel (`domain`, `port`, `usecase`)
- `@gen` por módulo
- `prompt-tape.json` com hashes
- `project.liao` e `project.lial` canônicos
- `decision-log.json` mostrando escolhas do linker
- `replay` validando se os prompts batem com o artefato

## Como vender isso na demo

A mensagem certa não é “a IA gerou código”.

A mensagem é:

1. a IA gerou uma representação intermediária com papéis arquiteturais explícitos
2. o toolchain validou essa representação antes de qualquer lowering
3. o build final ficou auditável, replayable e comparável

## Limite atual

O comando já cria um projeto LIA do zero a partir de uma spec estruturada e agora já existe lowering real para Java.

Então hoje a demonstração correta pode ser:

- **prompt/spec -> projeto LIA validado**
- **prompt/spec -> LIA -> projeto Java compilável**

Ainda não:

- **prompt/spec -> aplicação Java final pronta para produção com typechecker semântico, runtime e policies completas**
