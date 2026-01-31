# Visão Ger

al do LIA

LIA (Linguagem Intermediária Assistida) é uma **representação intermedi ária model-first** e um **toolchain** para gerar software de forma segura e reprodutível.

Pense nele como duas coisas trabalhando juntas:

1) **Uma linguagem/IR** que modela a estrutura do software em um formato determinístico e linkável.
2) **Um sistema** (CLI + passes + linker + reprodutibilidade) que transforma a saída de LLM em artefatos auditáveis.

## Por que existe

A geração direta "prompt → código" é frágil:

- é difícil manter a arquitetura consistente entre módulos,
- é fácil quebrar fronteiras ou misturar camadas,
- é difícil reproduzir ou auditar decisões,
- geração paralela não é confiável sem interfaces estáveis.

O LIA move essas preocupações para um pipeline semelhante a um compilador:

- arquivos objeto estruturados (`.liao`),
- verificação (constraints/policies),
- linking (seleção/merge determinístico),
- reprodutibilidade (prompt tape + hashes + decision logs).

## O que é

- Um IR determinístico com JSON canônico.
- Um toolchain que pode parsear, validar, linkar e fazer lowering de artefatos.
- Uma base para perfis de arquitetura (hexagonal, em camadas, etc.) como **packs/policies**, não dogma hardcoded.

## O que ainda não é

- Uma linguagem de programação completa com sintaxe completa.
- Um compilador completo para binários nativos.
- Um motor de segurança completo (políticas ainda são parcialmente enforçadas).

## Princípios principais

- **Reprodutibilidade primeiro:** hashes e proveniência são obrigatórios.
- **Determinismo por construção:** representação canônica, ordenação estável.
- **Modularidade:** arquivos objeto se linkam de forma segura.
- **Arquitetura orientada a políticas:** regras são código, não texto.

Se você quer um MVP executável de 10 minutos, vá para `docs/guide/getting-started.md`.
