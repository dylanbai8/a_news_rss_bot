package tgraph

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func PublishHtml(sourceTitle string, title string, rawLink string, html string) (string, error) {
	//html = fmt.Sprintf(
	//	"<p>本文章由 <a href=\"https://github.com/indes/flowerss-bot\">flowerss</a> 抓取自RSS，版权归<a href=\"\">源站点</a>所有。</p><hr>",
	//) + html + fmt.Sprintf(
	//	"<hr><p>本文章由 <a href=\"https://github.com/indes/flowerss-bot\">flowerss</a> 抓取自RSS，版权归<a href=\"\">源站点</a>所有。</p><p>查看原文：<a href=\"%s\">%s - %s</p>",
	//	rawLink,
	//	title,
	//	sourceTitle,
	//)

	html = html + fmt.Sprintf(
		"<hr><p>本文章抓取自RSS，版权归源站点所有。</p><p>查看原文：<a href=\"%s\">%s - %s</p><hr><p><a href=\"https://t.me/ideahub_ml\">[点击] 加入书友群 @ideahub_ml</a></p><hr><p><img src=\"https://jp.ijysc.com/adpic_1.png\"></p><p><a href=\"https://jp.ijysc.com/jump_1.html\">[点击] 跳转到商家页面</a></p><p><img src=\"https://jp.ijysc.com/adpic_2.png\"></p><p><a href=\"https://jp.ijysc.com/jump_2.html\">[点击] 跳转到商家页面</a></p>",
		rawLink,
		title,
		sourceTitle,
	)
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	client := clientPool[rand.Intn(len(clientPool))]

	if page, err := client.CreatePageWithHTML(title+" - "+sourceTitle, sourceTitle, rawLink, html, true); err == nil {
		log.Printf("Created telegraph page url: %s", page.URL)
		return page.URL, err
	} else {
		log.Printf("Create telegraph page error: %s", err)
		return "", nil
	}
}
