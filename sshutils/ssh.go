package sshutils

import (
	"crypto.go/ssh"
	"os/user"
	"io"
	"io/ioutil"
	"strings"
	"fmt"
	"bytes"
	"time"
	"errors"
)

const (
	ECHO          = 53
	TTY_OP_ISPEED = 128
	TTY_OP_OSPEED = 129
)

type keychain struct {
	keys []ssh.Signer
}

func (k *keychain) Key(i int) (ssh.PublicKey, error) {
	if i < 0 || i >= len(k.keys) {
		return nil, nil
	}
	return k.keys[i].PublicKey(), nil
}

func (k *keychain) Sign(i int, rand io.Reader, data []byte) (sig *ssh.Signature, err error) {
	return k.keys[i].Sign(rand, data)
}

func (k *keychain) loadPEM(file string) error {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return err
	}
	k.keys = append(k.keys, key)
	return nil
}

//cmd as shell command need absolute path 
func SshRun(hostaddr, username, password, cmd string) (rst string, err error) {
	var auths []ssh.AuthMethod
	var keypath []string
	userentry, err := user.Lookup(username)
	if err != nil {
		return err.Error(), err
	}

	keyRSA := fmt.Sprintf("%s/.ssh/id_rsa", userentry.HomeDir)
	keyDSA := fmt.Sprintf("%s/.ssh/id_dsa", userentry.HomeDir)
	keypath = append(keypath, keyRSA, keyDSA)
	
	k := new(keychain)
	k.loadPEM(keyRSA)
	k.loadPEM(keyDSA)
	
	auths = append(auths, ssh.PublicKeys(k.keys...), ssh.Password(password))
	
	config := &ssh.ClientConfig{
		User: username,
		Auth: auths,
	}

	if ! strings.Contains(hostaddr, ":") {
		hostaddr = hostaddr + ":22"	
	}

	client, err := ssh.Dial("tcp", hostaddr, config)
	if err != nil {
		return err.Error(), err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err.Error(), err
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ECHO:          0,     // disable echoing
		TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err = session.RequestPty("xterm", 80, 40, modes); err != nil {
		return err.Error(), err
	}
	
	var bstderr bytes.Buffer
	var bstdout bytes.Buffer
	session.Stderr = &bstderr
	session.Stdout = &bstdout

	if err := session.Run(cmd); err != nil {
		return bstderr.String(), err
	}

	return bstdout.String(), nil 
}

type Results struct {
	err    error
	rc     int
	stdout string
	stderr string
}

//support ssh timeout 
func sshExec(hostaddr, username, password, cmd string, ch chan Results) {
        var auths []ssh.AuthMethod
        var keypath []string
        userentry, err := user.Lookup(username)
        if err != nil {
                ch <- Results{err: err, stderr: err.Error()}
		return
        }

        keyRSA := fmt.Sprintf("%s/.ssh/id_rsa", userentry.HomeDir)
        keyDSA := fmt.Sprintf("%s/.ssh/id_dsa", userentry.HomeDir)
        keypath = append(keypath, keyRSA, keyDSA)
        
        k := new(keychain)
        k.loadPEM(keyRSA)
        k.loadPEM(keyDSA)
        
        auths = append(auths, ssh.PublicKeys(k.keys...), ssh.Password(password))
        
        config := &ssh.ClientConfig{
                User: username,
                Auth: auths,
        }

        if ! strings.Contains(hostaddr, ":") {
                hostaddr = hostaddr + ":22"     
        }

        client, err := ssh.Dial("tcp", hostaddr, config)
        if err != nil {
		ch <- Results{err: err, stderr: err.Error()}
                return 
        }
        defer client.Close()

        session, err := client.NewSession()
        if err != nil {
		ch <- Results{err: err, stderr: err.Error()}
                return
        }
        defer session.Close()

        modes := ssh.TerminalModes{
                ECHO:          0,     // disable echoing
                TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
                TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
        }

        if err = session.RequestPty("xterm", 80, 40, modes); err != nil {
                ch <- Results{err: err, stderr: err.Error()}
		return
        }
        
        var bstderr bytes.Buffer
        var bstdout bytes.Buffer
        session.Stderr = &bstderr
        session.Stdout = &bstdout
	rc := 0
	
        if err := session.Run(cmd); err != nil {
		if ugh, ok := err.(*ssh.ExitError); ok {
			rc = ugh.Waitmsg.ExitStatus()
		}
	
	ch <- Results{err: err, rc: rc, stdout: bstdout.String(), stderr: bstderr.String()}
	return 
        }
	
	ch <- Results{err: nil, rc: rc, stdout: bstdout.String(), stderr: bstderr.String()}
        return 
}

func SshExec(hostaddr, username, password, cmd string, timeout int) (err error, rc int, stdout, stderr string) {
	ch := make(chan Results)
	go sshExec(hostaddr, username, password, cmd, ch)

	for {
		select {
		case r := <-ch:
			err, rc, stdout, stderr = r.err, r.rc, r.stdout, r.stderr
			return
		case <-time.After(time.Duration(timeout) * time.Second):
			err = errors.New(fmt.Sprintf("Timed out after %d seconds", timeout))
			return
		}
	}
}
