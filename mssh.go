package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/Unknwon/com"
	"log"
	"mssh/helper"
	"mssh/sshutils"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/utils"
//	"github.com/skoo87/log4go"
)

const (
	APP_VER = "0.1.3"
)

const confTpl = `
;main config section
[main]
username=root
password=rootpass
command=echo ok
;ssh timeout default 30s
timeout=30


;crypto config section
[crypto]


;mail config section
[mail]
;mail title
title=mssh
;mail body
body=mssh exec task
;mail mode, eg text,html
mode=text
;multiple users divided by comma, mail to addr, modify it for your mailto addr
maillist=xxx@xxx.com
;mail from addr, modify it for your mailfrom addr
fr_addr=xuewuhen@xxx.com
;mail username, modify it for your username
username=xxx
;mail password, modify it for your password
password=xxx
;mail host, modify it for your smtp host
host=smtp.163.com
;mail port, modify it for your smtp server port
port=25
`

var (
	gUsername string
	gPassword string
	gCommand  string
	gLogFile  string
	cnt       int64
	gTitle    string
	gBody     string
	gMode     string
	gMaillist string
	mailUsername  string
	mailPassword  string
	mailHost  string
	mailPort  int
	gFr_addr  string
	gTimeout  int
	gCfg      string
	buf       bytes.Buffer
)

var (
	cmd         = flag.String("cmd", "", "shell cmds")
	infile      = flag.String("f", "", "ip lists file")
	cfg         = flag.String("cfg", "./mssh.conf", "mssh config file")
	n           = flag.Int("n", 100, "the number of goroutines")
	israndompwd = flag.Bool("rand", false, "random password")
	ismail      = flag.Bool("m", false, "sending mail or not default for no sending mail")
	shellmode   = flag.Bool("s", false, "scp scripts from src host to dst hosts")
	verbose     = flag.Bool("v", false, "show details")
)

type SshClient struct {
	Hostaddr string
	Username string
	Password string
	Cmd      string
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	buf = bytes.Buffer{}
}

