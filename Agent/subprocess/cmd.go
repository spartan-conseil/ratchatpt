package subprocess

import (
	"os/exec"
	"bytes"
	"syscall"
)

func getErrorExitCode(err error) int {
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return 1
}

type Response struct {
        StdOut   string
        StdErr   string
        ExitCode int
}


type Cmd struct {
	*exec.Cmd
	outbuf, errbuf bytes.Buffer
}

func (c *Cmd) Run() Response {
	var res Response
	c.Cmd.Stdout = &c.outbuf
	c.Cmd.Stderr = &c.errbuf
	err := c.Cmd.Run()
        res.StdOut = c.outbuf.String()
        res.StdErr = c.errbuf.String()
        if err != nil {
                res.ExitCode = getErrorExitCode(err)
        } else {
                res.ExitCode = c.Cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
        }
        if res.StdErr == "" && res.ExitCode != 0 {
                res.StdErr = err.Error()
        }
        return res
}

func newCmd(cmd *exec.Cmd) *Cmd {
	wrapped := &Cmd{Cmd: cmd}
	return wrapped
}
