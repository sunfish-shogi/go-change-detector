package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func GetRootPath(ctx context.Context, workingDir string) (string, error) {
	cmd := exec.CommandContext(ctx, gitCmd, "rev-parse", "--show-toplevel")
	cmd.Dir = workingDir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			println(stderr.String())
		}
		return "", fmt.Errorf("failed to get git root path: %w", err)
	}
	return strings.TrimSpace(stdout.String()), nil
}
