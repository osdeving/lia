# LIA-WS — Linguagem Intermediária Assistida (Weaver Stack)

> **Status:** Draft v0.1 (opiniões + alternativas).
> **Objetivo:** uma linguagem/IR *model-first* para geração confiável de software com **artefatos linkáveis**, **policies/constraints** (NFR), **requisitos/aceitação** (FR) e **reprodutibilidade** (tapes + hashes), suportando **elicitation** (perguntas) durante build/link.

---

## 0) Princípios

1. **Determinismo por padrão**: mesma entrada (incluindo decisions) ⇒ mesmo `.lial` e mesma saída transpillada.
2. **Fail-closed**: se violar policy/constraint, **não linka**.
3. **Model-first, human-usable**: humanos conseguem escrever o essencial (policies, requisitos, packs), LLMs completam.
4. **Modularidade**: tudo se expressa como unidades linkáveis (`.liao`).
5. **Rastreabilidade**: requisito → símbolos → testes → políticas → decisões (matriz canônica).

---

## 1) Artefatos e fases (C-like)

* `*.lia`  — source (texto)
* `*.liao` — object file semântico (IR canônica + símbolos + metadados)
* `*.lial` — linked unit (IR final linkada)
* `*.liap` — pack (stdlib/plugins/policies versionados)

Pipeline:

1. `lia parse`: `.lia → AST → IR → .liao`
2. `lia check`: valida IR + constraints locais
3. `lia link`: resolve `requires/provides`, aplica policies, emite `.lial` + decision-log + traceability
4. `lia lower/transpile`: `.lial → target` (Java/Python/etc.)
5. `lia replay`: reexecuta build com prompt/decision tape

---

## 2) Decisões de design (com alternativas)

### 2.1 Estilo de sintaxe

**Opções**

* **(A) C-like com chaves** `{}` e `;` opcionais: familiar, fácil de parsear, bom para snippets em chat.
* (B) Indentation-based (Python-like): mais limpo, mas editor/tooling fica mais delicado.
* (C) S-expression (Lisp): AST-friendly, mas baixa adoção para humanos.
* (D) YAML/JSON DSL: bom para dados, ruim para expressar código/escopo com clareza.

**Escolha recomendada:** **(A) C-like com chaves** (mais “default” para devs e para geração por LLM).

### 2.2 Comentários e metadados

**Opções**

* (A) Anotações `@tag(...)` (estilo Java/Swift)
* (B) Atributos `#[tag]` (estilo Rust)
* (C) Comentários estruturados `///` (estilo doc)

**Escolha recomendada:** **(A) `@` annotations** + `///` para docs. Anotações são parseáveis e fáceis de impor.

### 2.3 Keywords: “port/usecase/adapter” vs termos genéricos

**Opções**

* (A) Vocabulário arquitetural explícito: `port`, `usecase`, `adapter`, `wiring`, `domain`.
* (B) Vocabulário genérico: `interface`, `service`, `impl`, `module`.

**Escolha recomendada:** **(A)** porque o core do valor é “organização por construção”.

### 2.4 Onde colocar políticas (NFR)

**Opções**

* (A) `policy` dentro do `.lia` (primeira classe)
* (B) políticas em arquivo separado (ex.: Rego/CUE only)

**Escolha recomendada:** **(A)** como *primeira classe*, mas com “backends” de execução (OPA/CUE) opcionais.

### 2.5 Requisitos (FR) — Gherkin completo vs BDD-light

**Opções**

* (A) Exigir `.feature` Gherkin
* (B) BDD-light no `.lia`: `story/rule/scenario` com exemplos

**Escolha recomendada:** **(B)**: humano escreve NL simples, LIA/LLM normaliza para FR estruturado.

---

## 3) Léxico

### 3.1 Identificadores

* `Ident` = letra ou `_` seguido de letras/dígitos/`_`
* `QName` = `Ident ('.' Ident)*` (ex.: `orders.core`)

### 3.2 Literais

* `String` com aspas duplas
* `Int`, `Float`, `Bool`

### 3.3 Comentários

* `//` linha
* `/* ... */` bloco
* `///` doc

### 3.4 Palavras-chave (v0.1)

**Core**: `package`, `import`, `pack`, `module`, `role`, `domain`, `type`, `enum`, `struct`, `fn`, `port`, `usecase`, `adapter`, `wiring`, `provides`, `requires`, `policy`, `constraint`, `prefer`, `capability`, `effect`, `test`, `story`, `rule`, `scenario`, `examples`, `question`, `decision`, `assumption`.

