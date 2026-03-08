package java

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/willams/lia/internal/ir"
)

// Profile selects a deterministic lowering strategy for Java targets.
type Profile string

const (
	ProfilePlain      Profile = "plain"
	ProfileSpringBoot Profile = "spring-boot"
	ProfileQuarkus    Profile = "quarkus"
)

// Options configures Java lowering.
type Options struct {
	Profile Profile
}

func (o Options) normalizedProfile() Profile {
	switch o.Profile {
	case "", ProfilePlain:
		return ProfilePlain
	case ProfileSpringBoot:
		return ProfileSpringBoot
	case ProfileQuarkus:
		return ProfileQuarkus
	default:
		return ProfilePlain
	}
}

func ParseProfile(raw string) (Profile, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", "plain":
		return ProfilePlain, nil
	case "spring", "spring-boot", "springboot":
		return ProfileSpringBoot, nil
	case "quarkus":
		return ProfileQuarkus, nil
	default:
		return "", fmt.Errorf("unknown java profile: %s", raw)
	}
}

// LowerProject lowers with the default plain Java profile.
func LowerProject(p *ir.Program) (*Project, error) {
	return LowerProjectWithOptions(p, Options{Profile: ProfilePlain})
}

// LowerProjectWithOptions lowers using the requested deterministic Java profile.
func LowerProjectWithOptions(p *ir.Program, opts Options) (*Project, error) {
	return lowerProjectWithProfile(p, opts.normalizedProfile())
}

func renderPomForProfile(projectName string, profile Profile) string {
	switch profile {
	case ProfileSpringBoot:
		return strings.TrimSpace(fmt.Sprintf(`
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>3.3.2</version>
    <relativePath/>
  </parent>
  <groupId>lia.generated</groupId>
  <artifactId>%s</artifactId>
  <version>0.1.0-SNAPSHOT</version>
  <properties>
    <java.version>21</java.version>
  </properties>
  <dependencies>
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter</artifactId>
    </dependency>
  </dependencies>
  <build>
    <plugins>
      <plugin>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-maven-plugin</artifactId>
      </plugin>
    </plugins>
  </build>
</project>
`, sanitizeArtifact(projectName))) + "\n"
	case ProfileQuarkus:
		return strings.TrimSpace(fmt.Sprintf(`
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>lia.generated</groupId>
  <artifactId>%s</artifactId>
  <version>0.1.0-SNAPSHOT</version>
  <properties>
    <maven.compiler.release>21</maven.compiler.release>
    <maven.compiler.source>21</maven.compiler.source>
    <maven.compiler.target>21</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    <quarkus.platform.group-id>io.quarkus.platform</quarkus.platform.group-id>
    <quarkus.platform.artifact-id>quarkus-bom</quarkus.platform.artifact-id>
    <quarkus.platform.version>3.15.1</quarkus.platform.version>
    <skipITs>true</skipITs>
  </properties>
  <dependencyManagement>
    <dependencies>
      <dependency>
        <groupId>${quarkus.platform.group-id}</groupId>
        <artifactId>${quarkus.platform.artifact-id}</artifactId>
        <version>${quarkus.platform.version}</version>
        <type>pom</type>
        <scope>import</scope>
      </dependency>
    </dependencies>
  </dependencyManagement>
  <dependencies>
    <dependency>
      <groupId>io.quarkus</groupId>
      <artifactId>quarkus-arc</artifactId>
    </dependency>
  </dependencies>
  <build>
    <plugins>
      <plugin>
        <groupId>${quarkus.platform.group-id}</groupId>
        <artifactId>quarkus-maven-plugin</artifactId>
        <version>${quarkus.platform.version}</version>
        <extensions>true</extensions>
      </plugin>
    </plugins>
  </build>
</project>
`, sanitizeArtifact(projectName))) + "\n"
	default:
		return renderPom(projectName)
	}
}

func renderReadmeForProfile(projectName, basePackage string, p *ir.Program, profile Profile) string {
	switch profile {
	case ProfileSpringBoot, ProfileQuarkus:
		var b strings.Builder
		b.WriteString("# ")
		b.WriteString(projectName)
		b.WriteString("\n\n")
		b.WriteString("Projeto Java gerado pelo lower do LIA.\n\n")
		b.WriteString("- Profile: `")
		b.WriteString(string(profile))
		b.WriteString("`\n")
		b.WriteString("- Base package: `")
		b.WriteString(basePackage)
		b.WriteString("`\n")
		b.WriteString("- Modulos LIA: ")
		b.WriteString(strconv.Itoa(len(p.Modules)))
		b.WriteString("\n")
		b.WriteString("- Compilacao local: `mvn -q -DskipTests compile`\n")
		return b.String()
	default:
		return renderReadme(projectName, basePackage, p)
	}
}

