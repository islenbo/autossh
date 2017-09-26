# autossh

go写的一个ssh远程客户端。可一键登录远程服务器，主要用来弥补Mac/Linux Terminal ssh无法保存密码的不足。

使用Mac开发已有几个月，一直没有找到比较好用的ssh客户端。SecureCRT有Mac版，始终觉得没有自带的Terminal好用。而Terminal只是一个终端，
对于经常要通过ssh远程操作的人来说，功能还是太弱了。

其间，我也试过自己写一些shell来辅助，如：`alias sshlocal="ssh root@192.168.33.10"`，但是它无法记住密码自动登录。
再如，使用sshpass实现记住密码，但用着还是各种别扭。原因：
- 这些功能都是编写shell实现的，本人对shell编程并不擅长
- shell脚本逼格不够高

最后，下定决心用golang写一个ssh client。为什么不用C或者Java？因为golang是世界上最好的编译语言，PHP是世界上最好的脚本语言。

## 版本
v0.2

## 下载
[https://github.com/islenbo/autossh/releases](https://github.com/islenbo/autossh/releases)

## 配置
下载编译好的二进制包autossh，在autossh同级目录下新建一个servers.json文件。
编辑servers.json，内容可参考server.example.json
```json
[
  {
    "name": "vagrant", // 显示名称
    "ip": "192.168.33.10", // 服务器IP或域名
    "port": 22, // 端口
    "user": "root", // 用户名
    "password": "vagrant", // 密码
    "method": "password" // 认证方式，目前支持password和pem
  },
  {
    "name": "ssh-pem",
    "ip": "192.168.33.11",
    "port": 22,
    "user": "root",
    "password": "your pem file password or empty", // pem密钥密码，若无密码则留空
    "method": "pem", // pem密钥认证
    "key": "your pem file path" // pem密钥文件绝对路径
  }
  // ...可配置多个
]
```
保存servers.json，执行autossh，即可看到相应服务器信息，输入对应序号，即可自动登录到服务器
![登录演示](https://github.com/islenbo/autossh/raw/master/doc/images/ezgif-4-c8145f96ce.gif)

## 高级用法
设置alias，可在任意目录下调用
```bash
[root@localhost ~]# vim /etc/profile
在行尾追加 alias autossh="~/autossh_path/autossh"
[root@localhost ~]# . /etc/profile
```
更多快捷操作，可调用 `--help` 查看
```bash
[root@localhost autossh]# autossh --help
go写的一个ssh远程客户端。可一键登录远程服务器，主要用来弥补Mac/Linux Terminal ssh无法保存密码的不足。
基本用法：
  直接输入autossh不带任何参数，列出所有服务器，输入对应编号登录。
参数：
  -v, --version 	 显示 autossh 的版本信息。
  -h, --help    	 显示帮助信息。
操作：
  list          	 显示所有server。
  add <name>    	 添加一个 server。如：autossh add vagrant。
  edit <name>   	 编辑一个 server。如：autossh edit vagrant。
  remove <name> 	 删除一个 server。如：autossh remove vagrant
```

## Q&amp;A
- Q: Downloads中为什么没有Windows的包？
- A: Windows下有很多优秀的ssh工具，autossh主要面向Mac/Linux群体。

- Q: 为什么要设置alias而不将autossh放到/usr/bin/下？
- A: autossh核心文件有两个，autossh和servers.json且必须处于同级目录下，所以建议放在其他目录，通过alias调用。

## 编译
go build main.go

## 依赖包
- golang.org/x/crypto/ssh

## TODO
- [x] -v, --version 查看版本号
- [x] -h, --help 显示帮助
- [x] list 显示所有server
- [x] add 添加一个server
- [x] remove name 删除一个server
- [x] edit name 编辑一个server

