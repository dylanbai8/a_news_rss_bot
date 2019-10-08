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
		"<hr><p>本文章抓取自RSS，版权归源站点所有。</p><p>查看原文：<a href=\"%s\">%s - %s</p><hr><p><a href=\"https://t.me/ideahub_ml\">[点击] 加入书友群 1.5TB电子书资源 @ideahub_ml</a></p><p><a href=\"https://t.me/rss_news_list\">[点击] 全网福利|薅羊毛·省钱中心 @rss_news_list</a></p><hr><p><a href=\"https://t.me/lutouzhongwen_rss\">[点击] 路透中文网 @lutouzhongwen_rss</a></p><p><a href=\"https://t.me/niuyueshibao_rss\">[点击] 纽约时报中文网 @niuyueshibao_rss</a></p><p><a href=\"https://t.me/meiguozhiyin_rss\">[点击] 美国之音 @meiguozhiyin_rss</a></p><p><a href=\"https://t.me/zhihuribao_rss\">[点击] 知乎日报 @zhihuribao_rss</a></p><p><a href=\"https://t.me/bbczhongwen_rss\">[点击] BBC中文 @bbczhongwen_rss</a></p><p><a href=\"https://t.me/ftzhongwen_rss\">[点击] FT中文网 @ftzhongwen_rss</a></p><p><a href=\"https://t.me/shuangyunews_rss\">[点击] 纽约时报 ChinaDaily 双语新闻 @shuangyunews_rss</a></p><hr><p><a href=\"https://t.me/rfi_rss\">[点击] 法国 国际广播电台 @rfi_rss</a></p><p><a href=\"https://t.me/dw_rss\">[点击] 德国 德国之声 @dw_rss</a></p><p><a href=\"https://t.me/abc_rss\">[点击] 澳大利亚 广播公司 @abc_rss</a></p><p><a href=\"https://t.me/ru_rss\">[点击] 俄罗斯 卫星通讯社 @ru_rss</a></p><p><a href=\"https://t.me/sg_rss\">[点击] 新加坡 联合早报 @sg_rss</a></p><p><a href=\"https://t.me/korea_rss\">[点击] 韩国 朝鲜日报 中央日报 @korea_rss</a></p><p><a href=\"https://t.me/jp_rss\">[点击] 日本 共同网 朝日新闻 日经中文网 @jp_rss</a></p><p><a href=\"https://t.me/ttww_rss\">[点击] 台湾 中央社 香港 苹果日报 @ttww_rss</a></p><hr><p><img src=\"https://jp.ijysc.com/adpic_1.png\"></p><p><a href=\"https://jp.ijysc.com/jump_1.html\">[点击] 跳转到商家页面</a></p><p><img src=\"https://jp.ijysc.com/adpic_2.png\"></p><p><a href=\"https://jp.ijysc.com/jump_2.html\">[点击] 跳转到商家页面</a></p>",
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