func renderMainClassForProfile(index *javaIndex, p *ir.Program, profile Profile) string {
	switch profile {
	case ProfileSpringBoot:
		return renderJavaFile(index.BasePackage, []string{
			"org.springframework.boot.SpringApplication",
			"org.springframework.boot.autoconfigure.SpringBootApplication",
		}, joinLines([]string{
			"@SpringBootApplication",
			"public final class " + projectMainClass(index.ProjectName) + " {",
			indent(joinLines([]string{
				"public static void main(String[] args) {",
				indent("SpringApplication.run("+projectMainClass(index.ProjectName)+".class, args);", "  "),
				"}",
			}), "  "),
			"}",
		}))
	case ProfileQuarkus:
		return renderJavaFile(index.BasePackage, []string{
			"io.quarkus.runtime.Quarkus",
			"io.quarkus.runtime.QuarkusApplication",
			"io.quarkus.runtime.annotations.QuarkusMain",
		}, joinLines([]string{
			"@QuarkusMain",
			"public final class " + projectMainClass(index.ProjectName) + " implements QuarkusApplication {",
			indent(joinLines([]string{
				"public static void main(String[] args) {",
				indent("Quarkus.run("+projectMainClass(index.ProjectName)+".class, args);", "  "),
				"}",
				"",
				"@Override",
				"public int run(String... args) {",
				indent(`System.out.println("LIA generated Quarkus project: `+escapeJavaString(index.ProjectName)+`");`, "  "),
				indent("return 0;", "  "),
				"}",
			}), "  "),
			"}",
		}))
	default:
		return renderMainClass(index, p)
	}
}

func renderWiringForProfile(index *javaIndex, moduleName, pkg string, decl ir.WiringDecl, usecases []usecaseDeps, profile Profile) string {
	switch profile {
	case ProfileSpringBoot:
		return renderSpringBootWiring(index, moduleName, pkg, decl, usecases)
	case ProfileQuarkus:
		return renderQuarkusWiring(index, moduleName, pkg, decl, usecases)
	default:
		return renderWiring(index, moduleName, pkg, decl, usecases)
	}
}

func renderSpringBootWiring(index *javaIndex, moduleName, pkg string, decl ir.WiringDecl, usecases []usecaseDeps) string {
	imports := map[string]bool{
		"org.springframework.context.annotation.Bean":          true,
		"org.springframework.context.annotation.Configuration": true,
	}
	methods, portFields := buildProfileWiringMethods(index, moduleName, pkg, decl, usecases, false)
	var body []string
	body = append(body, "@Configuration")
	body = append(body, "public class "+wiringClassName(decl.Name)+" {")
	body = append(body, indent(joinLines(methods), "  "))
	body = append(body, "}")
	for _, imp := range portFields {
		imports[imp] = true
	}
	return renderJavaFile(pkg, sortedImports(imports), joinLines(body))
}

func renderQuarkusWiring(index *javaIndex, moduleName, pkg string, decl ir.WiringDecl, usecases []usecaseDeps) string {
	imports := map[string]bool{
		"jakarta.enterprise.context.ApplicationScoped": true,
		"jakarta.enterprise.inject.Produces":           true,
		"jakarta.inject.Singleton":                     true,
	}
	methods, extraImports := buildProfileWiringMethods(index, moduleName, pkg, decl, usecases, true)
	var body []string
	body = append(body, "@ApplicationScoped")
	body = append(body, "public class "+wiringClassName(decl.Name)+" {")
	body = append(body, indent(joinLines(methods), "  "))
	body = append(body, "}")
	for _, imp := range extraImports {
		imports[imp] = true
	}
	return renderJavaFile(pkg, sortedImports(imports), joinLines(body))
}

