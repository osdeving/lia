package cli

import (
	"path/filepath"
	"strings"

	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/ir"
)

func loadProgram(path string) (*ir.Program, error) {
	return codec.LoadProgramFile(path)
}

func replaceExt(path, ext string) string {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	return base + ext
}
