package bot

import (
	"bytes"
	"fmt"
	"github.com/indes/flowerss-bot/bot/fsm"
	"github.com/indes/flowerss-bot/config"
	"github.com/indes/flowerss-bot/model"
	"golang.org/x/net/proxy"
	tb "gopkg.in/tucnak/telebot.v2"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	botToken                             = config.BotToken
	socks5Proxy                          = config.Socks5
	UserState   map[int64]fsm.UserStatus = make(map[int64]fsm.UserStatus)

	B              *tb.Bot
	botSettingTmpl = `
订阅<b>设置</b>
[id] {{ .sub.ID}}
[标题] {{ .source.Title }}
[Link] {{.source.Link}}
[抓取更新] {{if ge .source.ErrorCount 100}}暂停{{else if lt .source.ErrorCount 100}}抓取中{{end}}
[通知] {{if eq .sub.EnableNotification 0}}关闭{{else if eq .sub.EnableNotification 1}}开启{{end}}
[Telegraph] {{if eq .sub.EnableTelegraph 0}}关闭{{else if eq .sub.EnableTelegraph 1}}开启{{end}}
`
)

var (
	toggleNoticeKey = tb.InlineButton{
		Unique: "toggle_notice",
		Text:   "切换通知",
	}

	toggleTelegraphKey = tb.InlineButton{
		Unique: "toggle_telegraph",
		Text:   "切换 Telegraph",
	}

	toggleEnabledKey = tb.InlineButton{
		Unique: "toggle_enabled",
		Text:   "抓取开关",
	}

	confirmButton = tb.InlineButton{
		Unique: "confirm",
		Text:   "确认",
	}

	cancelButton = tb.InlineButton{
		Unique: "cancel",
		Text:   "取消",
	}
)

func init() {
	poller := &tb.LongPoller{Timeout: 10 * time.Second}
	spamProtected := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {

		if !CheckAdmin(upd) {
			return false
		}

		return true
	})
	if socks5Proxy != "" {
		log.Printf("Bot Token: %s Proxy: %s\n", botToken, socks5Proxy)

		dialer, err := proxy.SOCKS5("tcp", socks5Proxy, nil, proxy.Direct)
		if err != nil {
			log.Fatal("Error creating dialer, aborting.")
		}

		httpTransport := &http.Transport{}
		httpClient := &http.Client{Transport: httpTransport}
		httpTransport.Dial = dialer.Dial

		// creat bot
		B, err = tb.NewBot(tb.Settings{
			Token:  botToken,
			Poller: spamProtected,
			Client: httpClient,
		})

		if err != nil {
			log.Fatal(err)
			return
		}
	} else {
		log.Printf("Bot Token: %s", botToken)

		var err error
		// creat bot
		B, err = tb.NewBot(tb.Settings{
			Token:  botToken,
			Poller: spamProtected,
		})
		if err != nil {
			log.Fatal(err)
			return
		}
	}

}

func toggleCtrlButtons(c *tb.Callback, action string) {
	msg := strings.Split(c.Message.Text, "\n")
	subID, err := strconv.Atoi(strings.Split(msg[1], " ")[1])
	if err != nil {
		_ = B.Respond(c, &tb.CallbackResponse{
			Text: "error",
		})
		return
	}
	sub, err := model.GetSubscribeByID(int(subID))
	if sub == nil || err != nil {
		_ = B.Respond(c, &tb.CallbackResponse{
			Text: "error",
		})
		return
	}

	source, _ := model.GetSourceById(int(sub.SourceID))
	t := template.New("setting template")
	_, _ = t.Parse(botSettingTmpl)
	feedSettingKeys := [][]tb.InlineButton{}

	switch action {
	case "toggleNoticeKey":
		err = sub.ToggleNotification()
	case "toggleTelegraphKey":
		err = sub.ToggleTelegraph()
	case "toggleEnabledKey":
		err = source.ToggleEnabled()
	}

	if err != nil {
		_ = B.Respond(c, &tb.CallbackResponse{
			Text: "error",
		})
		return
	}

	sub.Save()

	text := new(bytes.Buffer)
	if sub.EnableNotification == 1 {
		toggleNoticeKey.Text = "关闭通知"
	} else {
		toggleNoticeKey.Text = "开启通知"
	}
	if sub.EnableTelegraph == 1 {
		toggleTelegraphKey.Text = "关闭 Telegraph 转码"
	} else {
		toggleTelegraphKey.Text = "开启 Telegraph 转码"
	}
	if source.ErrorCount >= 100 {
		toggleEnabledKey.Text = "重启更新"
	} else {
		toggleEnabledKey.Text = "暂停更新"
	}

	feedSettingKeys = append(feedSettingKeys, []tb.InlineButton{toggleEnabledKey, toggleNoticeKey, toggleTelegraphKey})
	_ = t.Execute(text, map[string]interface{}{"source": source, "sub": sub})
	_ = B.Respond(c, &tb.CallbackResponse{
		Text: "修改成功",
	})
	_, _ = B.Edit(c.Message, text.String(), &tb.SendOptions{
		ParseMode: tb.ModeHTML,
	}, &tb.ReplyMarkup{
		InlineKeyboard: feedSettingKeys,
	})
}

