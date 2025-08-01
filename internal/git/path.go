package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func GetRootPath(workingDir string) (string, error) {
	cmd := exec.Command(gitCmd, "rev-parse", "--show-toplevel")
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