func main() {
	flag.Usage = usage
	flag.Parse()

//	SetLog()
//	defer log4go.Close()

	args := flag.Args()
	argnums := flag.NArg()

	if flag.NArg() < 1 && flag.NFlag() < 1 {
		usage()
	}

	if argnums >= 1 {
		switch args[0] {
		case "help":
			usage()
		case "version":
			version()
			os.Exit(2)
		case "&", "|", "<", ">":

		default:
			if args[argnums -1] == "&" {
				args = args[:argnums-1]
			} 

		}
	}

	var err error

	if *infile == "" {
		fmt.Println("please input your infile which included 'hostaddr username password command' one per line")
		usage()
	} else if !com.IsExist(*infile) {
		fmt.Printf("not found input file: %s\n", *infile)
		usage()
	}


	if *cfg == "" {
		gCfg = "./mssh.conf"
	} else {
		gCfg = *cfg
	}

	if !parseconf(gCfg) {
		fmt.Printf("init conf file %s failed, exit\n", *cfg)
		os.Exit(2)
	}

	if *cmd != "" {
		if len(args) > 0 {
			gCommand = fmt.Sprintf("%s %s", *cmd, strings.Join(args, " "))
		} else {
			if *shellmode {
				 curpath, _ := os.Getwd()
                		if ! strings.Contains(*cmd, "/") {
                        		*cmd = curpath + "/" + *cmd
                		}
				gCommand = fmt.Sprintf("%s %s", "/bin/bash", *cmd)	
			} else {
				gCommand = *cmd
			}
		}
	}
	
	gCommand = helper.Abs(gCommand)
	vprintf("gCommnad: %s\n", gCommand)
	vprintf("gUsername: %s\n", gUsername)
	vprintf("gPassword: %s\n", gPassword)
	vprintf("gTimeout: %v\n", gTimeout)

	sshobjs, err := parse(*infile)
	if err != nil {
		log.Fatal(err)
	}
	vprintf("%#v\n", sshobjs)

	if *n < 1 || *n > 10000 {
		*n = runtime.NumCPU()
	}

	cmdout, _, err := com.ExecCmd("logname")
	vprintf("cmdout: %s", cmdout)
	if err == nil {
		cmdout = strings.TrimSpace(cmdout)
		if len(gMaillist) == 0 {
			gMaillist = fmt.Sprintf("%s@163.com", cmdout)
        	}
	}
	vprintf("gMaillist: %s\n", gMaillist)
	
//	basename := filepath.Base(*cmd)
	var r  = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_")
	randomfname  :=  string(helper.RandomCreateBytes(32, r...)) + ".sh"	
	vprintf("randomfname: %s\n", randomfname)

	chFirst := make(chan SshClient, *n)
	//	chSecond := make(chan SshClient, *n)
	oldtime := time.Now()
	scpTout := fmt.Sprintf("ConnectTimeout=%d", gTimeout)

	go func() {
		defer close(chFirst)
		for i, _ := range sshobjs {
			if *israndompwd {
				sshobjs[i].Cmd = fmt.Sprintf("echo %s | passwd --stdin %s", sshobjs[i].Password, sshobjs[i].Username)
				vprintf("%#v\n", sshobjs[i])
			} else if *shellmode {
				
				com.ExecCmd("scp", "-o", scpTout, "-o", "BatchMode=yes", "-r", *cmd, fmt.Sprintf("%s:/tmp/%s", helper.GetHost(sshobjs[i].Hostaddr, ":"), randomfname))
				sshobjs[i].Cmd = fmt.Sprintf("/bin/bash /tmp/%s && /bin/rm -f /tmp/%s", randomfname, randomfname)
				vprintf("%#v\n", sshobjs[i])
			}
			chFirst <- sshobjs[i]
		}
	}()

	wg1 := new(sync.WaitGroup)
	wg2 := new(sync.WaitGroup)
	tryslice := make([]SshClient, 0)
	sem := make(chan struct{}, *n)
	mux := new(sync.Mutex)

	for i := 0; i < *n; i++ {
		wg1.Add(1)
		go func(ch chan SshClient) {
			defer wg1.Done()

			for v := range ch {
				vprintf("%#v\n", v)
				func(sshitem SshClient) {
					err, _, stdout, _ := sshutils.SshExec(sshitem.Hostaddr, sshitem.Username, sshitem.Password, sshitem.Cmd, gTimeout)
					if err != nil {
						mux.Lock()
						tryslice = append(tryslice, sshitem)
						mux.Unlock()
						return
					}

					mux.Lock()
					cnt++
					fmt.Println(strings.Repeat("*", 40), "[", cnt, "]", strings.Repeat("*", 40))
					fmt.Println(fmt.Sprintf("%s:\n%s\n", sshitem.Hostaddr, stdout))
					mux.Unlock()
				}(v)
			}
		}(chFirst)
	}

	wg1.Wait()

	//for /usr/bin/ssh execute shell cmds

	buf.WriteString("\nrunning results(error host lists):\n\n")

	for _, v := range tryslice {
		sem <- struct{}{}
		wg2.Add(1)
		go func(sshitem SshClient) {
			defer wg2.Done()
			err, _, stdout, _ := sshutils.SshCmdExec(sshitem.Hostaddr, sshitem.Username, sshitem.Password, sshitem.Cmd, gTimeout)
			if err != nil {
				log.Println("[/usr/bin/ssh]", sshitem.Hostaddr, err)
//				log4go.Error("[/usr/bin/ssh] %s %v", sshitem.Hostaddr, err)
				buf.WriteString(sshitem.Hostaddr + "\n")
				<-sem
				return
			}

			mux.Lock()
			cnt++
			fmt.Println("[", cnt, "]", strings.Repeat("*", 80))
			fmt.Println(fmt.Sprintf("%s %s:\n%s\n", "[/usr/bin/ssh]", sshitem.Hostaddr, stdout))
//			log4go.Info(fmt.Sprintf("%s %s:\n%s\n", "[/usr/bin/ssh]", sshitem.Hostaddr, stdout))
			mux.Unlock()
			<-sem
		}(v)

	}

	wg2.Wait()

	//kill ssh process which is timeout
	vprintf("Kill ssh for timeout\n")
	kill(sshutils.Cmds)

	errMsg := buf.String()
	htmlErrMsg := strings.Replace(errMsg, "\n", "<br>", -1)
	now := time.Now().String()
	title := gTitle
	mode := gMode
	execTime := time.Since(oldtime)
	totalTime := execTime.String()
	var body string
	if mode == "html" {
		body = now + "<br>" + gBody + "<br>" + htmlErrMsg + "<br>total time: " + totalTime
	} else {
		body = now + "\n" + gBody + "\n" + errMsg + "\ntotal time: " + totalTime
	}
	maillist := []string{gMaillist}
	
	config := fmt.Sprintf(`{"username":"%s","password":"%s","host":"%s","port":%d}`, mailUsername, mailPassword, mailHost, mailPort) 

	mail := utils.NewEMail(config)
	mail.To = maillist
	mail.From = gFr_addr
	mail.Subject = title

	if mode == "html" {
		mail.HTML = body
	} else { 
		mail.Text = body
	}

	if *ismail {
        	err = mail.Send()
        	if err != nil {
                	log.Println("send mail failed:", err)
                	goto LEAVE
        }

        	fmt.Println("send mail successful")
        }
	
LEAVE:
	newtime := time.Since(oldtime)
	fmt.Println("run time:", newtime)

}

