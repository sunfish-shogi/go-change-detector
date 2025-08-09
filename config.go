package detector

import (
	"github.com/sunfish-shogi/go-change-detector/internal/git"
)

func SetGitCommandPath(path string) {
	git.SetGitCommandPath(path)
}