//Start bot
func Start() {
	makeHandle()
	B.Start()
}

func makeHandle() {

	B.Handle(&toggleNoticeKey, func(c *tb.Callback) {
		toggleCtrlButtons(c, "toggleNoticeKey")
	})

	B.Handle(&toggleTelegraphKey, func(c *tb.Callback) {
		toggleCtrlButtons(c, "toggleTelegraphKey")
	})

	B.Handle(&toggleEnabledKey, func(c *tb.Callback) {
		toggleCtrlButtons(c, "toggleEnabledKey")
	})

	B.Handle(&confirmButton, func(c *tb.Callback) {

		mention := GetMentionFromMessage(c.Message)
		var msg string
		if mention == "" {
			success, fail, err := model.UnsubAllByUserID(int64(c.Sender.ID))
			if err != nil {
				msg = "退订失败"
			} else {
				msg = fmt.Sprintf("退订成功：%d\n退订失败：%d", success, fail)
			}

		} else {
			channelChat, err := B.ChatByID(mention)

			if err != nil {
				_, _ = B.Edit(c.Message, "error")
				return
			}

			if UserIsAdminChannel(c.Sender.ID, channelChat) {
				success, fail, err := model.UnsubAllByUserID(channelChat.ID)
				if err != nil {
					msg = "退订失败"

				} else {
					msg = fmt.Sprintf("退订成功：%d\n退订失败：%d", success, fail)
				}

			} else {
				msg = "非频道管理员无法执行此操作"
			}
		}

		_, _ = B.Edit(c.Message, msg)

	})

	B.Handle(&cancelButton, func(c *tb.Callback) {
		_, _ = B.Edit(c.Message, "操作取消")
	})

	B.Handle("/start", func(m *tb.Message) {
		user := model.FindOrInitUser(m.Chat.ID)
		log.Printf("/start %d", user.ID)
		_, _ = B.Send(m.Chat, fmt.Sprintf("欢迎使用！ /help  打开帮助菜单。"))
	})

	B.Handle("/export", func(m *tb.Message) {

		mention := GetMentionFromMessage(m)
		var sourceList []model.Source
		var err error
		if mention == "" {

			sourceList, err = model.GetSourcesByUserID(m.Chat.ID)
			if err != nil {
				log.Println(err.Error())
				_, _ = B.Send(m.Chat, fmt.Sprintf("导出失败"))
				return
			}
		} else {
			channelChat, err := B.ChatByID(mention)

			if err != nil {
				_, _ = B.Send(m.Chat, "error")
				return
			}

			adminList, err := B.AdminsOf(channelChat)
			if err != nil {
				_, _ = B.Send(m.Chat, "error")
				return
			}

			senderIsAdmin := false
			for _, admin := range adminList {
				if m.Sender.ID == admin.User.ID {
					senderIsAdmin = true
				}
			}

			if !senderIsAdmin {
				_, _ = B.Send(m.Chat, fmt.Sprintf("非频道管理员无法执行此操作"))
				return
			}

			sourceList, err = model.GetSourcesByUserID(channelChat.ID)
			if err != nil {
				log.Println(err.Error())
				_, _ = B.Send(m.Chat, fmt.Sprintf("导出失败"))
				return
			}
		}

		if len(sourceList) == 0 {
			_, _ = B.Send(m.Chat, fmt.Sprintf("订阅列表为空"))
			return
		}

		opmlStr, err := ToOPML(sourceList)

		if err != nil {
			_, _ = B.Send(m.Chat, fmt.Sprintf("导出失败"))
			return
		}
		opmlFile := &tb.Document{File: tb.FromReader(strings.NewReader(opmlStr))}
		opmlFile.FileName = fmt.Sprintf("subscriptions_%d.opml", time.Now().Unix())
		_, err = B.Send(m.Chat, opmlFile)

		if err != nil {
			_, _ = B.Send(m.Chat, fmt.Sprintf("导出失败"))
			log.Println("[export]", err)
		}

	})

	B.Handle("/sub", func(m *tb.Message) {
		url, mention := GetUrlAndMentionFromMessage(m)

		if mention == "" {
			if url != "" {
				registFeed(m.Chat, url)
			} else {
				_, err := B.Send(m.Chat, "请回复RSS URL", &tb.ReplyMarkup{ForceReply: true})
				if err == nil {
					UserState[m.Chat.ID] = fsm.Sub
				}
			}
		} else {
			if url != "" {
				FeedForChannelRegister(m, url, mention)
			} else {
				_, _ = B.Send(m.Chat, "频道订阅请使用' /sub @ChannelID URL ' 命令")
			}
		}

	})

	B.Handle("/list", func(m *tb.Message) {
		mention := GetMentionFromMessage(m)
		if mention != "" {
			channelChat, err := B.ChatByID(mention)
			if err != nil {
				_, _ = B.Send(m.Chat, "error")
				return
			}
			adminList, err := B.AdminsOf(channelChat)
			if err != nil {
				_, _ = B.Send(m.Chat, "error")
				return
			}

			senderIsAdmin := false
			for _, admin := range adminList {
				if m.Sender.ID == admin.User.ID {
					senderIsAdmin = true
				}
			}

			if !senderIsAdmin {
				_, _ = B.Send(m.Chat, fmt.Sprintf("非频道管理员无法执行此操作"))
				return
			}

			sources, _ := model.GetSourcesByUserID(channelChat.ID)
			message := fmt.Sprintf("频道 [%s](https://t.me/%s) 订阅列表：\n", channelChat.Title, channelChat.Username)
			if len(sources) == 0 {
				message = fmt.Sprintf("频道 [%s](https://t.me/%s) 订阅列表为空", channelChat.Title, channelChat.Username)
			} else {
				for index, source := range sources {
					message = message + fmt.Sprintf("[[%d]] [%s](%s)\n", index+1, source.Title, source.Link)
				}
			}

			_, _ = B.Send(m.Chat, message, &tb.SendOptions{
				DisableWebPagePreview: true,
				ParseMode:             tb.ModeMarkdown,
			})

		} else {
			sources, _ := model.GetSourcesByUserID(m.Chat.ID)
			message := "当前订阅列表：\n"
			if len(sources) == 0 {
				message = "订阅列表为空"
			} else {
				for index, source := range sources {
					message = message + fmt.Sprintf("[[%d]] [%s](%s)\n", index+1, source.Title, source.Link)
				}
			}
			_, _ = B.Send(m.Chat, message, &tb.SendOptions{
				DisableWebPagePreview: true,
				ParseMode:             tb.ModeMarkdown,
			})
		}

	})

	B.Handle("/set", func(m *tb.Message) {

		sources, _ := model.GetSourcesByUserID(m.Chat.ID)

		if len(sources) <= 0 {
			_, _ = B.Send(m.Chat, "当前没有订阅源")
			return
		}
		var replyButton []tb.ReplyButton
		replyKeys := [][]tb.ReplyButton{}
		for _, source := range sources {
			// 添加按钮
			text := fmt.Sprintf("%s %s", source.Title, source.Link)
			replyButton = []tb.ReplyButton{
				tb.ReplyButton{Text: text},
			}
			replyKeys = append(replyKeys, replyButton)
		}
		_, err := B.Send(m.Chat, "请选择你要设置的源", &tb.ReplyMarkup{
			ForceReply:    true,
			ReplyKeyboard: replyKeys,
		})

		if err == nil {
			UserState[m.Chat.ID] = fsm.Set
		}

	})

	B.Handle("/unsub", func(m *tb.Message) {

		url, mention := GetUrlAndMentionFromMessage(m)

		if mention == "" {
			if url != "" {
				//Unsub by url
				source, _ := model.GetSourceByUrl(url)
				if source == nil {
					_, _ = B.Send(m.Chat, "未订阅该RSS源")
				} else {
					err := model.UnsubByUserIDAndSource(m.Chat.ID, source)
					if err == nil {
						_, _ = B.Send(
							m.Chat,
							fmt.Sprintf("[%s](%s) 退订成功！", source.Title, source.Link),
							&tb.SendOptions{
								DisableWebPagePreview: true,
								ParseMode:             tb.ModeMarkdown,
							},
						)
						log.Printf("%d unsubscribe [%d]%s %s", m.Chat.ID, source.ID, source.Title, source.Link)
					} else {
						_, err = B.Send(m.Chat, err.Error())
					}
				}
			} else {
				//Unsub by button
				sources, _ := model.GetSourcesByUserID(m.Chat.ID)
				if len(sources) > 0 {
					var replyButton []tb.ReplyButton
					replyKeys := [][]tb.ReplyButton{}
					for _, source := range sources {
						// 添加按钮
						text := fmt.Sprintf("[%d] %s", source.ID, source.Title)
						replyButton = []tb.ReplyButton{
							tb.ReplyButton{Text: text},
						}

						replyKeys = append(replyKeys, replyButton)
					}
					_, err := B.Send(m.Chat, "请选择你要退订的源", &tb.ReplyMarkup{
						ForceReply:    true,
						ReplyKeyboard: replyKeys,
					})

					if err == nil {
						UserState[m.Chat.ID] = fsm.UnSub
					}
				} else {
					_, _ = B.Send(m.Chat, "当前没有订阅源")
				}

			}
		} else {
			if url != "" {
				channelChat, err := B.ChatByID(mention)
				if err != nil {
					_, _ = B.Send(m.Chat, "error")
					return
				}
				adminList, err := B.AdminsOf(channelChat)
				if err != nil {
					_, _ = B.Send(m.Chat, "error")
					return
				}

				senderIsAdmin := false
				for _, admin := range adminList {
					if m.Sender.ID == admin.User.ID {
						senderIsAdmin = true
					}
				}

				if !senderIsAdmin {
					_, _ = B.Send(m.Chat, fmt.Sprintf("非频道管理员无法执行此操作"))
					return
				}

				source, _ := model.GetSourceByUrl(url)
				sub, err := model.GetSubByUserIDAndURL(channelChat.ID, url)

				if err != nil {
					if err.Error() == "record not found" {
						_, _ = B.Send(
							m.Chat,
							fmt.Sprintf("频道 [%s](https://t.me/%s) 未订阅该RSS源", channelChat.Title, channelChat.Username),
							&tb.SendOptions{
								DisableWebPagePreview: true,
								ParseMode:             tb.ModeMarkdown,
							},
						)

					} else {
						_, _ = B.Send(m.Chat, "退订失败")
					}
					return

				} else {

					err := sub.Unsub()
					if err == nil {
						_, _ = B.Send(
							m.Chat,
							fmt.Sprintf("频道 [%s](https://t.me/%s) 退订 [%s](%s) 成功", channelChat.Title, channelChat.Username, source.Title, source.Link),
							&tb.SendOptions{
								DisableWebPagePreview: true,
								ParseMode:             tb.ModeMarkdown,
							},
						)
						log.Printf("%d for [%s]%s unsubscribe [%d]%s %s", m.Chat.ID, source.ID, source.Title, source.Link)
					} else {
						_, err = B.Send(m.Chat, err.Error())
					}
					return
				}

			} else {
				_, _ = B.Send(m.Chat, "频道退订请使用' /unsub @ChannelID URL ' 命令")
			}
		}

	})

	B.Handle("/unsuball", func(m *tb.Message) {
		mention := GetMentionFromMessage(m)
		confirmKeys := [][]tb.InlineButton{}
		confirmKeys = append(confirmKeys, []tb.InlineButton{confirmButton, cancelButton})
		var msg string

		if mention == "" {
			msg = "是否退订当前用户的所有订阅？"
		} else {
			msg = fmt.Sprintf("%s 是否退订该 Channel 所有订阅？", mention)
		}

		_, _ = B.Send(
			m.Chat,
			msg,
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
			}, &tb.ReplyMarkup{
				InlineKeyboard: confirmKeys,
			},
		)
	})

	B.Handle("/ping", func(m *tb.Message) {

		_, _ = B.Send(m.Chat, "pong")
	})

	B.Handle("/help", func(m *tb.Message) {
		message := `
使用命令：

/sub - 添加订阅
/unsub - 取消订阅
/list - 查看当前订阅
/set - 设置推送模式
/import - 导入OPML文件
/export - 导出 OPML 文件
/unsuball - 取消所有订阅
/help - 帮助菜单
`
		_, _ = B.Send(m.Chat, message)
	})

	B.Handle("/import", func(m *tb.Message) {
		message := `请直接发送OPML文件。`
		_, _ = B.Send(m.Chat, message)
	})

	B.Handle(tb.OnText, func(m *tb.Message) {
		switch UserState[m.Chat.ID] {
		case fsm.UnSub:
			{
				str := strings.Split(m.Text, " ")

				if len(str) < 2 && (strings.HasPrefix(str[0], "[") && strings.HasSuffix(str[0], "]")) {
					_, _ = B.Send(m.Chat, "请选择正确的指令！")
				} else {

					var sourceId int
					if _, err := fmt.Sscanf(str[0], "[%d]", &sourceId); err != nil {
						_, _ = B.Send(m.Chat, "请选择正确的指令！")
						return
					}

					source, err := model.GetSourceById(sourceId)

					if err != nil {
						_, _ = B.Send(m.Chat, "请选择正确的指令！")
						return
					}

					err = model.UnsubByUserIDAndSource(m.Chat.ID, source)

					if err != nil {
						_, _ = B.Send(m.Chat, "请选择正确的指令！")
						return
					} else {
						_, _ = B.Send(
							m.Chat,
							fmt.Sprintf("[%s](%s) 退订成功", source.Title, source.Link),
							&tb.SendOptions{
								ParseMode: tb.ModeMarkdown,
							}, &tb.ReplyMarkup{
								ReplyKeyboardRemove: true,
							},
						)
						UserState[m.Chat.ID] = fsm.None
						return
					}
				}
			}

		case fsm.Sub:
			{
				url := strings.Split(m.Text, " ")
				if !CheckUrl(url[0]) {
					_, _ = B.Send(m.Chat, "请回复正确的URL", &tb.ReplyMarkup{ForceReply: true})
					return
				}

				registFeed(m.Chat, url[0])
				UserState[m.Chat.ID] = fsm.None
			}

		case fsm.Set:
			{

				str := strings.Split(m.Text, " ")
				url := str[len(str)-1]
				if len(str) != 2 && !CheckUrl(url) {
					_, _ = B.Send(m.Chat, "请选择正确的指令！")
				} else {
					source, err := model.GetSourceByUrl(url)

					if err != nil {
						_, _ = B.Send(m.Chat, "请选择正确的指令！")
						return
					}
					sub, err := model.GetSubscribeByUserIDAndSourceID(m.Chat.ID, source.ID)
					if err != nil {
						_, _ = B.Send(m.Chat, "请选择正确的指令！")
						return
					}
					t := template.New("setting template")
					_, _ = t.Parse(botSettingTmpl)

					feedSettingKeys := [][]tb.InlineButton{}

					text := new(bytes.Buffer)
					if sub.EnableNotification == 1 {
						toggleNoticeKey.Text = "关闭通知"
					} else {
						toggleNoticeKey.Text = "开启通知"
					}

					if sub.EnableTelegraph == 1 {
						toggleTelegraphKey.Text = "关闭 Telegraph 转码"
					} else {
						toggleTelegraphKey.Text = "开启 Telegraph 转码"
					}

					if source.ErrorCount >= 100 {
						toggleEnabledKey.Text = "重启更新"
					} else {
						toggleEnabledKey.Text = "暂停更新"
					}

					feedSettingKeys = append(feedSettingKeys, []tb.InlineButton{toggleEnabledKey, toggleNoticeKey, toggleTelegraphKey})
					_ = t.Execute(text, map[string]interface{}{"source": source, "sub": sub})

					// send null message to remove old keyboard
					delKeyMessage, err := B.Send(m.Chat, "processing", &tb.ReplyMarkup{ReplyKeyboardRemove: true})
					err = B.Delete(delKeyMessage)

					_, _ = B.Send(
						m.Chat,
						text.String(),
						&tb.SendOptions{
							ParseMode: tb.ModeHTML,
						}, &tb.ReplyMarkup{
							InlineKeyboard: feedSettingKeys,
						},
					)
					UserState[m.Chat.ID] = fsm.None
				}
			}
		}
	})

	B.Handle(tb.OnDocument, func(m *tb.Message) {

		if m.Document.MIME == "text/x-opml+xml" {

			url, _ := B.FileURLByID(m.Document.FileID)
			opml, err := GetOPMLByURL(url)

			if err != nil {
				if err.Error() == "fetch opml file error" {
					_, _ = B.Send(m.Chat,
						"下载 OPML 文件失败，请检查 bot 服务器能否正常连接至 telegram 服务器或过段时间再尝试导入。")

				} else {
					_, _ = B.Send(m.Chat, "如果需要导入订阅，请发送正确的 OPML 文件。")
				}
				return
			}

			message, _ := B.Send(m.Chat, "处理中，请稍后。")
			outlines, _ := opml.GetFlattenOutlines()
			var failImportList []Outline
			var successImportList []Outline

			for _, outline := range outlines {
				source, err := model.FindOrNewSourceByUrl(outline.XMLURL)
				if err != nil {
					failImportList = append(failImportList, outline)
					continue
				}
				err = model.RegistFeed(m.Chat.ID, source.ID)
				if err != nil {
					failImportList = append(failImportList, outline)
					continue
				}
				log.Printf("%d subscribe [%d]%s %s", m.Chat.ID, source.ID, source.Title, source.Link)
				successImportList = append(successImportList, outline)
			}

			importReport := fmt.Sprintf("<b>导入成功：%d，导入失败：%d</b>", len(successImportList), len(failImportList))
			if len(successImportList) != 0 {
				successReport := "\n\n<b>以下订阅源导入成功:</b>"
				for i, line := range successImportList {
					if line.Text != "" {
						successReport += fmt.Sprintf("\n[%d] <a href=\"%s\">%s</a>", i+1, line.XMLURL, line.Text)
					} else {
						successReport += fmt.Sprintf("\n[%d] %s", i+1, line.XMLURL)
					}
				}
				importReport += successReport
			}

			if len(failImportList) != 0 {
				failReport := "\n\n<b>以下订阅源导入失败:</b>"
				for i, line := range failImportList {
					if line.Text != "" {
						failReport += fmt.Sprintf("\n[%d] <a href=\"%s\">%s</a>", i+1, line.XMLURL, line.Text)
					} else {
						failReport += fmt.Sprintf("\n[%d] %s", i+1, line.XMLURL)
					}
				}
				importReport += failReport
			}
			_, err = B.Edit(message, importReport, &tb.SendOptions{
				DisableWebPagePreview: true,
				ParseMode:             tb.ModeHTML,
			})

			if err != nil {
				log.Println(err.Error())
			}

		} else {
			_, _ = B.Send(m.Chat, "如果需要导入订阅，请发送正确的OPML文件。")
		}

	})
}
