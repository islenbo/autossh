# autossh

go写的一个ssh远程客户端。可一键登录远程服务器，主要用来弥补Mac/Linux Terminal ssh无法保存密码的不足。

![演示](https://github.com/islenbo/autossh/raw/master/doc/images/ezgif-1-a4ddae192f.gif)

## 版本说明
这是一个全新的autossh，无法兼容v0.2及以下版本，升级前请做好备份！新版配置文件由原来的`servers.json`改为`config.json`，
升级时可将旧配置文件的列表插入到新配置文件的`servers`节点下

注：旧版servers中method=pem需要更新为method=key

## 功能说明
- 支持分组
- 支持显示/隐藏主机详情（show_detail）
- 支持options（目前仅支持ServerAliveInterval）
- 允许配置文件中server默认值为空
- 允许指定配置文件目径
- 修复终端窗口大小改变时无法自适应的bug

## 下载
[https://github.com/islenbo/autossh/releases](https://github.com/islenbo/autossh/releases)

## 安装
- 下载编译好的二进制包autossh，放在指目录下，如`~/autossh`或`/usr/loca/autossh`
- 同级目录下新建`config.json`文件，参考`config.example.json`
- 将安装目录加入环境变量中，或指定别名`alias autossh=your autossh path/autossh`

## config.json
```json
{
  "show_detail": true, // 显示主机详情
  "options": { // 全局配置
    "ServerAliveInterval": 30 // 发送心跳包时间，同 ssh -o ServerAliveInterval=30
  },
  "servers": [
    {
      "name": "vagrant", // 显示名称
      "ip": "192.168.33.10", // 主机地址
      "port": 22, // 端口号，可省略，默认为22
      "user": "root", // 用户名
      "password": "vagrant", // 密码，使用无密码的key登录时可省略
      "method": "password", // 认证方式，可省略，默认值为password，可选项有password、key
      "key": "", // 密钥路径，method=key时有效，可省略，默认为~/.ssh/id_rsa
      "options": { // 自定义配置，会覆盖配置中相同的值
        "ServerAliveInterval": 20
      }
    },
    {
      "name": "vagrant-key",
      "ip": "192.168.33.10",
      "user": "root",
      "method": "key"
    }
  ],
  "groups": [
    {
      "group_name": "your group name",
      "prefix": "a",
      "servers": [
        {
          "name": "example1",
          "ip": "192.168.33.10",
          "user": "root",
          "password": "root"
        },
        {
          "name": "example2",
          "ip": "192.168.33.10",
          "user": "root",
          "password": "root"
        }
      ]
    },
    {
      "group_name": "group2",
      "prefix": "b",
      "servers": [
      ]
    }
  ]
}

```

## Q&amp;A
- Q: Downloads中为什么没有Windows的包？
- A: Windows下有很多ssh工具，autossh主要是面向Mac/Linux群体。

## 编译
go build main.go

## 依赖包
- golang.org/x/crypto/ssh

