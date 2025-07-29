package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func ChangedFilesFrom(gitRootPath, revision string) ([]string, error) {
	cmd := exec.Command(gitCmd, "diff", "--name-only", revision)
	cmd.Dir = gitRootPath
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}
	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	return files, nil
}
