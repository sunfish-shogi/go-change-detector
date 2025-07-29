package golang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

type GoPackage struct {
	Dir        string // full path to the package directory
	ImportPath string // import path of the package
	Name       string // package name
	Root       string // full path to the root directory of the project
	Module     struct {
		Path      string // module path
		Main      bool   // is this the main module?
		Dir       string // full path to the module directory
		GoMod     string // full path to the go.mod file
		GoVersion string // Go version used in the module
	}
	GoFiles        []string // relative paths of Go source files in this package
	IgnoredGoFiles []string // relative paths of Go source files ignored by the build tags
	Imports        []string // import paths of the packages imported by this package
	Deps           []string // paths of all packages imported by this package, recursively
}

func ListPackages(workingDir string) ([]GoPackage, error) {
	return goListJSON[GoPackage](workingDir, "./...")
}

func goListJSON[T any](workingDir string, options ...string) ([]T, error) {
	cmd := exec.Command(goCmd, append([]string{"list", "-json"}, options...)...)
	cmd.Dir = workingDir
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run go-list: %w", err)
	}

	dec := json.NewDecoder(&out)
	results := make([]T, 0, 64)
	for {
		var obj T
		if err := dec.Decode(&obj); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}
		results = append(results, obj)
	}
	return results, nil
}
