package gor

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wendal/mustache"
)

// 全局插件列表
var Plugins []Plugin

func init() {
	// 载入默认的插件
	Plugins = make([]Plugin, 3)
	Plugins[0] = &RssPlugin{}
	Plugins[1] = &SitemapPlugin{}
	Plugins[2] = &JekyllOff{}
}

// 插件本身应该是线程安全的
type Plugin interface {
	Exec(mustache.Context)
}

//--------------------------------------------------------
// RSS 全文输出, 当前仅支持全部输出
type RssPlugin struct{}

type Rss struct {
	Version string      `xml:"version,attr"`
	Channel *RssChannel `xml:"channel"`
}

type RssChannel struct {
	Title   string    `xml:"title"`
	Link    string    `xml:"link"`
	PubDate string    `xml:"pubDate"`
	Items   []RssItem `xml:"item"`
}

type RssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func (*RssPlugin) Exec(topCtx mustache.Context) {
	base_path := FromCtx(topCtx, "urls.base_path").(string)
	title := FromCtx(topCtx, "site.title").(string)
	production_url := FromCtx(topCtx, "site.config.production_url").(string)
	pubDate := time.Now().Format(time.RFC822)
	post_ids := FromCtx(topCtx, "db.posts.chronological").([]string)
	posts := FromCtx(topCtx, "db.posts.dictionary").(map[string]Mapper)
	items := make([]RssItem, 0)
	for _, id := range post_ids {
		post := posts[id]
		item := RssItem{post.GetString("title"), production_url + post.Url(), post["_date"].(time.Time).Format("2006-01-02 03:04:05 +0800"), post["_content"].(*DocContent).Main}
		items = append(items, item)
	}
	rss := &Rss{"2.0", &RssChannel{title, production_url, pubDate, items}}
	f, err := os.OpenFile("compiled"+base_path+"rss.xml", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Println("ERR When Create RSS", err)
		return
	}
	defer f.Close()
	data, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		log.Println("ERR When Create RSS", err)
		return
	}
	// FUCK!! 官方的xml库极其弱智,无法为struct指定名字
	f.WriteString(`<?xml version="1.0"  encoding="UTF-8"?>` + "\n" + `<rss version="2.0">`)
	str := string(data)
	f.Write([]byte(str[len(`<rss version="2.0">`)+1 : len(str)-len("</rss>")]))
	f.WriteString("</rss>")
	f.Sync()
	return
}

//----------------------------------------------------------------------------------------------------
// 生成sitemap, 可以说已经完整实现

type SitemapPlugin struct{}

func (SitemapPlugin) Exec(topCtx mustache.Context) {
	base_path := FromCtx(topCtx, "urls.base_path").(string)
	f, err := os.OpenFile("compiled"+base_path+"sitemap.xml", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Println("Error when create sitemap", err)
		return
	}
	defer f.Close()

	//自行拼接XML比官方的xml包还靠谱

	f.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	f.WriteString("\n")
	f.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	f.WriteString("\n")

	production_url := FromCtx(topCtx, "site.config.production_url").(string)

	f.WriteString("\t<url>\n")
	f.WriteString("\t\t<loc>")
	xml.Escape(f, []byte(production_url+"/")) //够弱智不? 竟然要传入一个io.Reader
	f.WriteString("</loc>\n")
	f.WriteString("\t</url>\n")

	post_ids := FromCtx(topCtx, "db.posts.chronological").([]string)
	posts := FromCtx(topCtx, "db.posts.dictionary").(map[string]Mapper)
	for _, id := range post_ids {
		f.WriteString("\t<url>\n")
		post := posts[id]
		f.WriteString("\t\t<loc>")
		xml.Escape(f, []byte(production_url))
		xml.Escape(f, []byte(post.Url()))
		f.WriteString("</loc>\n")
		f.WriteString(fmt.Sprintf("\t\t<lastmod>%s</lastmod>\n", post["date"])) // 是否应该抹除呢? 考虑中
		f.WriteString("\t\t<changefreq>weekly</changefreq>\n")
		f.WriteString("\t</url>\n")
	}

	f.WriteString(`</urlset>`)
	f.Sync()
	// ~_~ 大功告成!
}

type JekyllOff struct{}

func (*JekyllOff) Exec(topCtx mustache.Context) {
	base_path := FromCtx(topCtx, "urls.base_path").(string)
	f, err := os.OpenFile("compiled"+base_path+".nojekyll", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Println("Error when create .nojekyll", err)
		return
	}
	defer f.Close()
	f.WriteString("Gor Here. http://github.com/wendal/gor")
}
