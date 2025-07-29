package git

import (
	"os/exec"
	"strings"
)

func ChangedFilesFrom(gitRootPath, revision string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", revision)
	cmd.Dir = gitRootPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	return files, nil
}
