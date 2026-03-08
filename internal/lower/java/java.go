package java

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/willams/lia/internal/ir"
)

// File is one lowered Java file.
type File struct {
	Path    string
	Content []byte
}

// Project is a lowered Java project with multiple source files.
type Project struct {
	Name        string
	BasePackage string
	Files       []File
}

type typeRef struct {
	Module string
	Name   string
	FQCN   string
}

type portInfo struct {
	Module  string
	Name    string
	QName   string
	Package string
	Decl    *ir.PortDecl
}

type adapterInfo struct {
	Module  string
	Name    string
	QName   string
	Package string
	Decl    *ir.AdapterDecl
}

type methodOwner struct {
	Port       portInfo
	Method     ir.FuncDecl
	JavaMethod string
}

type javaIndex struct {
	ProjectName   string
	BasePackage   string
	ModulePackage map[string]string
	TypesByName   map[string][]typeRef
	EnumsByName   map[string][]typeRef
	Placeholders  map[string][]string
	PortsByQName  map[string]portInfo
	AdaptersByQ   map[string]adapterInfo
	MethodOwners  map[string][]methodOwner
	Bindings      map[string]string
}

type javaType struct {
	Name      string
	Imports   []string
	ZeroValue string
}

type renderContext struct {
	Index             *javaIndex
	CurrentModule     string
	DependencyMethods map[string]methodOwner
}

type usecaseDeps struct {
	Usecase ir.UsecaseDecl
	Ports   []portInfo
}

