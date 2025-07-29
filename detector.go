package detector

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sunfish-shogi/go-change-detector/internal/git"
	"github.com/sunfish-shogi/go-change-detector/internal/golang"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

type Config struct {
	GitRootPath   string   // path to the root directory of the git repository
	BaseCommit    string   // base commit revision to compare against, e.g., "HEAD~" for the previous commit
	GoModulePaths []string // paths to go modules
}

func DetectChangedPackages(config *Config) ([]string, error) {
	detector, err := newChangeDetector(config)
	if err != nil {
		return nil, err
	}
	return detector.detectChangedPackages()
}

type changeDetector struct {
	config              *Config
	gitRootFullPath     string                         // full path to the git root directory
	changedModulesCache map[string]map[string]struct{} // cache for changed modules
	changedGoFiles      map[string]struct{}            // full paths of changed Go source files
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
	changedFiles, err := git.ChangedFilesFrom(gitRootFullPath, config.BaseCommit)
	if err != nil {
		return nil, err
	}
	changedGoFiles := make(map[string]struct{})
	for _, file := range changedFiles {
		if strings.HasSuffix(file, ".go") {
			changedGoFiles[filepath.Join(gitRootFullPath, file)] = struct{}{}
		}
	}
	return &changeDetector{
		config:              config,
		gitRootFullPath:     gitRootFullPath,
		changedModulesCache: make(map[string]map[string]struct{}),
		changedGoFiles:      changedGoFiles,
	}, nil
}

func (cd *changeDetector) detectChangedPackages() ([]string, error) {
	goPackages := make([]golang.GoPackage, 0, 64)
	for _, modulePath := range cd.config.GoModulePaths {
		pkgs, err := golang.ListPackages(modulePath)
		if err != nil {
			return nil, err
		}
		goPackages = append(goPackages, pkgs...)
	}

	var changedPackages = make(map[string]struct{})
	for _, goPackage := range goPackages {
		if changed, err := cd.isPackageChanged(&goPackage); err != nil {
			return nil, err
		} else if changed {
			changedPackages[goPackage.ImportPath] = struct{}{}
		}
	}

	for _, goPackage := range goPackages {
		if _, exists := changedPackages[goPackage.ImportPath]; exists {
			continue // Skip packages that are already marked as changed
		} else if updated, err := cd.isModuleUpdated(&goPackage, changedPackages); err != nil {
			return nil, err
		} else if updated {
			changedPackages[goPackage.ImportPath] = struct{}{}
		}
	}

	var results []string
	for pkg := range changedPackages {
		results = append(results, pkg)
	}
	return results, nil
}

func (cd *changeDetector) isPackageChanged(goPackage *golang.GoPackage) (bool, error) {
	// 1. Check if the go source files have changed
	for _, file := range goPackage.GoFiles {
		fullPath := filepath.Join(goPackage.Dir, file)
		if _, exists := cd.changedGoFiles[fullPath]; exists {
			return true, nil
		}
	}

	// 2. Check if the go:embed files have changed
	// TODO: implement

	return false, nil
}

func (cd *changeDetector) isModuleUpdated(goPackage *golang.GoPackage, changedPackages map[string]struct{}) (bool, error) {
	changedModules, err := cd.listChangedModules(goPackage)
	if err != nil {
		return false, err
	}
	for _, module := range goPackage.Deps {
		// 1 Check other packages in the same module
		if _, exists := changedPackages[module]; exists {
			return true, nil
		}

		// 2 Check if the third-party module has changed
		for module != "" {
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
	if cache, ok := cd.changedModulesCache[goModFullPath]; ok {
		return cache, nil
	}
	currentGoMod, err := cd.readGoModFile(goModFullPath, "")
	if err != nil {
		return nil, err
	}
	previousGoMod, err := cd.readGoModFile(goModFullPath, cd.config.BaseCommit)
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
	cd.changedModulesCache[goModFullPath] = changedModules
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
