package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ReadFile(gitRootPath, gitRevision, path string) ([]byte, error) {
	cmd := exec.Command(gitCmd, "show", gitRevision+":"+path)
	cmd.Dir = gitRootPath
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			println(stderr.String())
		}
		return nil, fmt.Errorf("failed to read file %s at revision %s: %w", path, gitRevision, err)
	}
	return stdout.Bytes(), nil
}