// Lower emits a manifest for the lowered Java project.
func Lower(p *ir.Program) ([]byte, error) {
	project, err := LowerProject(p)
	if err != nil {
		return nil, err
	}

	sort.Slice(project.Files, func(i, j int) bool {
		return project.Files[i].Path < project.Files[j].Path
	})

	var buf bytes.Buffer
	buf.WriteString("# LIA lower (java)\n")
	buf.WriteString("# project: ")
	buf.WriteString(project.Name)
	buf.WriteByte('\n')
	buf.WriteString("# base_package: ")
	buf.WriteString(project.BasePackage)
	buf.WriteByte('\n')
	for _, file := range project.Files {
		buf.WriteString(file.Path)
		buf.WriteByte('\n')
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// LowerProject lowers a linked unit into a multi-file Java project.
func LowerProject(p *ir.Program) (*Project, error) {
	if p == nil {
		return nil, fmt.Errorf("nil program")
	}

	index := buildIndex(p)
	project := &Project{
		Name:        index.ProjectName,
		BasePackage: index.BasePackage,
	}

	files := []File{
		{
			Path:    "pom.xml",
			Content: []byte(renderPom(index.ProjectName)),
		},
		{
			Path:    "README.md",
			Content: []byte(renderReadme(index.ProjectName, index.BasePackage, p)),
		},
		{
			Path:    javaFilePath(index.BasePackage, projectMainClass(index.ProjectName)),
			Content: []byte(renderMainClass(index, p)),
		},
	}

	usecasesByModule := map[string][]usecaseDeps{}
	for _, mod := range p.Modules {
		for _, uc := range mod.Usecases {
			usecasesByModule[mod.Name] = append(usecasesByModule[mod.Name], usecaseDeps{
				Usecase: uc,
				Ports:   discoverUsecaseDeps(index, uc),
			})
		}
	}

	for _, mod := range p.Modules {
		pkg := index.ModulePackage[mod.Name]

		for _, decl := range mod.Types {
			files = append(files, File{
				Path:    javaFilePath(pkg, toJavaTypeName(decl.Name)),
				Content: []byte(renderTypeDecl(index, mod.Name, pkg, decl)),
			})
		}
		for _, decl := range mod.Enums {
			files = append(files, File{
				Path:    javaFilePath(pkg, toJavaTypeName(decl.Name)),
				Content: []byte(renderEnumDecl(pkg, decl)),
			})
		}
		for _, name := range index.Placeholders[mod.Name] {
			files = append(files, File{
				Path:    javaFilePath(pkg, toJavaTypeName(name)),
				Content: []byte(renderPlaceholderType(pkg, name)),
			})
		}
		for _, decl := range mod.Ports {
			files = append(files, File{
				Path:    javaFilePath(pkg, toJavaTypeName(decl.Name)),
				Content: []byte(renderPortDecl(index, mod.Name, pkg, decl)),
			})
		}
		for _, uc := range usecasesByModule[mod.Name] {
			files = append(files, File{
				Path:    javaFilePath(pkg, usecaseInputName(uc.Usecase.Name)),
				Content: []byte(renderPayloadRecord(index, mod.Name, pkg, usecaseInputName(uc.Usecase.Name), uc.Usecase.Inputs)),
			})
			files = append(files, File{
				Path:    javaFilePath(pkg, usecaseOutputName(uc.Usecase.Name)),
				Content: []byte(renderPayloadRecord(index, mod.Name, pkg, usecaseOutputName(uc.Usecase.Name), uc.Usecase.Outputs)),
			})
			files = append(files, File{
				Path:    javaFilePath(pkg, usecaseClassName(uc.Usecase.Name)),
				Content: []byte(renderUsecase(index, mod.Name, pkg, uc)),
			})
		}
		for _, decl := range mod.Adapters {
			files = append(files, File{
				Path:    javaFilePath(pkg, adapterClassName(decl.Name)),
				Content: []byte(renderAdapter(index, mod.Name, pkg, decl)),
			})
		}
		for _, decl := range mod.Wirings {
			files = append(files, File{
				Path:    javaFilePath(pkg, wiringClassName(decl.Name)),
				Content: []byte(renderWiring(index, mod.Name, pkg, decl, usecasesByModule[mod.Name])),
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	project.Files = files
	return project, nil
}

// WriteProject writes a lowered Java project to disk.
func WriteProject(dir string, project *Project) error {
	if project == nil {
		return fmt.Errorf("nil project")
	}
	for _, file := range project.Files {
		target := filepath.Join(dir, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, file.Content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func buildIndex(p *ir.Program) *javaIndex {
	name := "lia-generated"
	if len(p.Projects) > 0 && strings.TrimSpace(p.Projects[0].Name) != "" {
		name = strings.TrimSpace(p.Projects[0].Name)
	}
	basePackage := "lia.generated." + sanitizePackageSegment(name)
	idx := &javaIndex{
		ProjectName:   name,
		BasePackage:   basePackage,
		ModulePackage: map[string]string{},
		TypesByName:   map[string][]typeRef{},
		EnumsByName:   map[string][]typeRef{},
		Placeholders:  map[string][]string{},
		PortsByQName:  map[string]portInfo{},
		AdaptersByQ:   map[string]adapterInfo{},
		MethodOwners:  map[string][]methodOwner{},
		Bindings:      map[string]string{},
	}

	for _, mod := range p.Modules {
		idx.ModulePackage[mod.Name] = modulePackage(basePackage, mod.Name)
	}
	for _, mod := range p.Modules {
		pkg := idx.ModulePackage[mod.Name]
		for _, decl := range mod.Types {
			ref := typeRef{
				Module: mod.Name,
				Name:   decl.Name,
				FQCN:   pkg + "." + toJavaTypeName(decl.Name),
			}
			idx.TypesByName[decl.Name] = append(idx.TypesByName[decl.Name], ref)
		}
		for _, decl := range mod.Enums {
			ref := typeRef{
				Module: mod.Name,
				Name:   decl.Name,
				FQCN:   pkg + "." + toJavaTypeName(decl.Name),
			}
			idx.EnumsByName[decl.Name] = append(idx.EnumsByName[decl.Name], ref)
		}
		for i := range mod.Ports {
			decl := &mod.Ports[i]
			info := portInfo{
				Module:  mod.Name,
				Name:    decl.Name,
				QName:   mod.Name + "::port:" + decl.Name,
				Package: pkg,
				Decl:    decl,
			}
			idx.PortsByQName[info.QName] = info
			for _, method := range decl.Methods {
				idx.MethodOwners[method.Name] = append(idx.MethodOwners[method.Name], methodOwner{
					Port:       info,
					Method:     method,
					JavaMethod: toLowerCamel(method.Name),
				})
			}
		}
		for i := range mod.Adapters {
			decl := &mod.Adapters[i]
			info := adapterInfo{
				Module:  mod.Name,
				Name:    decl.Name,
				QName:   mod.Name + "::adapter:" + decl.Name,
				Package: pkg,
				Decl:    decl,
			}
			idx.AdaptersByQ[info.QName] = info
		}
		for _, wiring := range mod.Wirings {
			for _, bind := range wiring.Binds {
				left, right, ok := parseBind(bind)
				if ok {
					idx.Bindings[left] = right
				}
			}
		}
	}
	for _, mod := range p.Modules {
		pkg := idx.ModulePackage[mod.Name]
		seen := map[string]bool{}
		for _, name := range collectUnknownTypeNames(mod) {
			if knownBuiltinType(name) {
				continue
			}
			if len(idx.TypesByName[name]) > 0 || len(idx.EnumsByName[name]) > 0 {
				continue
			}
			if seen[name] {
				continue
			}
			seen[name] = true
			idx.Placeholders[mod.Name] = append(idx.Placeholders[mod.Name], name)
			idx.TypesByName[name] = append(idx.TypesByName[name], typeRef{
				Module: mod.Name,
				Name:   name,
				FQCN:   pkg + "." + toJavaTypeName(name),
			})
		}
		sort.Strings(idx.Placeholders[mod.Name])
	}

	return idx
}

func renderPom(projectName string) string {
	artifact := sanitizeArtifact(projectName)
	return strings.TrimSpace(fmt.Sprintf(`
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>lia.generated</groupId>
  <artifactId>%s</artifactId>
  <version>0.1.0-SNAPSHOT</version>
  <properties>
    <maven.compiler.source>21</maven.compiler.source>
    <maven.compiler.target>21</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
</project>
`, artifact)) + "\n"
}

func renderReadme(projectName, basePackage string, p *ir.Program) string {
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(projectName)
	b.WriteString("\n\n")
	b.WriteString("Projeto Java gerado pelo lower do LIA.\n\n")
	b.WriteString("- Base package: `")
	b.WriteString(basePackage)
	b.WriteString("`\n")
	b.WriteString("- Modulos LIA: ")
	b.WriteString(strconv.Itoa(len(p.Modules)))
	b.WriteString("\n")
	b.WriteString("- Compilacao local: `javac $(find src/main/java -name '*.java')`\n")
	return b.String()
}

func renderMainClass(index *javaIndex, p *ir.Program) string {
	className := projectMainClass(index.ProjectName)
	imports := map[string]bool{}
	var initLines []string
	var wiringNames []string
	for _, mod := range p.Modules {
		for _, wiring := range mod.Wirings {
			name := wiringClassName(wiring.Name)
			pkg := index.ModulePackage[mod.Name]
			if pkg != index.BasePackage {
				imports[pkg+"."+name] = true
			}
			varName := toLowerCamel(name)
			initLines = append(initLines, fmt.Sprintf("var %s = new %s();", varName, name))
			wiringNames = append(wiringNames, varName)
		}
	}

	var body []string
	body = append(body, fmt.Sprintf("System.out.println(\"LIA generated Java project: %s\");", escapeJavaString(index.ProjectName)))
	if len(wiringNames) == 0 {
		body = append(body, "System.out.println(\"No wiring blocks were generated.\");")
	} else {
		body = append(body, initLines...)
		body = append(body, fmt.Sprintf("System.out.println(\"Wiring blocks available: %d\");", len(wiringNames)))
	}

	return renderJavaFile(index.BasePackage, sortedImports(imports), strings.Join([]string{
		"public final class " + className + " {",
		indent(joinLines([]string{
			"public static void main(String[] args) {",
			indent(joinLines(body), "    "),
			"}",
		}), "  "),
		"}",
	}, "\n"))
}

func renderTypeDecl(index *javaIndex, moduleName, pkg string, decl ir.TypeDecl) string {
	name := toJavaTypeName(decl.Name)
	base := resolveType(index, moduleName, decl.Base)
	imports := map[string]bool{}
	for _, imp := range base.Imports {
		if !samePackage(pkg, imp) {
			imports[imp] = true
		}
	}

	var ctorLines []string
	if base.Name != "Object" && !isPrimitiveJavaType(base.Name) {
		imports["java.util.Objects"] = true
		ctorLines = append(ctorLines, fmt.Sprintf("Objects.requireNonNull(value, \"%s value is required\");", name))
	}
	if strings.EqualFold(strings.TrimSpace(decl.Predicate), "nonEmpty") && base.Name == "String" {
		ctorLines = append(ctorLines, "if (value.isBlank()) {")
		ctorLines = append(ctorLines, "  throw new IllegalArgumentException(\"value must not be blank\");")
		ctorLines = append(ctorLines, "}")
	}
	if strings.TrimSpace(decl.Predicate) != "" && !strings.EqualFold(strings.TrimSpace(decl.Predicate), "nonEmpty") {
		ctorLines = append(ctorLines, "// Predicate retained from LIA: "+strings.TrimSpace(decl.Predicate))
	}

	var body []string
	body = append(body, fmt.Sprintf("public record %s(%s value) {", name, base.Name))
	if len(ctorLines) > 0 {
		body = append(body, indent(fmt.Sprintf("public %s {", name), "  "))
		body = append(body, indent(joinLines(ctorLines), "    "))
		body = append(body, indent("}", "  "))
	}
	body = append(body, "}")
	return renderJavaFile(pkg, sortedImports(imports), joinLines(body))
}

func renderEnumDecl(pkg string, decl ir.EnumDecl) string {
	values := make([]string, 0, len(decl.Values))
	for _, value := range decl.Values {
		values = append(values, sanitizeJavaConst(value))
	}
	return renderJavaFile(pkg, nil, fmt.Sprintf("public enum %s {\n  %s\n}\n", toJavaTypeName(decl.Name), strings.Join(values, ", ")))
}

func renderPlaceholderType(pkg, name string) string {
	return renderJavaFile(pkg, nil, fmt.Sprintf("public record %s(Object value) {\n}\n", toJavaTypeName(name)))
}

func renderPortDecl(index *javaIndex, moduleName, pkg string, decl ir.PortDecl) string {
	imports := map[string]bool{}
	var members []string
	for _, method := range decl.Methods {
		sig, nested, methodImports := renderPortMethod(index, moduleName, pkg, decl.Name, method)
		for _, imp := range methodImports {
			if !samePackage(pkg, imp) {
				imports[imp] = true
			}
		}
		members = append(members, sig)
		if nested != "" {
			members = append(members, nested)
		}
	}
	return renderJavaFile(pkg, sortedImports(imports), joinLines([]string{
		"public interface " + toJavaTypeName(decl.Name) + " {",
		indent(joinLines(members), "  "),
		"}",
	}))
}

func renderPortMethod(index *javaIndex, moduleName, pkg, portName string, method ir.FuncDecl) (string, string, []string) {
	imports := map[string]bool{}
	var params []string
	for _, field := range method.Params {
		jt := resolveType(index, moduleName, field.Type)
		for _, imp := range jt.Imports {
			if !samePackage(pkg, imp) {
				imports[imp] = true
			}
		}
		params = append(params, jt.Name+" "+sanitizeJavaIdentifier(field.Name))
	}

	returnType := "void"
	var nested string
	switch len(method.Returns) {
	case 0:
		returnType = "void"
	case 1:
		jt := resolveType(index, moduleName, method.Returns[0].Type)
		returnType = jt.Name
		for _, imp := range jt.Imports {
			if !samePackage(pkg, imp) {
				imports[imp] = true
			}
		}
	default:
		returnType = nestedPortResultName(method.Name)
		nested = renderNestedResultRecord(index, moduleName, pkg, returnType, method.Returns)
	}

	sig := fmt.Sprintf("%s %s(%s);", returnType, toLowerCamel(method.Name), strings.Join(params, ", "))
	return sig, nested, sortedImports(imports)
}

func renderNestedResultRecord(index *javaIndex, moduleName, pkg, name string, fields []ir.Field) string {
	var comps []string
	for _, field := range fields {
		jt := resolveType(index, moduleName, field.Type)
		comps = append(comps, jt.Name+" "+sanitizeJavaIdentifier(field.Name))
	}
	return fmt.Sprintf("record %s(%s) {}", name, strings.Join(comps, ", "))
}

func renderPayloadRecord(index *javaIndex, moduleName, pkg, name string, fields []ir.Field) string {
	imports := map[string]bool{}
	var comps []string
	var ctorLines []string
	for _, field := range fields {
		jt := resolveType(index, moduleName, field.Type)
		for _, imp := range jt.Imports {
			if !samePackage(pkg, imp) {
				imports[imp] = true
			}
		}
		comps = append(comps, jt.Name+" "+sanitizeJavaIdentifier(field.Name))
		if !isPrimitiveJavaType(jt.Name) && jt.Name != "Object" {
			imports["java.util.Objects"] = true
			ctorLines = append(ctorLines, fmt.Sprintf("Objects.requireNonNull(%s, \"%s is required\");", sanitizeJavaIdentifier(field.Name), sanitizeJavaIdentifier(field.Name)))
		}
	}

	var body []string
	body = append(body, fmt.Sprintf("public record %s(%s) {", name, strings.Join(comps, ", ")))
	if len(ctorLines) > 0 {
		body = append(body, indent(fmt.Sprintf("public %s {", name), "  "))
		body = append(body, indent(joinLines(ctorLines), "    "))
		body = append(body, indent("}", "  "))
	}
	body = append(body, "}")
	return renderJavaFile(pkg, sortedImports(imports), joinLines(body))
}

func renderUsecase(index *javaIndex, moduleName, pkg string, uc usecaseDeps) string {
	imports := map[string]bool{}
	className := usecaseClassName(uc.Usecase.Name)
	inputName := usecaseInputName(uc.Usecase.Name)
	outputName := usecaseOutputName(uc.Usecase.Name)
	dependencyMethods := map[string]methodOwner{}
	var fields []string
	var ctorParams []string
	var ctorBody []string

	for _, dep := range uc.Ports {
		typeName := toJavaTypeName(dep.Name)
		if dep.Package != pkg {
			imports[dep.Package+"."+typeName] = true
		}
		fieldName := dependencyFieldName(dep.Name)
		fields = append(fields, fmt.Sprintf("private final %s %s;", typeName, fieldName))
		ctorParams = append(ctorParams, fmt.Sprintf("%s %s", typeName, fieldName))
		ctorBody = append(ctorBody, fmt.Sprintf("this.%s = %s;", fieldName, fieldName))
		for _, method := range dep.Decl.Methods {
			dependencyMethods[method.Name] = methodOwner{
				Port:       dep,
				Method:     method,
				JavaMethod: toLowerCamel(method.Name),
			}
		}
	}

	for _, field := range uc.Usecase.Inputs {
		jt := resolveType(index, moduleName, field.Type)
		for _, imp := range jt.Imports {
			if !samePackage(pkg, imp) {
				imports[imp] = true
			}
		}
	}
	for _, field := range uc.Usecase.Outputs {
		jt := resolveType(index, moduleName, field.Type)
		for _, imp := range jt.Imports {
			if !samePackage(pkg, imp) {
				imports[imp] = true
			}
		}
	}
	if usesListConstructs(uc.Usecase.Body) {
		imports["java.util.List"] = true
	}

	ctx := renderContext{
		Index:             index,
		CurrentModule:     moduleName,
		DependencyMethods: dependencyMethods,
	}

	var methodBody []string
	for _, field := range uc.Usecase.Inputs {
		name := sanitizeJavaIdentifier(field.Name)
		methodBody = append(methodBody, fmt.Sprintf("var %s = input.%s();", name, name))
	}
	if len(uc.Usecase.Body) == 0 {
		methodBody = append(methodBody, defaultReturnLine(index, moduleName, outputName, uc.Usecase.Outputs))
	} else {
		methodBody = append(methodBody, renderStatements(uc.Usecase.Body, ctx, outputName, uc.Usecase.Outputs)...)
		if !hasTerminalReturn(uc.Usecase.Body) {
			methodBody = append(methodBody, defaultReturnLine(index, moduleName, outputName, uc.Usecase.Outputs))
		}
	}

	var classBody []string
	classBody = append(classBody, "public final class "+className+" {")
	if len(fields) > 0 {
		classBody = append(classBody, indent(joinLines(fields), "  "))
		classBody = append(classBody, "")
		classBody = append(classBody, indent(fmt.Sprintf("public %s(%s) {", className, strings.Join(ctorParams, ", ")), "  "))
		classBody = append(classBody, indent(joinLines(ctorBody), "    "))
		classBody = append(classBody, indent("}", "  "))
	} else {
		classBody = append(classBody, indent(fmt.Sprintf("public %s() {", className), "  "))
		classBody = append(classBody, indent("}", "  "))
	}
	classBody = append(classBody, "")
	if len(uc.Usecase.Inputs) > 0 {
		classBody = append(classBody, indent(fmt.Sprintf("public %s execute(%s input) {", outputName, inputName), "  "))
	} else {
		classBody = append(classBody, indent(fmt.Sprintf("public %s execute() {", outputName), "  "))
	}
	classBody = append(classBody, indent(joinLines(methodBody), "    "))
	classBody = append(classBody, indent("}", "  "))
	classBody = append(classBody, "}")

	return renderJavaFile(pkg, sortedImports(imports), joinLines(classBody))
}

func renderAdapter(index *javaIndex, moduleName, pkg string, decl ir.AdapterDecl) string {
	className := adapterClassName(decl.Name)
	imports := map[string]bool{}
	interfaceName := ""
	var methods []string
	if decl.Implements != "" {
		if sym, ok := parseSymbolQName(decl.Implements); ok {
			if port, ok := index.PortsByQName[sym.Module+"::"+sym.Kind+":"+sym.Name]; ok {
				interfaceName = toJavaTypeName(port.Name)
				if port.Package != pkg {
					imports[port.Package+"."+interfaceName] = true
				}
				for _, method := range port.Decl.Methods {
					for _, imp := range adapterMethodImports(index, moduleName, pkg, method) {
						imports[imp] = true
					}
					methods = append(methods, renderAdapterMethod(index, moduleName, method, interfaceName))
				}
			}
		}
	}
	if len(methods) == 0 {
		methods = append(methods, "// Adapter body retained conservatively; semantic lowering of adapter behavior is still partial.")
	}

	header := "public final class " + className
	if interfaceName != "" {
		header += " implements " + interfaceName
	}
	return renderJavaFile(pkg, sortedImports(imports), joinLines([]string{
		header + " {",
		indent(joinLines([]string{
			fmt.Sprintf("public %s() {", className),
			"}",
			"",
			joinLines(methods),
		}), "  "),
		"}",
	}))
}

func renderAdapterMethod(index *javaIndex, moduleName string, method ir.FuncDecl, interfaceName string) string {
	var params []string
	for _, field := range method.Params {
		jt := resolveType(index, moduleName, field.Type)
		params = append(params, jt.Name+" "+sanitizeJavaIdentifier(field.Name))
	}
	returnType, zero := adapterMethodReturnType(index, moduleName, method, interfaceName)
	var body []string
	body = append(body, "@Override")
	body = append(body, fmt.Sprintf("public %s %s(%s) {", returnType, toLowerCamel(method.Name), strings.Join(params, ", ")))
	body = append(body, "  // Deterministic adapter stub generated from LIA binding.")
	if returnType != "void" {
		body = append(body, "  return "+zero+";")
	}
	body = append(body, "}")
	return joinLines(body)
}

func adapterMethodImports(index *javaIndex, moduleName, pkg string, method ir.FuncDecl) []string {
	imports := map[string]bool{}
	for _, field := range method.Params {
		jt := resolveType(index, moduleName, field.Type)
		for _, imp := range jt.Imports {
			if !samePackage(pkg, imp) {
				imports[imp] = true
			}
		}
	}
	if len(method.Returns) == 1 {
		jt := resolveType(index, moduleName, method.Returns[0].Type)
		for _, imp := range jt.Imports {
			if !samePackage(pkg, imp) {
				imports[imp] = true
			}
		}
	}
	return sortedImports(imports)
}

func renderWiring(index *javaIndex, moduleName, pkg string, decl ir.WiringDecl, usecases []usecaseDeps) string {
	imports := map[string]bool{}
	className := wiringClassName(decl.Name)
	fieldNames := map[string]string{}
	var fields []string
	var ctorLines []string

	for _, bind := range decl.Binds {
		left, right, ok := parseBind(bind)
		if !ok {
			continue
		}
		portSym, ok := parseSymbolQName(left)
		if !ok {
			continue
		}
		adapterSym, ok := parseSymbolQName(right)
		if !ok {
			continue
		}
		port := index.PortsByQName[portSym.Module+"::"+portSym.Kind+":"+portSym.Name]
		adapter := index.AdaptersByQ[adapterSym.Module+"::"+adapterSym.Kind+":"+adapterSym.Name]
		portType := toJavaTypeName(port.Name)
		fieldName := dependencyFieldName(port.Name)
		fieldNames[port.QName] = fieldName
		fields = append(fields, fmt.Sprintf("private final %s %s;", portType, fieldName))
		if port.Package != pkg {
			imports[port.Package+"."+portType] = true
		}
		adapterType := adapterClassName(adapter.Name)
		if adapter.Package != pkg {
			imports[adapter.Package+"."+adapterType] = true
		}
		ctorLines = append(ctorLines, fmt.Sprintf("this.%s = new %s();", fieldName, adapterType))
	}

	var methods []string
	for _, bind := range decl.Binds {
		left, _, ok := parseBind(bind)
		if !ok {
			continue
		}
		portSym, ok := parseSymbolQName(left)
		if !ok {
			continue
		}
		port := index.PortsByQName[portSym.Module+"::"+portSym.Kind+":"+portSym.Name]
		portType := toJavaTypeName(port.Name)
		fieldName := fieldNames[port.QName]
		methods = append(methods, joinLines([]string{
			fmt.Sprintf("public %s %s() {", portType, fieldName),
			"  return " + fieldName + ";",
			"}",
		}))
	}

	for _, uc := range usecases {
		classNameUC := usecaseClassName(uc.Usecase.Name)
		if pkg != index.ModulePackage[moduleName] {
			imports[index.ModulePackage[moduleName]+"."+classNameUC] = true
		}
		var args []string
		for _, dep := range uc.Ports {
			args = append(args, fieldNames[dep.QName])
		}
		methods = append(methods, joinLines([]string{
			fmt.Sprintf("public %s %s() {", classNameUC, toLowerCamel(classNameUC)),
			fmt.Sprintf("  return new %s(%s);", classNameUC, strings.Join(args, ", ")),
			"}",
		}))
	}

	if len(ctorLines) == 0 {
		ctorLines = append(ctorLines, "// No explicit bindings declared in this wiring.")
	}

	return renderJavaFile(pkg, sortedImports(imports), joinLines([]string{
		"public final class " + className + " {",
		indent(joinLines(append([]string{
			joinLines(fields),
			"",
			fmt.Sprintf("public %s() {", className),
			indent(joinLines(ctorLines), "  "),
			"}",
			"",
		}, methods...)), "  "),
		"}",
	}))
}

func discoverUsecaseDeps(index *javaIndex, uc ir.UsecaseDecl) []portInfo {
	seen := map[string]portInfo{}
	for _, methodName := range collectCalledIdents(uc.Body) {
		owners := index.MethodOwners[methodName]
		if len(owners) == 1 {
			seen[owners[0].Port.QName] = owners[0].Port
		}
	}
	deps := make([]portInfo, 0, len(seen))
	for _, dep := range seen {
		deps = append(deps, dep)
	}
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].QName < deps[j].QName
	})
	return deps
}

func collectCalledIdents(stmts []ir.Stmt) []string {
	seen := map[string]bool{}
	var names []string
	var visitExpr func(ir.Expr)
	visitExpr = func(expr ir.Expr) {
		switch expr.Kind {
		case "call":
			if expr.Call != nil && expr.Call.Callee.Kind == "ident" {
				name := expr.Call.Callee.Ident
				if name != "" && !seen[name] {
					seen[name] = true
					names = append(names, name)
				}
			}
			if expr.Call != nil {
				visitExpr(expr.Call.Callee)
				for _, arg := range expr.Call.Args {
					visitExpr(arg)
				}
			}
		case "binary":
			if expr.Binary != nil {
				visitExpr(expr.Binary.Left)
				visitExpr(expr.Binary.Right)
			}
		case "unary":
			if expr.Unary != nil {
				visitExpr(expr.Unary.Expr)
			}
		case "member":
			if expr.Member != nil {
				visitExpr(expr.Member.Object)
			}
		case "index":
			if expr.Index != nil {
				visitExpr(expr.Index.Object)
				visitExpr(expr.Index.Index)
			}
		case "list":
			if expr.List != nil {
				for _, elem := range expr.List.Elements {
					visitExpr(elem)
				}
			}
		}
	}

	var visitStmt func(ir.Stmt)
	visitStmt = func(stmt ir.Stmt) {
		switch stmt.Kind {
		case "let":
			if stmt.Let != nil {
				visitExpr(stmt.Let.Value)
			}
		case "assign":
			if stmt.Assign != nil {
				visitExpr(stmt.Assign.Value)
			}
		case "return":
			if stmt.Return != nil && stmt.Return.Value != nil {
				visitExpr(*stmt.Return.Value)
			}
		case "expr":
			if stmt.ExprStmt != nil {
				visitExpr(*stmt.ExprStmt)
			}
		case "if":
			if stmt.If != nil {
				visitExpr(stmt.If.Cond)
				for _, inner := range stmt.If.Then {
					visitStmt(inner)
				}
				for _, inner := range stmt.If.Else {
					visitStmt(inner)
				}
			}
		case "while":
			if stmt.While != nil {
				visitExpr(stmt.While.Cond)
				for _, inner := range stmt.While.Body {
					visitStmt(inner)
				}
			}
		case "for":
			if stmt.For != nil {
				visitExpr(stmt.For.Iter)
				for _, inner := range stmt.For.Body {
					visitStmt(inner)
				}
			}
		}
	}

	for _, stmt := range stmts {
		visitStmt(stmt)
	}
	sort.Strings(names)
	return names
}

func renderStatements(stmts []ir.Stmt, ctx renderContext, outputName string, outputs []ir.Field) []string {
	var out []string
	for _, stmt := range stmts {
		out = append(out, renderStatement(stmt, ctx, outputName, outputs)...)
	}
	return out
}

func renderStatement(stmt ir.Stmt, ctx renderContext, outputName string, outputs []ir.Field) []string {
	switch stmt.Kind {
	case "let":
		if stmt.Let == nil {
			return []string{"// unsupported empty let"}
		}
		return []string{fmt.Sprintf("var %s = %s;", sanitizeJavaIdentifier(stmt.Let.Name), renderExpr(stmt.Let.Value, ctx))}
	case "assign":
		if stmt.Assign == nil {
			return []string{"// unsupported empty assignment"}
		}
		return []string{fmt.Sprintf("%s = %s;", sanitizeJavaIdentifier(stmt.Assign.Name), renderExpr(stmt.Assign.Value, ctx))}
	case "return":
		if stmt.Return == nil || stmt.Return.Value == nil {
			return []string{defaultReturnLine(ctx.Index, ctx.CurrentModule, outputName, outputs)}
		}
		return []string{returnValueLine(renderExpr(*stmt.Return.Value, ctx), outputName)}
	case "expr":
		if stmt.ExprStmt == nil {
			return []string{"// unsupported empty expression"}
		}
		return []string{renderExpr(*stmt.ExprStmt, ctx) + ";"}
	case "if":
		if stmt.If == nil {
			return []string{"// unsupported empty if"}
		}
		lines := []string{fmt.Sprintf("if (%s) {", renderExpr(stmt.If.Cond, ctx))}
		lines = append(lines, indent(joinLines(renderStatements(stmt.If.Then, ctx, outputName, outputs)), "  "))
		if len(stmt.If.Else) > 0 {
			lines = append(lines, "} else {")
			lines = append(lines, indent(joinLines(renderStatements(stmt.If.Else, ctx, outputName, outputs)), "  "))
		}
		lines = append(lines, "}")
		return lines
	case "while":
		if stmt.While == nil {
			return []string{"// unsupported empty while"}
		}
		return []string{
			fmt.Sprintf("while (%s) {", renderExpr(stmt.While.Cond, ctx)),
			indent(joinLines(renderStatements(stmt.While.Body, ctx, outputName, outputs)), "  "),
			"}",
		}
	case "for":
		if stmt.For == nil {
			return []string{"// unsupported empty for"}
		}
		return []string{
			fmt.Sprintf("for (var %s : %s) {", sanitizeJavaIdentifier(stmt.For.Var), renderExpr(stmt.For.Iter, ctx)),
			indent(joinLines(renderStatements(stmt.For.Body, ctx, outputName, outputs)), "  "),
			"}",
		}
	case "break":
		return []string{"break;"}
	case "continue":
		return []string{"continue;"}
	default:
		return []string{"// unsupported stmt: " + stmt.Kind}
	}
}

func renderExpr(expr ir.Expr, ctx renderContext) string {
	switch expr.Kind {
	case "ident":
		return sanitizeJavaIdentifier(expr.Ident)
	case "literal":
		return renderLiteral(expr.Literal)
	case "unary":
		if expr.Unary == nil {
			return "null"
		}
		return expr.Unary.Op + renderExpr(expr.Unary.Expr, ctx)
	case "binary":
		if expr.Binary == nil {
			return "null"
		}
		return "(" + renderExpr(expr.Binary.Left, ctx) + " " + expr.Binary.Op + " " + renderExpr(expr.Binary.Right, ctx) + ")"
	case "call":
		if expr.Call == nil {
			return "null"
		}
		var args []string
		for _, arg := range expr.Call.Args {
			args = append(args, renderExpr(arg, ctx))
		}
		if expr.Call.Callee.Kind == "ident" {
			if owner, ok := ctx.DependencyMethods[expr.Call.Callee.Ident]; ok {
				return dependencyFieldName(owner.Port.Name) + "." + owner.JavaMethod + "(" + strings.Join(args, ", ") + ")"
			}
		}
		return renderExpr(expr.Call.Callee, ctx) + "(" + strings.Join(args, ", ") + ")"
	case "member":
		if expr.Member == nil {
			return "null"
		}
		return renderExpr(expr.Member.Object, ctx) + "." + sanitizeJavaIdentifier(expr.Member.Field) + "()"
	case "index":
		if expr.Index == nil {
			return "null"
		}
		return renderExpr(expr.Index.Object, ctx) + ".get(" + renderExpr(expr.Index.Index, ctx) + ")"
	case "list":
		if expr.List == nil {
			return "List.of()"
		}
		var elems []string
		for _, elem := range expr.List.Elements {
			elems = append(elems, renderExpr(elem, ctx))
		}
		return "List.of(" + strings.Join(elems, ", ") + ")"
	default:
		return "null"
	}
}

func renderLiteral(lit *ir.Literal) string {
	if lit == nil {
		return "null"
	}
	switch lit.Kind {
	case "string":
		return strconv.Quote(lit.Value)
	case "bool":
		if strings.EqualFold(lit.Value, "true") {
			return "true"
		}
		return "false"
	default:
		return lit.Value
	}
}

func hasTerminalReturn(stmts []ir.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	last := stmts[len(stmts)-1]
	if last.Kind == "return" {
		return true
	}
	if last.Kind == "if" && last.If != nil && len(last.If.Then) > 0 && len(last.If.Else) > 0 {
		return hasTerminalReturn(last.If.Then) && hasTerminalReturn(last.If.Else)
	}
	return false
}

func defaultReturnLine(index *javaIndex, moduleName, outputName string, outputs []ir.Field) string {
	var values []string
	for _, field := range outputs {
		values = append(values, resolveType(index, moduleName, field.Type).ZeroValue)
	}
	return "return new " + outputName + "(" + strings.Join(values, ", ") + ");"
}

func returnValueLine(valueExpr, outputName string) string {
	return "return new " + outputName + "(" + valueExpr + ");"
}

func methodReturnType(index *javaIndex, moduleName, pkg string, method ir.FuncDecl) (string, string) {
	switch len(method.Returns) {
	case 0:
		return "void", ""
	case 1:
		jt := resolveType(index, moduleName, method.Returns[0].Type)
		return jt.Name, jt.ZeroValue
	default:
		typeName := nestedPortResultName(method.Name)
		var zeroParts []string
		for _, field := range method.Returns {
			zeroParts = append(zeroParts, resolveType(index, moduleName, field.Type).ZeroValue)
		}
		return typeName, "new " + typeName + "(" + strings.Join(zeroParts, ", ") + ")"
	}
}

func adapterMethodReturnType(index *javaIndex, moduleName string, method ir.FuncDecl, interfaceName string) (string, string) {
	switch len(method.Returns) {
	case 0:
		return "void", ""
	case 1:
		jt := resolveType(index, moduleName, method.Returns[0].Type)
		return jt.Name, jt.ZeroValue
	default:
		typeName := interfaceName + "." + nestedPortResultName(method.Name)
		var zeroParts []string
		for _, field := range method.Returns {
			zeroParts = append(zeroParts, resolveType(index, moduleName, field.Type).ZeroValue)
		}
		return typeName, "new " + typeName + "(" + strings.Join(zeroParts, ", ") + ")"
	}
}

func resolveType(index *javaIndex, currentModule, raw string) javaType {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return javaType{Name: "Object", ZeroValue: "null"}
	}
	if strings.HasPrefix(raw, "[]") {
		inner := resolveType(index, currentModule, strings.TrimSpace(strings.TrimPrefix(raw, "[]")))
		return javaType{
			Name:      "List<" + boxedJavaType(inner.Name) + ">",
			Imports:   append([]string{"java.util.List"}, inner.Imports...),
			ZeroValue: "List.of()",
		}
	}

	switch raw {
	case "String":
		return javaType{Name: "String", ZeroValue: "null"}
	case "Bool", "bool", "Boolean":
		return javaType{Name: "boolean", ZeroValue: "false"}
	case "Int", "int", "Integer":
		return javaType{Name: "int", ZeroValue: "0"}
	case "Long", "long":
		return javaType{Name: "long", ZeroValue: "0L"}
	case "Float", "float":
		return javaType{Name: "double", ZeroValue: "0.0d"}
	case "Double", "double":
		return javaType{Name: "double", ZeroValue: "0.0d"}
	}

	if strings.Contains(raw, "<") || strings.Contains(raw, ">") {
		return javaType{Name: "Object", ZeroValue: "null"}
	}

	if ref, ok := resolveNamedType(index.TypesByName[raw], currentModule); ok {
		return javaType{
			Name:      lastSegment(ref.FQCN),
			Imports:   []string{ref.FQCN},
			ZeroValue: "null",
		}
	}
	if ref, ok := resolveNamedType(index.EnumsByName[raw], currentModule); ok {
		return javaType{
			Name:      lastSegment(ref.FQCN),
			Imports:   []string{ref.FQCN},
			ZeroValue: "null",
		}
	}

	return javaType{Name: toJavaTypeName(raw), ZeroValue: "null"}
}

func collectUnknownTypeNames(mod ir.Module) []string {
	var names []string
	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "[]") || strings.Contains(raw, "<") || strings.Contains(raw, ">") {
			return
		}
		names = append(names, raw)
	}

	for _, decl := range mod.Types {
		add(decl.Base)
	}
	for _, decl := range mod.Ports {
		for _, method := range decl.Methods {
			for _, field := range method.Params {
				add(field.Type)
			}
			for _, field := range method.Returns {
				add(field.Type)
			}
		}
	}
	for _, decl := range mod.Usecases {
		for _, field := range decl.Inputs {
			add(field.Type)
		}
		for _, field := range decl.Outputs {
			add(field.Type)
		}
	}
	for _, decl := range mod.Adapters {
		for _, field := range decl.Inputs {
			add(field.Type)
		}
		for _, field := range decl.Outputs {
			add(field.Type)
		}
	}
	return names
}

func knownBuiltinType(raw string) bool {
	switch raw {
	case "String", "Bool", "bool", "Boolean", "Int", "int", "Integer", "Long", "long", "Float", "float", "Double", "double":
		return true
	default:
		return false
	}
}

func resolveNamedType(candidates []typeRef, currentModule string) (typeRef, bool) {
	if len(candidates) == 0 {
		return typeRef{}, false
	}
	for _, candidate := range candidates {
		if candidate.Module == currentModule {
			return candidate, true
		}
	}
	if len(candidates) == 1 {
		return candidates[0], true
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].FQCN < candidates[j].FQCN
	})
	return candidates[0], true
}

func usesListConstructs(stmts []ir.Stmt) bool {
	var visitExpr func(ir.Expr) bool
	visitExpr = func(expr ir.Expr) bool {
		switch expr.Kind {
		case "list", "index":
			return true
		case "binary":
			return expr.Binary != nil && (visitExpr(expr.Binary.Left) || visitExpr(expr.Binary.Right))
		case "unary":
			return expr.Unary != nil && visitExpr(expr.Unary.Expr)
		case "call":
			if expr.Call == nil {
				return false
			}
			if visitExpr(expr.Call.Callee) {
				return true
			}
			for _, arg := range expr.Call.Args {
				if visitExpr(arg) {
					return true
				}
			}
		case "member":
			return expr.Member != nil && visitExpr(expr.Member.Object)
		}
		return false
	}

	for _, stmt := range stmts {
		switch stmt.Kind {
		case "let":
			if stmt.Let != nil && visitExpr(stmt.Let.Value) {
				return true
			}
		case "assign":
			if stmt.Assign != nil && visitExpr(stmt.Assign.Value) {
				return true
			}
		case "return":
			if stmt.Return != nil && stmt.Return.Value != nil && visitExpr(*stmt.Return.Value) {
				return true
			}
		case "expr":
			if stmt.ExprStmt != nil && visitExpr(*stmt.ExprStmt) {
				return true
			}
		case "if":
			if stmt.If != nil {
				if visitExpr(stmt.If.Cond) || usesListConstructs(stmt.If.Then) || usesListConstructs(stmt.If.Else) {
					return true
				}
			}
		case "while":
			if stmt.While != nil && (visitExpr(stmt.While.Cond) || usesListConstructs(stmt.While.Body)) {
				return true
			}
		case "for":
			return true
		}
	}
	return false
}

func renderJavaFile(pkg string, imports []string, body string) string {
	var b strings.Builder
	if pkg != "" {
		b.WriteString("package ")
		b.WriteString(pkg)
		b.WriteString(";\n\n")
	}
	if len(imports) > 0 {
		for _, imp := range imports {
			b.WriteString("import ")
			b.WriteString(imp)
			b.WriteString(";\n")
		}
		b.WriteByte('\n')
	}
	b.WriteString(strings.TrimSpace(body))
	b.WriteByte('\n')
	return b.String()
}

func javaFilePath(pkg, className string) string {
	return "src/main/java/" + strings.ReplaceAll(pkg, ".", "/") + "/" + className + ".java"
}

func projectMainClass(name string) string {
	base := toJavaTypeName(name)
	if !strings.HasSuffix(base, "Application") {
		base += "Application"
	}
	return base
}

func usecaseClassName(name string) string {
	base := toJavaTypeName(name)
	if strings.HasSuffix(base, "UseCase") || strings.HasSuffix(base, "Usecase") {
		return base
	}
	return base + "UseCase"
}

func usecaseInputName(name string) string {
	return toJavaTypeName(name) + "Input"
}

func usecaseOutputName(name string) string {
	return toJavaTypeName(name) + "Output"
}

func adapterClassName(name string) string {
	base := toJavaTypeName(name)
	if strings.HasSuffix(base, "Adapter") {
		return base
	}
	return base + "Adapter"
}

func wiringClassName(name string) string {
	base := toJavaTypeName(name)
	if strings.HasSuffix(base, "Wiring") {
		return base
	}
	return base + "Wiring"
}

func nestedPortResultName(methodName string) string {
	return toJavaTypeName(methodName) + "Result"
}

func dependencyFieldName(portName string) string {
	return toLowerCamel(portName)
}

func modulePackage(basePackage, moduleName string) string {
	parts := strings.Split(strings.TrimSpace(moduleName), ".")
	var cleaned []string
	for _, part := range parts {
		if part == "" {
			continue
		}
		cleaned = append(cleaned, sanitizePackageSegment(part))
	}
	if len(cleaned) == 0 {
		return basePackage
	}
	return basePackage + "." + strings.Join(cleaned, ".")
}

func parseSymbolQName(q string) (struct {
	Module string
	Kind   string
	Name   string
}, bool) {
	var out struct {
		Module string
		Kind   string
		Name   string
	}
	parts := strings.SplitN(strings.TrimSpace(q), "::", 2)
	if len(parts) != 2 {
		return out, false
	}
	symbol := strings.SplitN(parts[1], ":", 2)
	if len(symbol) != 2 {
		return out, false
	}
	out.Module = strings.TrimSpace(parts[0])
	out.Kind = strings.TrimSpace(symbol[0])
	out.Name = strings.TrimSpace(symbol[1])
	return out, out.Module != "" && out.Kind != "" && out.Name != ""
}

func parseBind(bind string) (string, string, bool) {
	parts := strings.SplitN(bind, "->", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])
	return left, right, left != "" && right != ""
}

