package ir

// Program is the root unit for .lia/.liao/.lial artifacts.
type Program struct {
	Version  string    `json:"version"`
	Projects []Project `json:"projects,omitempty"`
	Packs    []Pack    `json:"packs,omitempty"`
	Modules  []Module  `json:"modules,omitempty"`
}

// Project represents a project block with policies and imports.
type Project struct {
	Name        string           `json:"name"`
	Repro       ReproProfile     `json:"repro,omitempty"`
	Tape        string           `json:"tape,omitempty"`
	Uses        []PackRef        `json:"uses,omitempty"`
	Policies    []PolicyDecl     `json:"policies,omitempty"`
	Constraints []ConstraintDecl `json:"constraints,omitempty"`
}

// Pack is a versioned collection of modules and policies.
type Pack struct {
	Name        string           `json:"name"`
	Version     string           `json:"version,omitempty"`
	Modules     []Module         `json:"modules,omitempty"`
	Policies    []PolicyDecl     `json:"policies,omitempty"`
	Constraints []ConstraintDecl `json:"constraints,omitempty"`
}

// PackRef references a pack by name/version.
type PackRef struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// Module represents a LIA module.
type Module struct {
	Name        string           `json:"name"`
	Role        string           `json:"role,omitempty"`
	Gen         *GenMeta         `json:"gen,omitempty"`
	Doc         *DocBlock        `json:"doc,omitempty"`
	Provides    []SymbolRef      `json:"provides,omitempty"`
	Requires    []SymbolRef      `json:"requires,omitempty"`
	Types       []TypeDecl       `json:"types,omitempty"`
	Enums       []EnumDecl       `json:"enums,omitempty"`
	Ports       []PortDecl       `json:"ports,omitempty"`
	Usecases    []UsecaseDecl    `json:"usecases,omitempty"`
	Adapters    []AdapterDecl    `json:"adapters,omitempty"`
	Wirings     []WiringDecl     `json:"wirings,omitempty"`
	Constraints []ConstraintDecl `json:"constraints,omitempty"`
	Preferences []PreferDecl     `json:"preferences,omitempty"`
	Holes       []HoleDecl       `json:"holes,omitempty"`
	Candidates  []CandidateDecl  `json:"candidates,omitempty"`
}

// GenMeta captures provenance for reproducibility.
type GenMeta struct {
	PromptRef      string   `json:"prompt_ref,omitempty"`
	PromptHash     string   `json:"prompt_hash,omitempty"`
	ModelID        string   `json:"model_id,omitempty"`
	ModelParams    []KV     `json:"model_params,omitempty"`
	ContextRefs    []string `json:"context_refs,omitempty"`
	ToolsTraceRefs []string `json:"tools_trace_refs,omitempty"`
	GeneratorPass  string   `json:"generator_pass,omitempty"`
	Timestamp      string   `json:"timestamp,omitempty"`
}

// DocBlock holds parsed doc and RTF metadata.
type DocBlock struct {
	Lines []string `json:"lines,omitempty"`
	RTF   *RTFMeta `json:"rtf,omitempty"`
}

// RTFMeta is a structured prompt capsule.
type RTFMeta struct {
	Role   string   `json:"role,omitempty"`
	Task   string   `json:"task,omitempty"`
	Format string   `json:"format,omitempty"`
	Inputs []string `json:"inputs,omitempty"`
	Must   []string `json:"must,omitempty"`
	Avoid  []string `json:"avoid,omitempty"`
	Tests  []string `json:"tests,omitempty"`
	Repair []string `json:"repair,omitempty"`
	Merge  string   `json:"merge,omitempty"`
}

// KV is a deterministic key-value pair used for canonical JSON.
type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ReproProfile defines reproducibility modes.
type ReproProfile string

const (
	ReproStrict     ReproProfile = "strict"
	ReproPinned     ReproProfile = "pinned"
	ReproBestEffort ReproProfile = "best_effort"
)

// SymbolRef references a symbol by qualified name.
type SymbolRef struct {
	QName string `json:"qname"`
	Kind  string `json:"kind,omitempty"`
}

// TypeDecl defines a type alias or newtype.
type TypeDecl struct {
	Name      string `json:"name"`
	Base      string `json:"base,omitempty"`
	Predicate string `json:"predicate,omitempty"`
}

// EnumDecl defines an enum.
type EnumDecl struct {
	Name   string   `json:"name"`
	Values []string `json:"values,omitempty"`
}

// PortDecl defines a port contract.
type PortDecl struct {
	Name    string     `json:"name"`
	Methods []FuncDecl `json:"methods,omitempty"`
}

// UsecaseDecl defines a use case contract.
type UsecaseDecl struct {
	Name    string   `json:"name"`
	Inputs  []Field  `json:"inputs,omitempty"`
	Outputs []Field  `json:"outputs,omitempty"`
	Effects []string `json:"effects,omitempty"`
	Body    []Stmt   `json:"body,omitempty"`
}

