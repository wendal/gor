package gor

import (
	//"bytes"
	// "github.com/knieriem/markdown"
	"github.com/russross/blackfriday"
	"log"
)

func MarkdownToHtml(content string) (str string) {
	defer func() {
		e := recover()
		if e != nil {
			log.Println(e)
		}
	}()
	/*
		mdParser := markdown.NewParser(&markdown.Extensions{Smart: true})
		buf := bytes.NewBuffer(nil)
		mdParser.Markdown(bytes.NewBufferString(content), markdown.ToHTML(buf))
		str = buf.String()
	*/
	str = string(blackfriday.MarkdownCommon([]byte(content)))
	return
}
