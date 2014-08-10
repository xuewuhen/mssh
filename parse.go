package main

import (
	"io/ioutil"
	"strings"
	"mssh/helper"
	"github.com/Unknwon/goconfig"
	"github.com/Unknwon/com"
)

func parse(filename string) (sshs []SshClient, err error) {
	
	var sshcli SshClient
	content, err  := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	sshclients := strings.Split(string(content), "\n")
//	sshclients := sshclients[:len(sshclients) - 1]
	for _, v := range sshclients {
		if len(v) == 0 {
			continue
		}
		sshfields := strings.Fields(v)
		for i, v := range sshfields {
			sshfields[i] = strings.TrimSpace(v)
		}
		if len(sshfields) == 1 {
			sshcli.Hostaddr = sshfields[0]
			sshcli.Username = gUsername
			sshcli.Password = gPassword
			sshcli.Cmd = gCommand
		} else if len(sshfields) == 2 {
			sshcli.Hostaddr = sshfields[0]
			sshcli.Username = sshfields[1]
			sshcli.Password = gPassword
                        sshcli.Cmd = gCommand
		} else if len(sshfields) == 3 {
			sshcli.Hostaddr = sshfields[0]
                        sshcli.Username = sshfields[1]
			sshcli.Password = sshfields[2]
			sshcli.Cmd = gCommand
		} else if len(sshfields) >= 4 {
			sshcli.Hostaddr = sshfields[0]
                        sshcli.Username = sshfields[1]
                        sshcli.Password = sshfields[2]
			sshcli.Cmd = strings.Join(sshfields[3:], " ")
			sshcli.Cmd = helper.Abs(sshcli.Cmd)
			vprintf("%#v\n", sshcli.Cmd)
		}

			sshs = append(sshs, sshcli)	
	}

	return 
}

func parseconf(filename string) bool {
	var err error
	if ! com.IsExist(filename) {
		com.WriteFile(filename, []byte(confTpl))
	}

	cfg, err := goconfig.LoadConfigFile(filename)
	if err != nil {
		com.ColorLog("[ERROR] Fail to load (%s) [%s]\n", filename, err)
		return false
	}

	gUsername = cfg.MustValue("main", "username")
	if len(gUsername) == 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'username'\n", filename)
		return false
	}

	gPassword = cfg.MustValue("main", "password")
	if len(gPassword) == 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'password'\n", filename)
		return false
	}
	
	gCommand = cfg.MustValue("main", "command")
	if len(gCommand) == 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'command'\n", filename)
		return false
	}
	
	gTitle = cfg.MustValue("mail", "title")
	if len(gTitle) == 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'title'\n", filename)
		return false
	}

	gBody = cfg.MustValue("mail", "body")
	if len(gBody) == 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'body'\n", filename)	
		return false
	}

	gMode = cfg.MustValue("mail", "mode")
	if len(gMode) == 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'mode'\n", filename)
		return false
	}

	gMaillist = cfg.MustValue("mail", "maillist")

	mailUsername = cfg.MustValue("mail", "username")
	if len(mailUsername) == 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'username'\n", filename)
		com.ColorLog("[HINT] Please set mail 'username' in '%s'\n", filename)
		return false
	}

	mailPassword = cfg.MustValue("mail", "password")
	if len(mailPassword) == 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'password'\n", filename)
		com.ColorLog("[HINT] Please set mail 'password' in '%s'\n", filename)
		return false
	}

	mailHost = cfg.MustValue("mail", "host")
	if len(mailHost) == 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'host'\n", filename)
	 	com.ColorLog("[HINT] Please set mail 'host' in '%s'\n", filename)
		return false
	}

	mailPort = cfg.MustInt("mail", "port")
        if mailPort < 0 {
                com.ColorLog("[ERROR] No valid setting in '%s' key 'port'\n", filename)
		com.ColorLog("[HINT] Please set mail 'port' in '%s'\n", filename)
                return false
        }

	gFr_addr = cfg.MustValue("mail", "fr_addr")
	if len(gFr_addr) == 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'fr_addr'\n", filename)
		return false
	}

	gTimeout = cfg.MustInt("main", "timeout")
	if gTimeout <= 0 {
		com.ColorLog("[ERROR] No valid setting in '%s' key 'timeout'\n", filename)
		return false
	}

	return true
}