**Alternativas (não escolhidas agora)**:

* `component` no lugar de `module`
* `contract` no lugar de `port`
* `service` no lugar de `usecase`

---

## 4) Modelo semântico (capabilities)

### 4.1 Roles (organização)

Roles padrão (packs podem estender):

* `domain` (puro)
* `port` (contrato)
* `usecase` (orquestra)
* `adapter` (I/O: DB, HTTP, MQ)
* `wiring` (composição)
* `policy` (NFR)
* `test` (oráculos)

### 4.2 Effects (efeitos)

Efeitos padronizam “o que esse código pode fazer”:

* `pure` (sem I/O)
* `io` (I/O externo)
* `tx` (transação)
* `emit` (eventos)

Regra típica:

* `domain` deve ser `pure`.

### 4.3 Capabilities (segurança por construção)

Capabilities representam permissões explícitas (NFR/security):

* `db.query`, `net.http`, `mq.publish`, `secrets.read`, `fs.write`, `crypto.sign`...

Regras típicas:

* `domain` não pode declarar/usar capabilities.
* `adapter` pode, mas deve declarar quais.

### 4.4 Taint (opcional v0.2)

Tipo de confiança para prevenir injection:

* `Tainted<String>` vs `Trusted<String>`
* sinks exigem `Trusted<>`.

---

## 5) FR vs NFR na linguagem

### 5.1 FR — requisitos e aceitação (BDD-light)

* `story`: “Como usuário…”
* `rule`: regra declarativa de negócio
* `scenario`: exemplo verificável
* `examples`: tabela opcional

Derivações esperadas (via passes):

* `usecase` + invariantes de domínio
* `test.case` / `test.property`

### 5.2 NFR — policies/constraints

* `policy`: conjunto de regras (permit/deny) aplicadas no build/link
* `constraint`: invariantes estruturais (dependências/camadas)
* `prefer`: heurísticas/score quando há múltiplas opções válidas

---

## 6) Elicitation (perguntas) como primeira classe

Objetivo: evitar que o sistema “invente” decisões de negócio.

### 6.1 Artefatos

* `question`: pergunta estruturada com schema (opções/valores)
* `decision`: resposta aceita (entra no decision tape)
* `assumption`: default assumido (permitido apenas em perfis permissivos)

### 6.2 Regra de build

Em `repro=strict`:

* nenhuma `question` pode permanecer `open` no `.lial`.
* toda `decision` deve ser registrada no `decision_tape`.

### 6.3 Como o linker usa elicitation

Quando faltar informação para satisfazer FR ou para resolver conflito de implementação:

1. linker emite `question` (com ID estável)
2. cliente responde
3. linker aplica `decision` e continua
4. decision log registra “por que perguntou”

---

## 7) Packs (stdlib/arquitetura/segurança)

* `pack` define módulos reutilizáveis e policies versionadas.
* `import pack <name>@<semver>`

Exemplos de packs:

* `ws-arch-baseline@1.0.0`
* `ws-security-baseline@1.0.0`

---

## 8) Linker: resolução e determinismo

### 8.1 provides / requires

* `provides`: símbolos exportados
* `requires`: dependências declaradas

### 8.2 Conflitos

* múltiplos providers para o mesmo símbolo
* mismatch de tipos/versions
* violações de policy

### 8.3 Seleção (determinística)

1. filtrar por constraints (hard)
2. aplicar policies (deny/allow)
3. computar score (prefer)
4. tie-break fixo (lexical + semver + menor deps)

Outputs:

* `.lial`
* `decision-log.json`
* `traceability.json`

---

## 9) AST (modelo lógico) — draft

> Este AST é “alto nível” e representa nós parseáveis. O IR (`.liao`) pode normalizar e enriquecer com símbolos e tipos.

### 9.1 Nós principais

* `Program`

  * `PackageDecl`
  * `Imports[]`
  * `Packs[]`
  * `Decls[]`

* `ModuleDecl`

  * `Name: QName`
  * `Role: Ident`
  * `Annotations[]`
  * `Decls[]` (types, ports, usecases, adapters, wiring, policies)

* `TypeDecl` (`struct|enum`)

* `PortDecl`

* `UsecaseDecl`

* `AdapterDecl`

* `WiringDecl`

* `PolicyDecl`

* `ConstraintDecl`

