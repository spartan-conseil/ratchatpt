//go:build windows
// +build windows

package subprocess

import (
	"os/exec"
	"syscall"
)

func ExecCommand(name string, arg ...string) (*Cmd, error) {
	cmd := exec.Command(name, arg...)
	var err error
	if err != nil {
		return nil, err
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return newCmd(cmd), nil
}
