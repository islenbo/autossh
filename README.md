# autossh

一个ssh远程客户端，可一键登录远程服务器，主要用来弥补Mac/Linux Terminal ssh无法保存密码的不足。

![演示](https://raw.githubusercontent.com/islenbo/autossh/b3e18c35ebced882ace59be7843d9a58d1ac74d7/doc/images/ezgif-1-a4ddae192f.gif)

## 功能说明
- 核心代码重构，使用go.mod管理依赖
- 新增分组折叠功能
- 新增自动更新检测功能
- 新增一键安装脚本
- 新增会话日志保存功能
- 删除功能支持ctrl+d退出
- 优化帮助菜单显示
- 修复若干Bug

## 下载
[https://github.com/islenbo/autossh/releases](https://github.com/islenbo/autossh/releases)

## 安装
- Mac/Linux用户直接下载安装包，运行install脚本即可。
- Windows用户可手动编译，参考编译章节。

## config.json字段说明
- TODO 字段说明

## Q&amp;A

##### Q: Downloads中为什么没有Windows的包？
##### A: Windows下有很多ssh工具，autossh主要是面向Mac/Linux群体。

----

##### Q: cp 命令出现报错: ssh: subsystem request failed
##### A: 修改服务器 /etc/ssh/sshd_config 将 `Subsystem       sftp    /usr/libexec/openssh/sftp-server` 的注释打开，重启 sshd 服务

## 编译
```bash
export GO111MODULE="on"
export GOFLAGS=" -mod=vendor"
go mod tidy
go build main.go
```

## 依赖
- 查阅 go.mod

## 注意
v0.X版本配置文件无法与v1.X版本兼容，请勿使用！