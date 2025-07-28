package git

import (
	"bytes"
	"os/exec"
)

func ReadFile(gitRootPath, gitRevision, path string) ([]byte, error) {
	cmd := exec.Command("git", "show", gitRevision+":"+path)
	cmd.Dir = gitRootPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
