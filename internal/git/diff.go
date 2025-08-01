package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func ChangedFilesFrom(gitRootPath, revision string) ([]string, error) {
	cmd := exec.Command(gitCmd, "diff", "--name-only", revision)
	cmd.Dir = gitRootPath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			println(stderr.String())
		}
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}
	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	return files, nil
}
