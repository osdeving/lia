package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/packs"
)

func addPackDirFlag(cmd *cobra.Command) {
	cmd.Flags().StringArray("pack-dir", nil, "additional pack search dir")
}

func resolvePackDirs(cmd *cobra.Command) []string {
	userDirs, _ := cmd.Flags().GetStringArray("pack-dir")
	if len(userDirs) == 0 {
		return packs.DefaultSearchDirs()
	}

	defaults := packs.DefaultSearchDirs()
	seen := map[string]bool{}
	var dirs []string
	for _, d := range append(userDirs, defaults...) {
		if d == "" {
			continue
		}
		if seen[d] {
			continue
		}
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			seen[d] = true
			dirs = append(dirs, d)
		}
	}
	return dirs
}

func gatherPackRefs(p *ir.Program) []ir.PackRef {
	var refs []ir.PackRef
	seen := map[string]bool{}
	for _, proj := range p.Projects {
		for _, ref := range proj.Uses {
			key := ref.Name + "@" + ref.Version
			if seen[key] {
				continue
			}
			seen[key] = true
			refs = append(refs, ref)
		}
	}
	return refs
}

func loadPacksForProgramWithDirs(dirs []string, p *ir.Program) ([]ir.Pack, []ir.Diagnostic) {
	refs := gatherPackRefs(p)
	if len(refs) == 0 {
		return nil, nil
	}
	loader := packs.Loader{SearchDirs: dirs}
	return loader.LoadAll(refs)
}

func loadPacksForProgram(cmd *cobra.Command, p *ir.Program) ([]ir.Pack, []ir.Diagnostic) {
	return loadPacksForProgramWithDirs(resolvePackDirs(cmd), p)
}
