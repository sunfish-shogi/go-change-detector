package detector

import (
	"github.com/sunfish-shogi/go-change-detector/internal/git"
	"github.com/sunfish-shogi/go-change-detector/internal/golang"
)

func SetGoCommandPath(path string) {
	golang.SetGoCommandPath(path)
}

func SetGitCommandPath(path string) {
	git.SetGitCommandPath(path)
}
