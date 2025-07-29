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
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get git root path: %w", err)
	}
	return strings.TrimSpace(out.String()), nil
}
