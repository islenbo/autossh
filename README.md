# autossh

一个SSH远程客户端，可一键登录远程服务器，主要用来弥补Mac/Linux Terminal SSH无法保存密码的不足。

![演示](https://raw.githubusercontent.com/islenbo/autossh/8456ea1e8cb82541018a4133227a257c70199e40/docs/images/ezgif-5-42b5117192fc.gif)

## Wiki
[Wiki](https://github.com/islenbo/autossh/wiki)

## 功能说明
- SSH 快速登录
- 支持 cp 命令文件/文件夹复制功能 `autossh cp source:/file target:/file`
- 支持自动更新检测功能 `autossh upgrade`
- 新增快捷登录功能 `autossh [序号/别名]`

## 安装
- Mac/Linux用户直接下载安装包，运行install脚本即可。
- Windows用户可手动编译，参考编译章节。

## 注意
v0.X版本配置文件无法与v1.X版本兼容，请勿使用！

## License
MIT
