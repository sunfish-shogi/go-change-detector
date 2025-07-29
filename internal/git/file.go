package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ReadFile(gitRootPath, gitRevision, path string) ([]byte, error) {
	cmd := exec.Command(gitCmd, "show", gitRevision+":"+path)
	cmd.Dir = gitRootPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to read file %s at revision %s: %w", path, gitRevision, err)
	}
	return out.Bytes(), nil
}
