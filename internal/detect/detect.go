package detect

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sunfish-shogi/go-change-detector/internal/golang"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

type Config struct {
	GitRootPath string // path to the root directory of the git repository
}

func DetectChangedPackages(config *Config) ([]string, error) {
	detector, err := newChangeDetector(config)
	if err != nil {
		return nil, err
	}
	return detector.detectChangedPackages()
}

type changeDetector struct {
	gitRootFullPath string // full path to the git root directory
}

func newChangeDetector(config *Config) (*changeDetector, error) {
	gitRootPath := config.GitRootPath
	if gitRootPath == "" {
		gitRootPath = "."
	}
	var gitRootFullPath string
	if config != nil && config.GitRootPath != "" {
		path, err := filepath.Abs(config.GitRootPath)
		if err != nil {
			return nil, err
		}
		gitRootFullPath = path
	}
	return &changeDetector{
		gitRootFullPath: gitRootFullPath,
	}, nil
}

func (cd *changeDetector) detectChangedPackages() ([]string, error) {
	goPackages, err := golang.ListPackages(cd.gitRootFullPath)
	if err != nil {
		return nil, err
	}

	var changes []string
	for _, goPackage := range goPackages {
		if changed, err := cd.isPackageChanged(&goPackage); err != nil {
			return nil, err
		} else if changed {
			changes = append(changes, goPackage.ImportPath)
		}
	}

	return changes, nil
}

func (cd *changeDetector) isPackageChanged(goPackage *golang.GoPackage) (bool, error) {
	// 1. Check if the go source files have changed
	// FIXME: implement

	// 2. Check if the go:embed files have changed
	// FIXME: implement

	// 3. Check if the dependencies have changed
	changedModules, err := cd.listChangedModules(goPackage)
	if err != nil {
		return false, err
	}
	for _, module := range goPackage.Deps {
		for module != "" {
			// 3.1 Check other packages in the same module
			// FIXME: implement

			// 3.2 Check if the third-party module has changed
			if _, exists := changedModules[module]; exists {
				return true, nil
			}

			// Move to the parent module
			lastSlashIndex := strings.LastIndex(module, "/")
			if lastSlashIndex == -1 {
				break // No more parent module
			} else {
				module = module[:lastSlashIndex]
			}
		}
	}
	return false, nil
}

func (cd *changeDetector) listChangedModules(goPackage *golang.GoPackage) (map[string]struct{}, error) {
	goModFullPath := goPackage.Module.GoMod
	// FIXME: 同じgo.modに対して何度も呼ばれるのでキャッシュを使用する
	currentGoMod, err := cd.readGoModFile(goModFullPath, "")
	if err != nil {
		return nil, err
	}
	previousGoMod, err := cd.readGoModFile(goModFullPath, "HEAD~")
	if err != nil {
		return nil, err
	}
	previousVersions := getModuleVersions(previousGoMod)
	currentVersions := getModuleVersions(currentGoMod)
	changedModules := make(map[string]struct{})
	for module, currentVersion := range currentVersions {
		previousVersion, exists := previousVersions[module]
		if !exists || currentVersion != previousVersion {
			changedModules[module] = struct{}{}
		}
	}
	return changedModules, nil
}

func getModuleVersions(goMod *modfile.File) map[string]string {
	replaceMap := make(map[module.Version]module.Version)
	for _, replace := range goMod.Replace {
		replaceMap[replace.Old] = replace.New
	}
	moduleVersions := make(map[string]string)
	for _, req := range goMod.Require {
		replace, exists := replaceMap[req.Mod]
		if !exists && req.Mod.Version != "" {
			replace, exists = replaceMap[module.Version{Path: req.Mod.Path, Version: ""}]
		}
		if exists {
			if replace.Version != "" {
				moduleVersions[req.Mod.Path] = fmt.Sprintf("%s@%s", replace.Path, replace.Version)
			} else {
				moduleVersions[req.Mod.Path] = fmt.Sprintf("%s@%s", replace.Path, req.Mod.Version)
			}
		} else {
			moduleVersions[req.Mod.Path] = fmt.Sprintf("%s@%s", req.Mod.Path, req.Mod.Version)
		}
	}
	return moduleVersions
}

func (cd *changeDetector) readGoModFile(fullPath string, gitRevision string) (*modfile.File, error) {
	var data []byte
	if gitRevision != "" {
		if !strings.HasPrefix(fullPath, cd.gitRootFullPath) {
			return nil, errors.New("go.mod file is not in the git root directory")
		}
		gitPath := strings.TrimPrefix(fullPath, cd.gitRootFullPath+"/")
		cmd := exec.Command("git", "show", gitRevision+":"+gitPath)
		cmd.Dir = cd.gitRootFullPath
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return nil, err
		}
		data = out.Bytes()
	} else {
		var err error
		data, err = os.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}
	}

	return modfile.Parse("go.mod", data, nil)
}