func sanitizePackageSegment(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "gen"
	}
	var out []rune
	for i, r := range raw {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			out = append(out, unicode.ToLower(r))
		default:
			if i == 0 || len(out) == 0 || out[len(out)-1] != '_' {
				out = append(out, '_')
			}
		}
	}
	clean := strings.Trim(string(out), "_")
	if clean == "" {
		return "gen"
	}
	if clean[0] >= '0' && clean[0] <= '9' {
		return "p" + clean
	}
	return clean
}

func sanitizeArtifact(raw string) string {
	return strings.ReplaceAll(sanitizePackageSegment(raw), "_", "-")
}

func toJavaTypeName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "GeneratedType"
	}
	var out []rune
	upper := true
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if upper {
				out = append(out, unicode.ToUpper(r))
				upper = false
			} else {
				out = append(out, r)
			}
			continue
		}
		upper = true
	}
	if len(out) == 0 {
		return "GeneratedType"
	}
	if unicode.IsDigit(out[0]) {
		return "T" + string(out)
	}
	return string(out)
}

func toLowerCamel(raw string) string {
	name := toJavaTypeName(raw)
	if name == "" {
		return "value"
	}
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return sanitizeJavaIdentifier(string(runes))
}

func sanitizeJavaIdentifier(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "value"
	}
	var out []rune
	for i, r := range raw {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_':
			out = append(out, r)
		default:
			if i == 0 || len(out) == 0 || out[len(out)-1] != '_' {
				out = append(out, '_')
			}
		}
	}
	name := strings.Trim(string(out), "_")
	if name == "" {
		name = "value"
	}
	if unicode.IsDigit(rune(name[0])) {
		name = "v" + name
	}
	switch name {
	case "class", "interface", "enum", "record", "package", "public", "private", "protected", "return", "default", "if", "else", "while", "for", "switch", "case", "new", "static", "void":
		return name + "Value"
	default:
		return name
	}
}

