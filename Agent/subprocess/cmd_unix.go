//go:build !windows
// +build !windows

package subprocess

import (
	"os/exec"
)

func ExecCommand(name string, arg ...string) (*Cmd, error) {
	cmd := exec.Command(name, arg...)
	return newCmd(cmd), nil
}
