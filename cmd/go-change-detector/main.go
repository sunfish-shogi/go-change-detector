package main

import (
	"context"
	"os"
	"path/filepath"

	detector "github.com/sunfish-shogi/go-change-detector"
	"github.com/sunfish-shogi/go-change-detector/internal/git"
	"github.com/sunfish-shogi/go-change-detector/internal/golang"
)

func main() {
	baseCommit := "HEAD~"
	if len(os.Args) > 1 {
		if os.Args[1] == "--help" || os.Args[1] == "-h" {
			println("Usage: go-change-detector [base-commit]")
			return
		}
		baseCommit = os.Args[1]
	}

	gitRootPath, err := git.GetRootPath(context.Background(), ".")
	if err != nil {
		panic(err)
	}

	goModPaths, err := golang.FindGoModFiles(gitRootPath)
	if err != nil {
		panic(err)
	}
	goModulePaths := make([]string, len(goModPaths))
	for i, goModPath := range goModPaths {
		goModulePaths[i] = filepath.Dir(goModPath)
	}

	changedPackages, err := detector.DetectChangedPackages(context.Background(), &detector.Config{
		GitRootPath:   gitRootPath,
		BaseCommit:    baseCommit,
		GoModulePaths: goModulePaths,
	})
	if err != nil {
		panic(err)
	}

	for _, pkg := range changedPackages {
		println(pkg.Dir)
	}
}
