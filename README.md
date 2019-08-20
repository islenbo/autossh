# autossh

一个ssh远程客户端，可一键登录远程服务器，主要用来弥补Mac/Linux Terminal ssh无法保存密码的不足。

![演示](https://raw.githubusercontent.com/islenbo/autossh/c9b52688dabbba8ef6403c2f83f8d758ae0e4dfe/doc/images/ezgif-5-42b5117192fc.gif)

## 功能说明
- 核心代码重构，使用go.mod管理依赖
- 新增分组折叠功能
- 新增自动更新检测功能
- 新增一键安装脚本
- 新增会话日志保存功能
- 删除功能支持ctrl+d退出
- 优化帮助菜单显示
- 修复若干Bug
- 支持SOCKS5代理

## 下载
[https://github.com/islenbo/autossh/releases](https://github.com/islenbo/autossh/releases)

## 安装
- Mac/Linux用户直接下载安装包，运行install脚本即可。
- Windows用户可手动编译，参考编译章节。

## config.json字段说明
```yaml
show_detail: bool <是否显示服务器详情(用户、IP)>
options:
  ServerAliveInterval: int <是否定时发送心跳包，与ssh ServerAliveInterval属性用法相同>
servers:
  - name: string <显示名称>
    ip: string <服务器IP>
    port: int <端口>
    user: string <用户名>
    password: string <密码>
    method: string <鉴权方式，password-密码 key-密钥>
    key: string <私钥路径>
    options:
      ServerAliveInterval: int <与根节点ServerAliveInterval用法相同，该值会覆盖根节点的值>
    alias: string <别名>
    log:
      enable: bool <是否启用日志>
      filename: string <日志路径, 如 /tmp/%n-%u-%dt.log 支持变量请参考下方《日志变量》章节>
      mode: string <遇到同名日志的处理模式，cover-覆盖 append-追加，默认为cover>
groups:
  - group_name: string <组名>
    prefix: string <组前缀>
    servers: array <服务器列表，配置与servers相同，配置说明略>
    collapse: bool <是否折叠>
    proxy:
      type: string <代理方式，目前仅支持SOCKS5>
      server: string <代理服务器地址>
      port: int <端口号>
      user: string <用户，若无留空>
      password: string <密码，若无留空>
```

## 日志变量
变量 | 说明 | 示例
--- | --- | ---
%g | 组名(group_name) | MyGroup1
%n | 显示名称(name) | vagrant
%u | 用户名(user) | root
%a | 别名(alias) | vagrant
%dt | 日期时间 | 20190821223010
%d | 日期 | 20190821

## Q&amp;A

##### Q: Downloads中为什么没有Windows的包？
A: Windows下有很多ssh工具，autossh主要是面向Mac/Linux群体。

----

##### Q: cp 命令出现报错: ssh: subsystem request failed
A: 修改服务器 /etc/ssh/sshd_config 将 `Subsystem       sftp    /usr/libexec/openssh/sftp-server` 的注释打开，重启 sshd 服务。

## 编译
```bash
sh build.sh
```

## 依赖
- 查阅 go.mod

## 注意
v0.X版本配置文件无法与v1.X版本兼容，请勿使用！