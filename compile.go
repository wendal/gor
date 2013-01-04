package gor

import (
	"bytes"
	"github.com/knieriem/markdown"
	"github.com/wendal/mustache"
	"io"
	"log"
)

func Compile() error {
	var payload Mapper
	//var ctx mustache.Context
	var docCont *DocContent
	//var tpl *mustache.Template
	var str string
	var err error
	var mdParser *markdown.Parser
	var buf *bytes.Buffer
	payload, err = BuildPlayload()
	if err != nil {
		return err
	}

	mdParser = markdown.NewParser(&markdown.Extensions{Smart: true})

	helpers := BaiscHelpers(payload)

	// Render Pages
	pages := payload["db"].(map[string]interface{})["pages"].(map[string]Mapper)
	for id, page := range pages {
		top := make(map[string]interface{})
		top["current_page_id"] = id
		docCont = page["_content"].(*DocContent)
		str, err = mustache.RenderString(docCont.Source, helpers)
		//str, err = mustache.RenderString(docCont.Source, page, top, payload, helpers)
		if err != nil {
			log.Printf("Error: Page[%s] -> %s", id, err.Error())
			return err
		}
		buf = bytes.NewBuffer(nil)
		mdParser.Markdown(bytes.NewBufferString(str), markdown.ToHTML(buf))
		pageContent := buf.String()
		log.Println("\n", pageContent)
	}

	// Render Posts

	// Render rss?

	log.Println("Done")
	return nil
}

func BaiscHelpers(payload Mapper) (helpers map[string]mustache.SectionRenderFunc) {
	helpers = make(map[string]mustache.SectionRenderFunc)
	helpers["posts_latest"] = func(nodes []mustache.Node, ctx mustache.Context, w io.Writer) error {
		log.Println("Good")
		return nil
	}
	return
}