// AdapterDecl defines an adapter.
type AdapterDecl struct {
	Name       string   `json:"name"`
	Implements string   `json:"implements,omitempty"`
	Inputs     []Field  `json:"inputs,omitempty"`
	Outputs    []Field  `json:"outputs,omitempty"`
	Effects    []string `json:"effects,omitempty"`
	Body       []Stmt   `json:"body,omitempty"`
}

// WiringDecl defines a wiring block.
type WiringDecl struct {
	Name  string   `json:"name"`
	Binds []string `json:"binds,omitempty"`
}

// ConstraintDecl defines a hard constraint expression.
type ConstraintDecl struct {
	Name string `json:"name"`
	Expr string `json:"expr"`
}

// PreferDecl defines a soft preference.
type PreferDecl struct {
	Name   string  `json:"name"`
	Expr   string  `json:"expr"`
	Weight float64 `json:"weight,omitempty"`
}

// HoleDecl defines a typed hole.
type HoleDecl struct {
	Name     string `json:"name"`
	Contract string `json:"contract"`
}

// CandidateDecl declares a candidate for a symbol slot.
type CandidateDecl struct {
	Symbol      string           `json:"symbol"`
	Score       float64          `json:"score,omitempty"`
	Constraints []ConstraintDecl `json:"constraints,omitempty"`
}

// FuncDecl defines a function signature.
type FuncDecl struct {
	Name    string  `json:"name"`
	Params  []Field `json:"params,omitempty"`
	Returns []Field `json:"returns,omitempty"`
}

// Field defines a typed field.
type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Stmt represents a minimal imperative statement.
type Stmt struct {
	Kind     string      `json:"kind"`
	Let      *LetStmt    `json:"let,omitempty"`
	Assign   *AssignStmt `json:"assign,omitempty"`
	If       *IfStmt     `json:"if,omitempty"`
	While    *WhileStmt  `json:"while,omitempty"`
	For      *ForStmt    `json:"for,omitempty"`
	Return   *ReturnStmt `json:"return,omitempty"`
	ExprStmt *Expr       `json:"expr,omitempty"`
}

// LetStmt declares a new local value.
type LetStmt struct {
	Name  string `json:"name"`
	Value Expr   `json:"value"`
}

// AssignStmt assigns to an existing variable.
type AssignStmt struct {
	Name  string `json:"name"`
	Value Expr   `json:"value"`
}

// IfStmt represents a conditional.
type IfStmt struct {
	Cond Expr   `json:"cond"`
	Then []Stmt `json:"then,omitempty"`
	Else []Stmt `json:"else,omitempty"`
}

// WhileStmt represents a loop.
type WhileStmt struct {
	Cond Expr   `json:"cond"`
	Body []Stmt `json:"body,omitempty"`
}

// ForStmt represents a collection loop.
type ForStmt struct {
	Var  string `json:"var"`
	Iter Expr   `json:"iter"`
	Body []Stmt `json:"body,omitempty"`
}

// ReturnStmt represents a return statement.
type ReturnStmt struct {
	Value *Expr `json:"value,omitempty"`
}

// Expr represents a minimal expression tree.
type Expr struct {
	Kind    string      `json:"kind"`
	Ident   string      `json:"ident,omitempty"`
	Literal *Literal    `json:"literal,omitempty"`
	Unary   *UnaryExpr  `json:"unary,omitempty"`
	Binary  *BinaryExpr `json:"binary,omitempty"`
	Call    *CallExpr   `json:"call,omitempty"`
	Member  *MemberExpr `json:"member,omitempty"`
	Index   *IndexExpr  `json:"index,omitempty"`
	List    *ListExpr   `json:"list,omitempty"`
}

// Literal represents a literal value.
type Literal struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// UnaryExpr represents a unary operation.
type UnaryExpr struct {
	Op   string `json:"op"`
	Expr Expr   `json:"expr"`
}

// BinaryExpr represents a binary operation.
type BinaryExpr struct {
	Op    string `json:"op"`
	Left  Expr   `json:"left"`
	Right Expr   `json:"right"`
}

// CallExpr represents a function call.
type CallExpr struct {
	Callee Expr   `json:"callee"`
	Args   []Expr `json:"args,omitempty"`
}

// MemberExpr represents field access.
type MemberExpr struct {
	Object Expr   `json:"object"`
	Field  string `json:"field"`
}

// IndexExpr represents index access.
type IndexExpr struct {
	Object Expr `json:"object"`
	Index  Expr `json:"index"`
}

// ListExpr represents a list literal.
type ListExpr struct {
	Elements []Expr `json:"elements,omitempty"`
}

// PolicyDecl defines a policy block.
type PolicyDecl struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

// Diagnostic is a shared diagnostic format.
type Diagnostic struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
}

// Patch is a minimal structured patch format for AST deltas (v0.1 stub).
type Patch struct {
	Ops []PatchOp `json:"ops"`
}

// PatchOp is a single patch operation.
type PatchOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value,omitempty"`
}
