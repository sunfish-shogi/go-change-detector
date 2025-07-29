package golang

import (
	"io/fs"
	"path/filepath"
)

func FindGoModFiles(rootDir string) ([]string, error) {
	goModFiles := make([]string, 0, 8)
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == "go.mod" {
			goModFiles = append(goModFiles, path)
		}
		return nil
	})
	return goModFiles, err
}
