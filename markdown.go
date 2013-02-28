package gor

import (
	//"bytes"
	// "github.com/knieriem/markdown"
	"github.com/russross/blackfriday"
	"log"
)

// 封装Markdown转换为Html的逻辑
func MarkdownToHtml(content string) (str string) {
	defer func() {
		e := recover()
		if e != nil {
			str = content
			log.Println("Render Markdown ERR:", e)
		}
	}()
	//注释掉的部分,是另外一个markdown渲染库,更传统一些
	/*
		mdParser := markdown.NewParser(&markdown.Extensions{Smart: true})
		buf := bytes.NewBuffer(nil)
		mdParser.Markdown(bytes.NewBufferString(content), markdown.ToHTML(buf))
		str = buf.String()
	*/
	str = string(blackfriday.MarkdownCommon([]byte(content)))
	return
}
