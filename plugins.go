package gor

import (
	"encoding/xml"
	"github.com/wendal/mustache"
	"log"
	"os"
	"time"
)

type Plugin interface {
	Exec(mustache.Context)
}

//--------------------------------------------------------

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
	title := FromCtx(topCtx, "site.title").(string)
	production_url := FromCtx(topCtx, "site.config.production_url").(string)
	pubDate := time.Now().Format("2006-01-02 03:04:05 +0800")
	post_ids := FromCtx(topCtx, "db.posts.chronological").([]string)
	posts := FromCtx(topCtx, "db.posts.dictionary").(map[string]Mapper)
	items := make([]RssItem, 0)
	for _, id := range post_ids {
		post := posts[id]
		item := RssItem{post.GetString("title"), production_url + post.Url(), post["_date"].(time.Time).Format("2006-01-02 03:04:05 +0800"), post["_content"].(*DocContent).Main}
		items = append(items, item)
	}
	rss := &Rss{"2.0", &RssChannel{title, production_url, pubDate, items}}
	f, err := os.OpenFile("compiled/rss.xml", os.O_CREATE|os.O_WRONLY, os.ModePerm)
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
	f.WriteString(`<?xml version="1.0"?>` + "\n" + `<rss version="2.0">`)
	f.Write(data[len(`<rss version="2.0">`)+1 : len(data)-len("</Rss>")])
	f.WriteString("</rss>")
	f.Sync()
	return
}