* `PreferDecl`

* `RequirementDecl` (`story|rule|scenario|examples`)

* `QuestionDecl` / `DecisionDecl` / `AssumptionDecl`

* `TestDecl`

### 9.2 Anotações

* `@gen(...)`: provenance/reprodutibilidade
* `@rtf(...)`: role-task-format (mini prompt)
* `@cap(...)`: capabilities
* `@trace(...)`: vínculos com requisitos

---

## 10) Gramática EBNF (v0.1, C-like escolhido)

> **Nota:** esta gramática é um recorte inicial para toolchain. Recursos avançados (taint, generics, macros) ficam para v0.2+.

### 10.1 Tokens básicos

```
Ident      = letter , { letter | digit | '_' } ;
QName      = Ident , { '.' , Ident } ;
String     = '"' , { char } , '"' ;
Int        = digit , { digit } ;
Bool       = 'true' | 'false' ;
```

### 10.2 Programa

```
Program     = PackageDecl , { ImportDecl } , { PackImportDecl } , { TopDecl } ;
PackageDecl = 'package' , QName , ';' ;
ImportDecl  = 'import' , String , ';' ;
PackImportDecl = 'import' , 'pack' , Ident , '@' , Version , ';' ;
Version     = Int , '.' , Int , '.' , Int ;
```

### 10.3 Declarações de topo

```
TopDecl   = ModuleDecl | PolicyDecl | RequirementBlock | TestDecl ;

ModuleDecl = { Annotation } , 'module' , QName , '{' , ModuleBody , '}' ;
ModuleBody = RoleDecl , { ModuleItem } ;
RoleDecl   = 'role' , Ident , ';' ;

ModuleItem = TypeDecl | PortDecl | UsecaseDecl | AdapterDecl | WiringDecl | ConstraintDecl | PreferDecl | RequirementBlock | TestDecl ;
```

### 10.4 Tipos

```
TypeDecl  = { Annotation } , ( StructDecl | EnumDecl ) ;
StructDecl = 'type' , Ident , 'struct' , '{' , { FieldDecl } , '}' ;
FieldDecl  = Ident , ':' , TypeRef , ';' ;
EnumDecl   = 'type' , Ident , 'enum' , '{' , Ident , { ',' , Ident } , '}' , ';' ;
TypeRef    = QName | Ident | 'String' | 'Int' | 'Bool' ;
```

### 10.5 Ports

```
PortDecl = { Annotation } , 'port' , Ident , '{' , { SigDecl } , '}' ;
SigDecl  = 'fn' , Ident , '(' , [ ParamList ] , ')' , [ ':' , TypeRef ] , ';' ;
ParamList = Param , { ',' , Param } ;
Param     = Ident , ':' , TypeRef ;
```

### 10.6 Use cases

```
UsecaseDecl = { Annotation } , 'usecase' , Ident , '{' , UsecaseBody , '}' ;
UsecaseBody = RequiresDecl , ProvidesDecl , [ EffectDecl ] , { UsecaseItem } ;
RequiresDecl = 'requires' , '{' , { SymbolRef , ';' } , '}' ;
ProvidesDecl = 'provides' , '{' , { SymbolRef , ';' } , '}' ;
EffectDecl   = 'effect' , ( 'pure' | 'io' | 'tx' | 'emit' ) , ';' ;
UsecaseItem  = 'step' , String , ';' | HoleDecl ;
HoleDecl     = 'hole' , Ident , ':' , TypeRef , ';' ;
SymbolRef    = QName | Ident ;
```

> `step` e `hole` são **model-first**: permitem descrever intenção sem implementar tudo em v0.1.

### 10.7 Adapters

```
AdapterDecl  = { Annotation } , 'adapter' , Ident , '{' , AdapterBody , '}' ;
AdapterBody  = RequiresDecl , ProvidesDecl , [ CapDecl ] , [ EffectDecl ] , { AdapterItem } ;
CapDecl      = 'capability' , '{' , { CapItem , ';' } , '}' ;
CapItem      = Ident | QName ;
AdapterItem  = 'bind' , Ident , '->' , QName , ';' | HoleDecl ;
```

### 10.8 Wiring

```
WiringDecl = { Annotation } , 'wiring' , Ident , '{' , { WireItem } , '}' ;
WireItem   = 'connect' , QName , '->' , QName , ';' | 'compose' , QName , ';' ;
```

### 10.9 Policies/Constraints/Preferências