var helpmsg string = `mssh is a tool for batching ssh execute commands.
Usage:

	mssh [command] [options] [arguments]

The commands are:
	mssh version

	mssh help


The options are:
	-f	input file(include ip|username|password|cmd)
	-cmd 	shell cmds or shell scripts   
	-cfg	mssh config file default for mssh.conf
	-s	shell mode switch default for false
	-n	the number of goroutines default for 100
	-rand	random password mode
	-m	send mail switch default for false
	-v	show details

The arguments are:
	The arguments of mssh will be passed to option cmd, will be part of cmd. 
	etc:
	mssh -f file1 -cmd ls /etc /home /root  --> ok
	mssh -f file1 -cmd ls -al /etc /home --> bad (-al will be dealed with mssh's option, result in undefined option)
	mssh -f file1 -cmd "ls -al" /etc /tmp /home --> ok
	mssh -f file1 -cmd tmp.sh -s -m  --> ok (exec shell scripts)
	....
	more info wait for you to explore!

For any bugs, please contact xuewuhen2015@gmail.com 
for source code:  https://github.com/xuewuhen/mssh
`

var AuthorInfo = `Powered by xuewuhen2015@gmail.com
Copyright @2014 xuewuhen
Report bugs to xuewuhen2015@gmail.com
https://github.com/xuewuhen/mssh
`

func usage() {
	fmt.Printf("%s", helpmsg)
	os.Exit(2)
}

func version() {
	fmt.Println(filepath.Base(os.Args[0]), "version", APP_VER)
	fmt.Printf("%s", AuthorInfo)
}

func vprintf(format string, a ...interface{}) {
	if *verbose {
		fmt.Printf(format, a...)
	}
}

func kill(s []*exec.Cmd) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("Kill.recover -> ", e)
		}
	}()

	if len(s) == 0 {
		return
	}

	for _, v := range s {
		if v != nil && v.Process != nil {
			if err := v.Process.Kill(); err != nil {
				vprintf("Kill -> %v\n", err)
			} else {
				vprintf("Kill -> %#v succ\n", v)
			}
		}
	}
}