func sanitizeJavaConst(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "UNKNOWN"
	}
	var out []rune
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out = append(out, unicode.ToUpper(r))
			continue
		}
		if len(out) == 0 || out[len(out)-1] != '_' {
			out = append(out, '_')
		}
	}
	name := strings.Trim(string(out), "_")
	if name == "" {
		return "UNKNOWN"
	}
	if unicode.IsDigit(rune(name[0])) {
		return "V_" + name
	}
	return name
}

func samePackage(currentPkg, fqcn string) bool {
	return strings.TrimSpace(packageOf(fqcn)) == strings.TrimSpace(currentPkg)
}

func packageOf(fqcn string) string {
	idx := strings.LastIndex(fqcn, ".")
	if idx == -1 {
		return ""
	}
	return fqcn[:idx]
}

func lastSegment(fqcn string) string {
	idx := strings.LastIndex(fqcn, ".")
	if idx == -1 {
		return fqcn
	}
	return fqcn[idx+1:]
}

func boxedJavaType(name string) string {
	switch name {
	case "boolean":
		return "Boolean"
	case "int":
		return "Integer"
	case "long":
		return "Long"
	case "double":
		return "Double"
	default:
		return name
	}
}

func isPrimitiveJavaType(name string) bool {
	switch name {
	case "boolean", "int", "long", "double", "float", "short", "byte", "char":
		return true
	default:
		return false
	}
}

func sortedImports(imports map[string]bool) []string {
	if len(imports) == 0 {
		return nil
	}
	items := make([]string, 0, len(imports))
	for imp := range imports {
		if strings.TrimSpace(imp) == "" {
			continue
		}
		items = append(items, imp)
	}
	sort.Strings(items)
	return items
}

func joinLines(lines []string) string {
	var kept []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if len(kept) == 0 || kept[len(kept)-1] == "" {
				continue
			}
			kept = append(kept, "")
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func indent(text, prefix string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line == "" {
			lines[i] = ""
			continue
		}
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func escapeJavaString(value string) string {
	return strings.ReplaceAll(value, "\"", "\\\"")
}