```
PolicyDecl = { Annotation } , 'policy' , Ident , '{' , { PolicyRule } , '}' ;
PolicyRule = ( 'deny' | 'allow' | 'require' ) , QName , [ 'when' , Expr ] , ';' ;

ConstraintDecl = { Annotation } , 'constraint' , Ident , '{' , { ConstraintRule } , '}' ;
ConstraintRule = 'forbid' , QName , 'depends_on' , QName , ';'
               | 'require' , QName , 'depends_on' , QName , ';' ;

PreferDecl = { Annotation } , 'prefer' , Ident , '{' , { PreferRule } , '}' ;
PreferRule = 'score' , QName , ':' , Int , ';' ;

Expr = String ;  // v0.1: string expression (backend pode ser CEL/OPA)
```

> **Alternativa**: tornar `Expr` um CEL real no parser. Recomendação: começar string e evoluir.

### 10.10 FR (BDD-light)

```
RequirementBlock = StoryDecl | RuleDecl | ScenarioDecl ;

StoryDecl   = { Annotation } , 'story' , ReqId , '{' , 'as' , String , ';' , 'i_want' , String , ';' , 'so_that' , String , ';' , '}' ;
RuleDecl    = { Annotation } , 'rule' , ReqId , '{' , 'text' , String , ';' , '}' ;
ScenarioDecl= { Annotation } , 'scenario' , ReqId , '{' , { StepLine } , [ ExamplesDecl ] , '}' ;
StepLine    = ( 'given' | 'when' | 'then' | 'and' ) , String , ';' ;
ExamplesDecl= 'examples' , '{' , { ExampleRow } , '}' ;
ExampleRow  = String , ';' ;
ReqId       = Ident , '-' , Int ;
```

> **Alternativa**: aceitar Gherkin inteiro em arquivo separado. Recomendação: manter BDD-light dentro da LIA.

### 10.11 Elicitation

```
QuestionDecl  = { Annotation } , 'question' , ReqId , '{' , 'text' , String , ';' , [ 'options' , '{' , { String , ';' } , '}' ] , '}' ;
DecisionDecl  = { Annotation } , 'decision' , ReqId , '{' , 'answer' , String , ';' , '}' ;
AssumptionDecl= { Annotation } , 'assumption' , ReqId , '{' , 'value' , String , ';' , '}' ;
```

### 10.12 Testes (oráculos)

```
TestDecl    = { Annotation } , 'test' , Ident , '{' , { TestItem } , '}' ;
TestItem    = 'case' , ReqId , ';' | 'property' , ReqId , ';' | 'contract' , QName , ';' ;
```

---

## 11) Recomendações para o MVP (para começar a implementar)

1. Implementar apenas:

   * `package/import`
   * `module + role`
   * `type struct/enum`
   * `port` (assinaturas)
   * `usecase` com `requires/provides` + `hole`
   * `adapter` com `capability` + `bind` + `hole`
   * `policy/constraint` (Expr como string)
2. `.liao` como JSON canônico (JCS) + sha256
3. Linker resolvendo `requires/provides` + aplicando `constraint` hard
4. Elicitation apenas como artefato (`question/decision`) primeiro; depois via MCP

---

## 12) Exemplos mínimos

### 12.1 Módulo de domínio (puro)

```
module orders.core {
  role domain;

  type OrderId struct { value: String; }
  type OrderStatus enum { NEW, PAID, SHIPPED, CANCELED };
}
```

### 12.2 Port e use case com hole

```
module orders.ports {
  role port;
  port OrderRepo {
    fn get(id: OrderId): Order;
    fn save(o: Order): Bool;
  }
}

module orders.usecases {
  role usecase;

  usecase CancelOrder {
    requires { orders.ports.OrderRepo; }
    provides { orders.usecases.CancelOrder; }
    effect tx;

    step "load order";
    step "validate status";
    hole cancel_rule: String;
  }
}
```

### 12.3 Policy + elicitation

```
policy security_baseline {
  deny db.raw_sql;
  require authn when "endpoint.public";
}

question RULE-1 {
  text "Can a PAID order be canceled?";
  options { "yes"; "no"; }
}
```

---

## 13) Próximas decisões (v0.2)

* Expr: string vs CEL parseado
* Taint types (Tainted/Trusted)
* Generics
* Patch format para passes/plugins
* ABI de plugin (WASM) e sandbox
* LSP (tree-sitter) vs parser único
