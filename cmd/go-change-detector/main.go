package main

import (
	"path/filepath"

	detector "github.com/sunfish-shogi/go-change-detector"
	"github.com/sunfish-shogi/go-change-detector/internal/git"
	"github.com/sunfish-shogi/go-change-detector/internal/golang"
)

func main() {
	gitRootPath, err := git.GetRootPath(".")
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

	changedPackages, err := detector.DetectChangedPackages(&detector.Config{
		GitRootPath:   gitRootPath,
		BaseCommit:    "HEAD~",
		GoModulePaths: goModulePaths,
	})
	if err != nil {
		panic(err)
	}

	for _, pkg := range changedPackages {
		println(pkg.Dir)
	}
}
