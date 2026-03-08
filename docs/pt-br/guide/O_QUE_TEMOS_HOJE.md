# O Que Temos Hoje

Este texto existe para responder uma pergunta simples:

**afinal, o que a LIA já é hoje, de verdade?**

A melhor forma de entender é esta:

## Em uma frase

Hoje a LIA já funciona como um **trilho intermediário entre um pedido em linguagem natural e um projeto Java compilável**, com mais estrutura, rastreabilidade e validação do que pedir código direto para a IA.

Ela ainda não é uma plataforma completa de geração de software pronta para qualquer cenário, mas já deixou de ser só ideia.

## O problema que estamos tentando resolver

Quando pedimos código direto para uma IA, normalmente ganhamos algo assim:

- uma resposta opaca
- difícil de auditar
- difícil de reproduzir
- fácil de quebrar por pequenos desvios
- sem uma etapa formal entre “o que eu pedi” e “o código final”

A proposta da LIA é mudar isso.

Em vez de ir direto de:

**prompt -> Java**

a LIA introduz um caminho mais controlado:

**prompt -> plano interno -> LIA -> validação -> Java**

Esse “meio do caminho” é justamente o diferencial.

## O que já existe de fato

Hoje já temos cinco coisas concretas.

### 1. Uma linguagem intermediária de verdade

A LIA já consegue representar um projeto em termos de:

- domínio
- portas/contratos
- casos de uso
- adapters
- wiring
- packs e restrições
- metadados de geração

Ou seja: ela já consegue expressar a arquitetura antes do código final.

### 2. Um pipeline que valida essa representação

Não é só “gerar um texto parecido com LIA”.

Hoje o toolchain já:

- parseia
- valida
- linka símbolos
- registra decision log
- registra prompt tape
- faz replay

Isso significa que a representação intermediária já passa por checagens antes de virar código final.

### 3. Lower real para Java

O projeto já não para mais no `.lial`.

Hoje já existe lowering real para Java, gerando:

- múltiplos arquivos
- estrutura de projeto
- classes, records, interfaces e wiring
- saída compilável
- perfis diferentes para o mesmo LIA: `plain`, `spring-boot` e `quarkus`

Então a LIA já consegue chegar a um projeto Java concreto.

### 4. Entrada por prompt livre

Esse é um avanço importante.

No começo, usamos `project-spec.json` como forma de controlar a demo e comparar cenários com mais justiça.

Mas esse não é mais o modelo ideal de uso.

Hoje já existe um caminho por prompt livre:

- o usuário descreve o que quer
- o sistema lê o contexto do workspace
- o sistema consulta os packs disponíveis
- a IA gera um plano interno
- esse plano vira LIA
- a LIA vira Java

Então o `project-spec` deixou de ser a única entrada possível. Ele passa a fazer mais sentido como artefato interno ou como modo explícito de controle, não como input obrigatório do usuário final.

### 5. Packs começaram a virar parte do sistema

Esse ponto é fundamental.

Se os packs forem só “arquivos que existem no disco”, a IA não sabe o que eles significam.

Por isso agora os packs já têm manifests legíveis pela máquina. Na prática isso quer dizer:

- o ambiente pode dizer quais packs existem
- a IA pode inferir packs a partir do prompt e do contexto
- `ArchBaseline` pode funcionar como built-in
- packs externos deixam de depender de “a IA adivinhar”

Esse passo é importante porque aproxima a LIA de algo nativo do produto, e afasta a solução de virar apenas uma coleção de prompts ou skills.

## Como o fluxo funciona hoje, sem detalhe técnico demais

Se eu resumisse o que acontece hoje para alguém numa call, eu diria assim:

1. a pessoa descreve o que quer construir
2. o sistema observa o contexto do workspace
3. o sistema olha o catálogo de packs
4. a IA monta um plano interno do projeto
5. a IA gera módulos LIA com papéis explícitos
6. o toolchain valida essa estrutura
7. o projeto é baixado para Java
8. todo o processo fica registrado

O ponto central é:

**a IA não está “mais inteligente”; ela está sendo obrigada a passar por um trilho mais verificável**

## O que já conseguimos provar

Hoje já conseguimos provar três coisas importantes.

### A. A LIA não é só um parser

Ela já tem geração, validação, replay e lowering.

### B. O caminho LIA já produz um resultado melhor que o pedido direto em Java em um caso real

Na demo comparativa que montamos:

- existe um Java de referência feito manualmente
- existe um ramo “Java direto”
- existe um ramo “LIA -> Java”

E o resultado concreto foi:

- o Java de referência compila
- o Java direto falha
- o ramo LIA -> Java compila

Isso não prova que a LIA resolve tudo.

Mas prova que a etapa intermediária melhora o processo.

### C. Já existe um caminho de prompt livre até Java

Esse ponto é importante porque muda a conversa.

Antes, alguém podia dizer:

“isso ainda depende de um arquivo estruturado que um humano precisou escrever”

Agora essa crítica enfraquece, porque já existe o planner interno por prompt livre.

## O que ainda não está pronto

Também é importante falar com honestidade.

Ainda não estamos em “geração geral de software pronta para produção”.

Ainda faltam, por exemplo:

- typechecker semântico mais forte
- inferência mais rica de packs de comunidade
- lowering completo para outros targets além de Java
- resolução mais sofisticada de lacunas semânticas
- runtime e convenções alvo mais completos

Ou seja:

o sistema já funciona, mas ainda está no estágio de **produto em formação**, não de solução universal acabada.

## O que é o diferencial real em relação a skills e agents.md

Esse é um ponto estratégico.

Skills, prompts estruturados e `agents.md` ajudam a IA a se comportar melhor.

Mas normalmente eles:

- não viram artefato normativo
- não entram no linker
- não entram no replay
- não são validados como parte do pipeline
- não viram uma linguagem intermediária com semântica própria

A vantagem da LIA só existe se ela continuar sendo isso:

**uma camada formal entre intenção e código**

Se ela virar apenas “mais um jeito de escrever instruções”, perde força.

## Como eu explicaria o estado atual para outra pessoa

Se eu tivesse que resumir para alguém do time, eu diria:

> Hoje a LIA já é um pipeline real que recebe um pedido, organiza a geração em uma representação intermediária arquitetural, valida essa representação e consegue produzir Java compilável.  
>  
> Ainda não é a solução final para qualquer projeto, mas já prova que existe uma vantagem concreta em passar por uma linguagem intermediária, em vez de pedir código direto para a IA.

## Resumo final

Hoje, a LIA já é:

- uma linguagem intermediária funcional
- um pipeline validável
- um gerador com replay e trilha
- um lower real para Java
- um sistema que já começa a entender packs e contexto de workspace
- uma base concreta para uma geração mais determinística

Hoje, a LIA ainda não é:

- um gerador universal pronto para produção
- um substituto completo de engenharia humana
- uma solução sem gaps de semântica e modelagem

Mas a diferença importante é esta:

**ela já saiu do campo da intenção e entrou no campo do que pode ser demonstrado**
