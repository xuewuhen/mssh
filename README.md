
#mssh使用帮助

##1、工具介绍

> mssh是一个批量远程ssh执行命令的工具。 它具有稳定、高效、准确，执行灵活方便，可以大幅度提高日常工作效率。

> 它的思想是：只用给它提供执行命令或者脚本和指定需要执行命令的机器列表，批量在指定机器上执行命令并返回执行命令的结果（包括执行失败的结果），并且邮件告知执行失败情况。

> 这样就可以只用专注于脚本的编写，不用考虑在多台机器上面批量执行。

##2、对比传统ssh命令

> 对比传统的ssh命令优势：

 1. go语言编写，使用最新的go语言ssh包
 2. 支持多线程
 3. 原生ssh协议支持，支持密码、证书认证
 4. 支持超时（各种密码错误，内存爆，网络不通，登录异常不能登录情况，不会中断批量执行）
 5. 支持发送电子邮件告知执行结果
 6. 采用ssh做认证，更安全
 7. 内存占用少
 8. 支持不同机器同时执行不同命令或者同台机器同时执行不同命令
 9. 执行速度非常快
 10. 可以自定义超时时间
 11. 支持配置文件和命令行选项

##3、mssh用法

###3.1

> ./mssh -h 或者 ./mssh --help 或者 ./mssh help 查看命令帮助信息 
mssh is a tool for batching ssh execute commands. 
Usage:
> 
>      mssh [command] [options] [arguments]
> 
> The commands are:
> 
>      mssh version
> 
>      mssh help


> The options are:

     -f     input file(include ip|username|password|cmd)
     -cmd   shell cmds or shell scripts  
     -cfg   mssh config file default for mssh.conf
     -s     shell mode switch default for false
     -n     the number of goroutines default for 100
     -rand  random password mode
     -m     send mail switch default for false
     -v     show details

> The arguments are:

     The arguments of mssh will be passed to option cmd, will be part of cmd.
     
     etc:
     mssh -f file1 -cmd ls /etc /home /root  --> ok
     mssh -f file1 -cmd ls -al /etc /home --> bad (-al will be dealed with mssh's option, result in undefined option)
     mssh -f file1 -cmd "ls -al" /etc /tmp /home --> ok
     mssh -f file1 -cmd tmp.sh -s -m  --> ok (exec shell scripts)
     ....
     more info wait for you to explore!
> 
> For any bugs, please contact xuewuhen2015@gmail.com.

###3.2
./mssh version 查看版本信息

mssh version 0.1.5

Powered by xuewuhen2015@gmail.com

Copyright @2014 xuewuhen

Report bugs to https://github.com/xuewuhen/mssh

###3.3

> 程序所有选项如下:

    -f     input file(include ip|username|password|cmd)
    -cmd      shell cmds or shell scripts   cdf
    -cfg     mssh config file default for mssh.conf
    -s     shell mode switch default for false
    -n     the number of goroutines default for 100
    -rand     random password mode
    -m     send mail switch default for false
    -v     show details
解析顺序：
命令行选项 >  配置文件

####3.3.1
./mssh -f file  

-f file 指定ip列表文件，每个一行，此选项必须指定，否则程序会退出。

其他选项若不指定,将会使用默认选项。

####3.3.2
./mssh -f file -cmd commands

-cmd 指定需要在远程机器上面执行的命令，如date等，如果不指定该选项默认使用配置文件里面的命令（echo ok）

####3.3.3
./mssh -f file -cfg conf

-cfg conf 指定mssh的配置文件，默认使用当前目录下的mssh.conf作为配置文件

此选项适合在多个配置文件之间切换的情况。

####3.3.4
./mssh -f file -s -cmd tmp.sh

-s 开启shell mode，可以远程执行shell scripts，此选项需要与-cmd xxx.sh配合使用

一般情况下不推荐使用，因为会影响执行速度

####3.3.5
./mssh -f file -n 100

-n num 指定mssh的开启的线程数量，默认是100，最小为1，最大为10000。

####3.3.6
./mssh -f file -rand

-rand 随机密码模式，需要与file里面的特定字段配合使用。

一般情况不会用到。

####3.3.7
./mssh -f file -m -cmd commands

-m 开启发送邮件模式，如果配置文件mssh.conf里面指定了maillist=xxx，那么执行命令失败的主机ip或域名会发送给
相应的maillist。如果没有指定maillist，那么将会发送给当前登录中控的用户（如登录用户是xxx，那么邮件将会发送给xxx@xxx.com)。

####3.3.8
./mssh -f file -v
-v 开启调试模式，将会显示详细的debug信息。

在某些情况下，对于调试命令执行失败非常有效。

####3.3.9
关于配置文件说明

配置文件格式如下：

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
    title=mssh
    body=mssh exec task
    mode=text
    ;multiple users divided by semicolon
    maillist=
    fr_addr=xxx@xxx
注:
发送邮件选项仅-m开启的时候才生效。


####3.3.10
一些常见的用法：

执行命令，命令结果输出到终端

./mssh -f file -cmd "wget -O /tmp/xxx.sh http://xxx/xxx.sh && sh /tmp/xxx.sh"


执行命令，命令结果输出到终端，并且将执行命令出错的机器列表发送给当前登录用户

./mssh -f file -m -cmd "wget -O /tmp/xxx.sh http://xxx/xxx.sh && sh /tmp/xxx.sh"


重定向标准输出和标准错误到文件

./mssh -f file -m -cmd "wget -O /tmp/xxx.sh http://xxx/xxx.sh && sh /tmp/xxx.sh" 1> `date %F`.log.txt 2>  `date %F`.err.txt


后台运行，并且将正常输出和错误输出到文件，方便后续查看

nohup ./mssh -f file -m -cmd "wget -O /tmp/xxx.sh http://xxx/xxx.sh && sh /tmp/xxx.sh" 1> `date %F`.log.txt 2>  `date %F`.err.txt &


执行shell脚本

./mssh -f file -m -s -cmd tmp.sh


执行命令(注意，mssh的非-参数，会被当成cmd的参数)

./mssh -f file -cmd "ls -al" /tmp /root


执行默认命令(echo ok)

./mssh -f file


####3.3.11
高级用法：
./mssh -f file

file 文件格式1

10.1.1.1

10.1.1.2


file 文件格式2

10.1.1.1 root

10.1.1.2 guest


file 文件格式3

10.1.1.1 root rootpass

10.1.1.2 guest guestpass


file 文件格式4

10.1.1.1 root rootpass ls -al

10.1.1.2 guest guestpass  echo ok


file 文件格式5

10.1.1.1 root rootpass

10.1.1.2 guest guestpass  echo ok

10.1.1.3  

10.1.1.4 root

如果相关字段missing，那么将会使用配置文件的配置。

如：file 文件格式5 中 10.1.1.3 用户名、用户名密码、命令missing，

那么对应默认值为10.1.1.3 root rootpass echo ok




##4、执行结果显示

    ./mssh -f ip1 -m 
    **************************************** [ 1 ] ****************************************
    192.168.100.1:
    ok
    
    
    **************************************** [ 2 ] ****************************************
    192.168.100.2:
    ok
    
    
    **************************************** [ 3 ] ****************************************
    192.168.100.3:
    ok



send mail successful

run time: 301.368036ms


##5、安装mssh

 - 配置go环境
 - go get -u github.com/xuewuhen/mssh
 - mkdir $GOPATH/src/crypto.go
 - mv ssh $GOPATH/src/crypto.go/
 - go install or go build
