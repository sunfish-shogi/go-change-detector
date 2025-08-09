package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func ChangedFilesFrom(ctx context.Context, gitRootPath, revision string) ([]string, error) {
	cmd := exec.CommandContext(ctx, gitCmd, "diff", "--name-only", revision)
	cmd.Dir = gitRootPath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			println(stderr.String())
		}
		return nil, fmt.Errorf("failed to execute git-diff: %w", err)
	}
	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	return files, nil
}
