package main

import (
	"github.com/sunfish-shogi/go-change-detector/internal/detect"
	"github.com/sunfish-shogi/go-change-detector/internal/git"
)

func main() {
	gitRootPath, err := git.GetRootPath(".")
	if err != nil {
		panic(err)
	}

	changedPackages, err := detect.DetectChangedPackages(&detect.Config{
		GitRootPath: gitRootPath,
		BaseCommit:  "HEAD~",
	})
	if err != nil {
		panic(err)
	}

	for _, pkg := range changedPackages {
		println(pkg)
	}
}
