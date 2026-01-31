package python

import (
	"bytes"
	"fmt"

	"github.com/willams/lia/internal/ir"
)

// Lower is a v0.1 stub that emits a placeholder Python output.
func Lower(p *ir.Program) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("# LIA lower (python) stub\n")
	buf.WriteString("# modules: ")
	buf.WriteString(fmt.Sprintf("%d", len(p.Modules)))
	buf.WriteString("\n")
	return buf.Bytes(), nil
}
