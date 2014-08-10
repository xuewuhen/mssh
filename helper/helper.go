package helper

import (
	"strings"
	"os/exec"
	"path/filepath"
	"crypto/rand"
	r "math/rand"
	"time"
)

func Abs(cmd string) string {
	
	cmds := strings.Fields(cmd)
	if len(cmds) > 0 {
		cmd0 ,err := exec.LookPath(cmds[0])
		if err != nil {
			return cmd
		}
		cmd0, err = filepath.Abs(cmd0)
		if err != nil {
			return cmd
		}
		cmds[0] = cmd0
		cmd =  strings.Join(cmds, " ")
		
	} 
	return cmd
}

// RandomCreateBytes generate random []byte by specify chars.
func RandomCreateBytes(n int, alphabets ...byte) []byte {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	var randby bool
	if num, err := rand.Read(bytes); num != n || err != nil {
		r.Seed(time.Now().UnixNano())
		randby = true
	}
	for i, b := range bytes {
		if len(alphabets) == 0 {
			if randby {
				bytes[i] = alphanum[r.Intn(len(alphanum))]
			} else {
				bytes[i] = alphanum[b%byte(len(alphanum))]
			}
		} else {
			if randby {
				bytes[i] = alphabets[r.Intn(len(alphabets))]
			} else {
				bytes[i] = alphabets[b%byte(len(alphabets))]
			}
		}
	}
	return bytes
}

func GetHost(s string, sep string) string {
	hostaddr := strings.Split(s, sep)
	return hostaddr[0]
}