func buildProfileWiringMethods(index *javaIndex, moduleName, pkg string, decl ir.WiringDecl, usecases []usecaseDeps, quarkus bool) ([]string, []string) {
	imports := map[string]bool{}
	var methods []string
	fieldNames := map[string]string{}

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
		adapterType := adapterClassName(adapter.Name)
		fieldName := dependencyFieldName(port.Name)
		fieldNames[port.QName] = fieldName
		if port.Package != pkg {
			imports[port.Package+"."+portType] = true
		}
		if adapter.Package != pkg {
			imports[adapter.Package+"."+adapterType] = true
		}

		var lines []string
		if quarkus {
			lines = append(lines, "@Produces", "@Singleton")
		} else {
			lines = append(lines, "@Bean")
		}
		lines = append(lines, fmt.Sprintf("public %s %s() {", portType, fieldName))
		lines = append(lines, fmt.Sprintf("  return new %s();", adapterType))
		lines = append(lines, "}")
		methods = append(methods, joinLines(lines))
	}

	for _, uc := range usecases {
		className := usecaseClassName(uc.Usecase.Name)
		if index.ModulePackage[moduleName] != pkg {
			imports[index.ModulePackage[moduleName]+"."+className] = true
		}
		var args []string
		for _, dep := range uc.Ports {
			portType := toJavaTypeName(dep.Name)
			fieldName := fieldNames[dep.QName]
			if dep.Package != pkg {
				imports[dep.Package+"."+portType] = true
			}
			args = append(args, fmt.Sprintf("%s %s", portType, fieldName))
		}
		var lines []string
		if quarkus {
			lines = append(lines, "@Produces", "@Singleton")
		} else {
			lines = append(lines, "@Bean")
		}
		signature := fmt.Sprintf("public %s %s(%s) {", className, toLowerCamel(className), strings.Join(args, ", "))
		lines = append(lines, signature)
		callArgs := make([]string, 0, len(uc.Ports))
		for _, dep := range uc.Ports {
			callArgs = append(callArgs, fieldNames[dep.QName])
		}
		lines = append(lines, fmt.Sprintf("  return new %s(%s);", className, strings.Join(callArgs, ", ")))
		lines = append(lines, "}")
		methods = append(methods, joinLines(lines))
	}

	if len(methods) == 0 {
		methods = append(methods, "// No explicit bindings declared in this wiring.")
	}
	return methods, sortedImports(imports)
}

func profileExtraFiles(profile Profile) []File {
	switch profile {
	case ProfileSpringBoot:
		return []File{
			{
				Path:    filepath.ToSlash(filepath.Join("src", "main", "resources", "application.properties")),
				Content: []byte("spring.main.banner-mode=off\n"),
			},
		}
	case ProfileQuarkus:
		return []File{
			{
				Path:    filepath.ToSlash(filepath.Join("src", "main", "resources", "application.properties")),
				Content: []byte("quarkus.banner.enabled=false\n"),
			},
		}
	default:
		return nil
	}
}

func wiringPackage(index *javaIndex, moduleName string, profile Profile) string {
	if profile == ProfilePlain {
		return index.ModulePackage[moduleName]
	}
	return modulePackage(index.BasePackage+".bootstrap", moduleName)
}

func renderLoweringReport(index *javaIndex, p *ir.Program, profile Profile) string {
	var b strings.Builder
	b.WriteString("# Lowering Report\n\n")
	b.WriteString("- Profile: `")
	b.WriteString(string(profile))
	b.WriteString("`\n")
	b.WriteString("- Project: `")
	b.WriteString(index.ProjectName)
	b.WriteString("`\n")
	b.WriteString("- Base package: `")
	b.WriteString(index.BasePackage)
	b.WriteString("`\n")
	b.WriteString("- Deterministic mapping:\n")
	b.WriteString("  - `type` -> Java `record`\n")
	b.WriteString("  - `enum` -> Java `enum`\n")
	b.WriteString("  - `port` -> Java `interface`\n")
	b.WriteString("  - `usecase` -> constructor-based application class\n")
	b.WriteString("  - `wiring bind` -> explicit composition in the selected profile\n")
	b.WriteString("\n## Modules\n\n")
	for _, mod := range p.Modules {
		b.WriteString("- `")
		b.WriteString(mod.Name)
		b.WriteString("` (`")
		b.WriteString(mod.Role)
		b.WriteString("`) -> `")
		b.WriteString(index.ModulePackage[mod.Name])
		b.WriteString("`\n")
	}
	var placeholders []string
	for moduleName, values := range index.Placeholders {
		for _, value := range values {
			placeholders = append(placeholders, moduleName+":"+value)
		}
	}
	sort.Strings(placeholders)
	b.WriteString("\n## Fidelity Notes\n\n")
	if len(placeholders) == 0 {
		b.WriteString("- Nenhum placeholder foi necessario.\n")
	} else {
		b.WriteString("- Placeholders conservadores gerados para manter compilacao:\n")
		for _, placeholder := range placeholders {
			b.WriteString("  - `")
			b.WriteString(placeholder)
			b.WriteString("`\n")
		}
	}
	b.WriteString("- Adapters continuam com lowering conservador quando a semantica comportamental nao esta completamente especificada na LIA.\n")
	switch profile {
	case ProfileSpringBoot:
		b.WriteString("- Wiring LIA foi traduzido para `@Configuration` + `@Bean`, preservando binds explicitos.\n")
	case ProfileQuarkus:
		b.WriteString("- Wiring LIA foi traduzido para `@Produces`, preservando binds explicitos sem esconder composicao.\n")
	default:
		b.WriteString("- Wiring LIA foi traduzido para composicao manual em Java puro.\n")
	}
	return b.String()
}
