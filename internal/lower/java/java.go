package java

import (
	"bytes"
	"fmt"

	"github.com/willams/lia/internal/ir"
)

// Lower is a v0.1 stub that emits a placeholder Java output.
func Lower(p *ir.Program) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("// LIA lower (java) stub\n")
	buf.WriteString("// modules: ")
	buf.WriteString(intToString(len(p.Modules)))
	buf.WriteString("\n")
	return buf.Bytes(), nil
}

func intToString(v int) string {
	return fmt.Sprintf("%d", v)
}
