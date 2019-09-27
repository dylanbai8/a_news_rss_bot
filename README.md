# flowerss bot

[![Build Status](https://travis-ci.org/indes/flowerss-bot.svg?branch=master)](https://travis-ci.org/indes/flowerss-bot)
[![Go Report Card](https://goreportcard.com/badge/github.com/indes/rssflow)](https://goreportcard.com/report/github.com/indes/flowerss-bot)
![GitHub](https://img.shields.io/github/license/indes/flowerss-bot.svg)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Findes%2Fflowerss-bot.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Findes%2Fflowerss-bot?ref=badge_shield)

DEMO: [https://t.me/rssflowbot](https://t.me/rssflowbot)  
[问题反馈群组](https://t.me/joinchat/FJ-cikd-yN1Bf1SxWbAKjw)

<img src="https://github.com/rssflow/img/raw/master/images/rssflow_demo.gif" width = "300"/>

## Features

- 常见的 RSS Bot 该有的功能
- 支持 Telegram 应用内 instant view
- 支持为 Group 和 Channel 订阅 RSS 消息
- 丰富的订阅设置

## 安装

### Docker 部署

```shell
docker run -d -v ~/data/flowerss:/var/flowerss indes/flowerss-bot -b <bot token> -t <telegraph token 可省略>
```

### 下载二进制

**由于 GoReleaser 不支持 Cgo，如果要使用 SQLite 做为数据库，请下载源码自行编译。**

从 [Releases](https://github.com/indes/flowerss-bot/releases) 页面下载对应的版本。

### 源码安装

```shell
git clone https://github.com/indes/flowerss-bot && cd flowerss-bot
make build
./flowerss-bot
```

## 配置

根据以下模板，新建 `config.yml` 文件。

```yml
bot_token: XXX
telegraph_token: xxxx
socks5: 127.0.0.1:1080
update_interval: 10
error_threshold: 100
mysql:
  host: 127.0.0.1
  port: 3306
  user: user
  password: pwd
  database: flowerss
sqlite:
  path: ./data.db
```

配置说明：

| 配置项          | 含义                                      | 必填                                       |
| --------------- | ----------------------------------------- | ------------------------------------------ |
| bot_token       | Telegram Bot Token                        | 必填                                       |
| telegraph_token | Telegraph Token, 用于转存原文到 Telegraph | 可忽略（不转存原文到 Telegraph ）          |
| update_interval | RSS 源扫描间隔（分钟）                    | 可忽略（默认 10）                          |
| error_threshold | 源最大出错次数                    | 可忽略（默认 100）                          |
| socks5          | 用于无法正常 Telegram API 的环境          | 可忽略（能正常连接上 Telegram API 服务器） |
| mysql           | MySQL 数据库配置                          | 可忽略（使用 SQLite ）                     |
| sqlite          | SQLite 配置                               | 可忽略（已配置mysql时，该项失效）          |

### Telegraph Token 申请

```shell
curl https://api.telegra.ph/createAccount?short_name=flowerss&author_name=flowerss&author_url=https://github.com/indes/flowerss-bot
```

返回的 JSON 中 access_token 字段值即为 Telegraph Token

## 使用

命令：

```shell
/sub [url] 订阅（url 为可选）
/unsub [url] 取消订阅（url 为可选）
/list 查看当前订阅
/set 设置订阅
/import 导入OPML文件
/help 帮助
```

### Channel 订阅使用方法

1. 将 Bot 添加为 Channel 管理员
2. 发送相关命令给 Bot

Channel 订阅支持的命令：

```
/sub @ChannelID [url] 订阅
/unsub @ChannelID [url] 取消订阅
/list @ChannelID 查看当前订阅
```

**ChannelID 只有设置为 Public Channel 才有。如果是 Private Channel，可以暂时设置为 Public，订阅完成后改为 Private，不影响 Bot 推送消息。**

例如要给 t.me/debug 频道订阅 [阮一峰的网络日志](http://www.ruanyifeng.com/blog/atom.xml) RSS 更新：

1. 将 Bot 添加到 debug 频道管理员列表中
2. 给 Bot 发送 `/sub @debug http://www.ruanyifeng.com/blog/atom.xml` 命令

### 问题反馈

如果你在使用过程中遇到问题，请提交 Issue，或者到[问题反馈群组](https://t.me/joinchat/FJ-cikd-yN1Bf1SxWbAKjw) 反馈。

## License

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Findes%2Fflowerss-bot.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Findes%2Fflowerss-bot?ref=badge_large)
