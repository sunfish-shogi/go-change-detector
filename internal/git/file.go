package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func ReadFile(ctx context.Context, gitRootPath, gitRevision, path string) ([]byte, bool, error) {
	cmd := exec.CommandContext(ctx, gitCmd, "show", gitRevision+":"+path)
	cmd.Dir = gitRootPath
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 {
			// Exit code 128 indicates that the file does not exist at the specified revision.
			return nil, false, nil
		}
		if stderr.Len() > 0 {
			println(stderr.String())
		}
		return nil, false, fmt.Errorf("failed to read file %s at revision %s: %w", path, gitRevision, err)
	}
	return stdout.Bytes(), true, nil
}
