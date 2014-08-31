package sshutils

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
	//	"mssh/helper"
)

var Cmds []*exec.Cmd = make([]*exec.Cmd, 0)

func SshCmdRun(hostaddr, username, password, cmd string) (rstmsg string, err error) {
	var port string = "22"
	if strings.Contains(hostaddr, ":") {
		host := strings.Split(hostaddr, ":")
		hostaddr = host[0]
		if len(host) > 1 {
			if host[1] != "" {
				port = host[1]
			}
		}
	}
	ssharg1 := fmt.Sprintf("-p %s", port)
	ssharg2 := fmt.Sprintf("%s@%s", username, hostaddr)
	args := []string{ssharg1, ssharg2, cmd}
	command := exec.Command("ssh", args...)

	var bstdout bytes.Buffer
	var bstderr bytes.Buffer

	command.Stdout = &bstdout
	command.Stderr = &bstderr
	err = command.Run()
	if err != nil {
		return bstderr.String(), err
	}
	return bstdout.String(), nil
}

//support timeout for /usr/bin/ssh
func sshCmdExec(hostaddr, username, password, cmd string, ch chan Results) {

	var port string = "22"
	if strings.Contains(hostaddr, ":") {
		host := strings.Split(hostaddr, ":")
		hostaddr = host[0]
		if len(host) > 1 {
			if host[1] != "" {
				port = host[1]
			}
		}
	}
	ssharg1 := fmt.Sprintf("-p %s", port)
	ssharg2 := fmt.Sprintf("%s@%s", username, hostaddr)
	args := []string{"-o", "BatchMode=yes", ssharg1, ssharg2, cmd}
	command := exec.Command("ssh", args...)

	var bstdout bytes.Buffer
	var bstderr bytes.Buffer

	command.Stdout = &bstdout
	command.Stderr = &bstderr
	Cmds = append(Cmds, command)

	err := command.Run()
	ch <- Results{err: err, stdout: bstdout.String(), stderr: bstderr.String()}
	return
}

func SshCmdExec(hostaddr, username, password, cmd string, timeout int) (err error, rc int, stdout, stderr string) {
	ch := make(chan Results)
	go sshCmdExec(hostaddr, username, password, cmd, ch)

	select {
	case r := <-ch:
		err, rc, stdout, stderr = r.err, r.rc, r.stdout, r.stderr
		return
	case <-time.After(time.Duration(timeout) * time.Second):
		err = errors.New(fmt.Sprintf("Timed out after %d seconds", timeout))
		return
	}
}
