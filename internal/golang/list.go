package golang

import (
	"context"

	"golang.org/x/tools/go/packages"
)

func ListPackages(ctx context.Context, workingDir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.LoadFiles | packages.LoadImports |
			packages.NeedDeps | packages.NeedForTest |
			packages.NeedModule | packages.NeedEmbedFiles |
			packages.NeedTarget,
		Context: ctx,
		Dir:     workingDir,
		Tests:   true,
	}
	return packages.Load(cfg, "./...")
}
