# Roadmap / Status

Este documento descreve o que existe hoje versus o que está planejado.

## Existe (MVP)

- CLI com parse/check/link/explain/lower
- Estruturas IR + JSON canônico + hashing
- Carregador de packs (pseudo-LIA) e motor de políticas mínimo
- Scaffolding de reprodutibilidade (`@gen`, prompt tape)
- Projeto de exemplo e prompt tape

## Parcialmente implementado

- Enforcement de políticas (apenas subconjunto)
- Linker (stub, sem resolução de grafo)
- Parser (subconjunto muito pequeno)
- Lowerers (stubs)

## Ainda não implementado

- Gramática completa e parser
- Tabela de símbolos completa e `requires/provides`
- Seleção com scoring + tie-break
- Passes de synth/transform
- Capabilities e taint/dataflow
- Lowering real para Java/Python
- Registry de packs / resolução de semver

## Próximos marcos imediatos

1) Parse de `provides/requires` + efeitos
2) Construir grafo de dependências e enforçar deps de roles
3) Implementar seleção determinística + detalhe do decision log
4) Adicionar primeiro target de lowering real
