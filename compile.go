package gor

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/knieriem/markdown"
	"github.com/wendal/mustache"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Compile() error {
	var payload Mapper
	var ctx mustache.Context
	var docCont *DocContent
	//var tpl *mustache.Template
	var str string
	var err error
	//var mdParser *markdown.Parser
	//var buf *bytes.Buffer
	var layouts map[string]Mapper
	payload, err = BuildPlayload()
	if err != nil {
		return err
	}
	payload_ctx := mustache.MakeContextDir(payload, "partials/")

	db_posts_dict, _ := payload_ctx.Get("db.posts.dictionary")
	//log.Println(">>>>>>>>>>>>", len(db_posts_dict.Val.Interface().(map[string]Mapper)))
	for id, post := range db_posts_dict.Val.Interface().(map[string]Mapper) {
		_tmp, err := PrapreMainContent(id, post["_content"].(*DocContent).Source, payload_ctx)
		if err != nil {
			return err
		}
		post["_content"].(*DocContent).Main = _tmp
		//log.Fatal(_tmp)
	}
	themeName := FromCtx(payload_ctx, "site.config.theme").(string)

	//mdParser = markdown.NewParser(&markdown.Extensions{Smart: true})
	helpers := make(map[string]mustache.SectionRenderFunc)
	ctxHelpers := make(map[string]func(interface{}) interface{})

	dynamicCtx := make(Mapper)
	topCtx := mustache.MakeContexts(payload_ctx, helpers, ctxHelpers, map[string]Mapper{"dynamic": dynamicCtx})

	widgets, err := LoadWidgets(topCtx)
	if err != nil {
		return err
	}

	//log.Println(">>>", payload_ctx.Dir(), "?>", topCtx.Dir())

	BaiscHelpers(payload, helpers, topCtx)
	CtxHelpers(payload, ctxHelpers, topCtx)
	layouts = payload["layouts"].(map[string]Mapper)

	widget_assets := ""
	if len(widgets) > 0 {
		widget_assets = PrapareAssets(themeName, "widgets", topCtx)
	}

	CopyResources(themeName)

	// Render Pages
	pages := payload["db"].(map[string]interface{})["pages"].(map[string]Mapper)
	for id, page := range pages {
		docCont = page["_content"].(*DocContent)
		top := make(map[string]interface{})
		top["current_page_id"] = id
		top["page"] = page
		top["assets"] = PrapareAssets(themeName, page.Layout(), topCtx) + widget_assets
		widgetCtx := PrapareWidgets(widgets, page, topCtx)
		ctx = mustache.MakeContexts(page, top, topCtx, widgetCtx)
		//log.Println(">>", ctx.Dir(), topCtx.Dir())
		_tmp, err := PrapreMainContent(id, docCont.Source, ctx)
		if err != nil {
			return err
		}
		page["_content"].(*DocContent).Main = _tmp

		str, err = RenderInLayout(docCont.Main, page.Layout(), layouts, ctx)
		if err != nil {
			return errors.New(id + ">" + err.Error())
		}
		WriteTo(page.Url(), str)
	}

	// Render Posts
	for id, post := range db_posts_dict.Val.Interface().(map[string]Mapper) {
		top := make(map[string]interface{})
		top["current_page_id"] = id
		top["page"] = post
		top["assets"] = PrapareAssets(themeName, post.Layout(), topCtx) + widget_assets
		docCont = post["_content"].(*DocContent)
		widgetCtx := PrapareWidgets(widgets, post, topCtx)
		ctx = mustache.MakeContexts(post, top, topCtx, widgetCtx)

		str, err = RenderInLayout(docCont.Main, post.Layout(), layouts, ctx)
		if err != nil {
			return errors.New(id + ">" + err.Error())
		}

		WriteTo(post.Url(), str)
	}
	_ = str

	// Render rss?

	log.Println("Done")
	return nil
}

func RenderInLayout(content string, layoutName string, layouts map[string]Mapper, ctx mustache.Context) (string, error) {
	//log.Println("Render Layout", layoutName, ">>", content, "<<END")
	ctx2 := make(map[string]string)
	ctx2["content"] = content
	layout := layouts[layoutName]
	str, err := mustache.RenderString(layout["_content"].(*DocContent).Source, ctx2, ctx)
	if err != nil {
		return content, err
	}
	if layout.Layout() != "" {
		return RenderInLayout(str, layout.Layout(), layouts, ctx)
	}
	return str, nil
}

func BaiscHelpers(payload Mapper, helpers map[string]mustache.SectionRenderFunc, topCtx mustache.Context) {
	var err error
	chronological := FromCtx(topCtx, "db.posts.chronological").([]string)
	latest_size := FromCtx(topCtx, "site.config.posts.latest").(int)
	dict := FromCtx(topCtx, "db.posts.dictionary").(map[string]Mapper)
	summary_lines := FromCtx(topCtx, "site.config.posts.summary_lines").(int)
	latest_posts := make([]Mapper, 0)
	for _, id := range chronological {
		latest_posts = append(latest_posts, dict[id])
		if len(latest_posts) >= latest_size {
			break
		}
	}
	//-------------------------------
	helpers["posts_latest"] = func(nodes []mustache.Node, inverted bool, ctx mustache.Context, w io.Writer) error {
		if inverted {
			return nil
		}

		for _, post := range latest_posts {
			top := map[string]interface{}{}
			top["summary"] = MakeSummary(post, summary_lines, topCtx)
			//log.Println(top["summary"])
			top["content"] = post["_content"].(*DocContent).Main
			for _, node := range nodes {
				err = node.Render(mustache.MakeContexts(post, top, ctx), w)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	pages := FromCtx(topCtx, "db.pages").(map[string]Mapper)

	helpers["pages"] = func(nodes []mustache.Node, inverted bool, ctx mustache.Context, w io.Writer) error {
		if inverted {
			return nil
		}
		for id, page := range pages {
			top := map[string]interface{}{}
			current_page_id := FromCtx(ctx, "current_page_id")
			if current_page_id != nil {
				top["is_active_page"] = id == current_page_id.(string)
			}
			for _, node := range nodes {
				err = node.Render(mustache.MakeContexts(page, top, ctx), w)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	helpers["posts"] = func(nodes []mustache.Node, inverted bool, ctx mustache.Context, w io.Writer) error {
		if inverted {
			return nil
		}
		for _, post := range dict {
			for _, node := range nodes {
				err = node.Render(mustache.MakeContexts(post, ctx), w)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	categories := FromCtx(topCtx, "db.posts.categories").(map[string]*Catalog)
	helpers["categories"] = func(nodes []mustache.Node, inverted bool, ctx mustache.Context, w io.Writer) error {
		if inverted {
			return nil
		}
		//log.Println("Using #categories")
		for _, categorie := range categories {
			for _, node := range nodes {
				err = node.Render(mustache.MakeContexts(categorie, ctx), w)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	tags := FromCtx(topCtx, "db.posts.tags").(map[string]*Tag)
	helpers["tags"] = func(nodes []mustache.Node, inverted bool, ctx mustache.Context, w io.Writer) error {
		if inverted {
			return nil
		}
		for _, tag := range tags {
			for _, node := range nodes {
				err = node.Render(mustache.MakeContexts(tag, ctx), w)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	return
}

func CtxHelpers(payload Mapper, ctxHelper map[string]func(interface{}) interface{}, topCtx mustache.Context) {

	chronological := FromCtx(topCtx, "db.posts.chronological").([]string)
	categories := FromCtx(topCtx, "db.posts.categories").(map[string]*Catalog)
	tags := FromCtx(topCtx, "db.posts.tags").(map[string]*Tag)
	pages := FromCtx(topCtx, "db.pages").(map[string]Mapper)
	dict := FromCtx(topCtx, "db.posts.dictionary").(map[string]Mapper)
	post_list_sz := len(chronological)

	post_id_map := map[string]int{}
	for index, post_id := range chronological {
		post_id_map[post_id] = index
	}

	ctxHelper["to_pages"] = func(in interface{}) interface{} {
		//log.Println(in)
		ids, ok := in.([]interface{})
		if !ok {
			log.Println("Not String Array?")
			return make(Mapper)
		}

		_pages := make([]Mapper, 0)
		for _, id := range ids {
			_pages = append(_pages, pages[id.(string)])
		}
		return _pages
	}
	ctxHelper["to_posts"] = func(in interface{}) interface{} {
		ids, ok := in.([]string)
		if !ok {
			return make(Mapper)
		}
		_posts := make([]Mapper, 0)
		for _, id := range ids {
			_posts = append(_posts, dict[id])
		}
		return _posts
	}
	ctxHelper["next"] = func(in interface{}) interface{} {
		post, ok := in.(Mapper)
		if !ok {
			id, ok := in.(string)
			if !ok {
				return make(Mapper)
			}
			post = dict[id]
		}
		index := post_id_map[post.Id()]
		if index >= post_list_sz-1 {
			return make(Mapper)
		}
		return dict[chronological[index+1]]
	}
	ctxHelper["previous"] = func(in interface{}) interface{} {
		post, ok := in.(Mapper)
		if !ok {
			id, ok := in.(string)
			if !ok {
				return make(Mapper)
			}
			post = dict[id]
		}
		index := post_id_map[post.Id()]
		if index-1 < 0 {
			return make(Mapper)
		}
		return dict[chronological[index-1]]
	}

	ctxHelper["to_categories"] = func(in interface{}) interface{} {
		ids, ok := in.([]string)
		if !ok {
			log.Println("BAD to_categories")
			return make(Mapper)
		}
		_catalogs := make([]*Catalog, 0)
		for _, id := range ids {
			_catalogs = append(_catalogs, categories[id])
		}
		return _catalogs
	}

	ctxHelper["to_tags"] = func(in interface{}) interface{} {
		ids, ok := in.([]string)
		if !ok {
			return make(Mapper)
		}
		_tags := make([]*Tag, 0)
		for _, id := range ids {
			_tags = append(_tags, tags[id])
		}
		return _tags
	}

	return
}

func PrapreMainContent(id string, content string, ctx mustache.Context) (string, error) {
	mdParser := markdown.NewParser(&markdown.Extensions{Smart: true})
	str, err := mustache.RenderString(content, ctx)
	if err != nil {
		return str, err
	}
	if strings.HasSuffix(id, ".md") || strings.HasSuffix(id, ".markdown") {
		log.Println("R: MD : " + id)
		buf := bytes.NewBuffer(nil)
		mdParser.Markdown(bytes.NewBufferString(str), markdown.ToHTML(buf))
		str = buf.String()
	}
	return str, nil
}

func FromCtx(ctx mustache.Context, key string) interface{} {
	val, found := ctx.Get(key)
	if found {
		return val.Val.Interface()
	}
	return nil
}

func WriteTo(url string, content string) {
	if strings.HasSuffix(url, "/") {
		url = url + "index.html"
	} else if !strings.HasSuffix(url, ".html") {
		url = url + "/index.html"
	}
	dstPath := "compiled" + url
	os.MkdirAll(filepath.Dir(dstPath), os.ModePerm)
	ioutil.WriteFile(dstPath, []byte(content), os.ModePerm)
}

func PrapareAssets(theme string, layoutName string, topCtx mustache.Context) string {
	//themeCnf := FromCtx(topCtx, "theme").(map[string]interface{})
	urls := FromCtx(topCtx, "urls").(map[string]string)
	//theme_base_path := urls["theme"]
	//theme_media_path := urls["theme_media"]
	theme_javascripts_path := urls["theme_javascripts"]
	theme_stylesheets_path := urls["theme_stylesheets"]

	assets := make([]string, 0)

	stylesheets := FromCtx(topCtx, "theme.stylesheets."+layoutName)
	if stylesheets == nil && layoutName != "widgets" {
		stylesheets = FromCtx(topCtx, "theme.stylesheets.default")
	}
	if stylesheets != nil {
		switch stylesheets.(type) {
		case []interface{}:
			for _, _stylesheet := range stylesheets.([]interface{}) {
				stylesheet := _stylesheet.(string)
				if strings.HasPrefix(stylesheet, "http://") || strings.HasPrefix(stylesheet, "https:") {
					assets = append(assets, fmt.Sprintf("<link href=\"%s\" type=\"text/css\" rel=\"stylesheet\" media=\"all\">", stylesheet))
				} else {
					assets = append(assets, fmt.Sprintf("<link href=\"%s/%s\" type=\"text/css\" rel=\"stylesheet\" media=\"all\">", theme_stylesheets_path, stylesheet))
				}
			}
		case map[string]interface{}:
			for widgetName, _stylesheet := range stylesheets.(map[string]interface{}) {
				stylesheet := widgetName + "/stylesheets/" + _stylesheet.(string)
				if strings.HasPrefix(stylesheet, "http://") || strings.HasPrefix(stylesheet, "https:") {
					assets = append(assets, fmt.Sprintf("<link href=\"%s\" type=\"text/css\" rel=\"stylesheet\" media=\"all\">", stylesheet))
				} else {
					assets = append(assets, fmt.Sprintf("<link href=\"%s/%s\" type=\"text/css\" rel=\"stylesheet\" media=\"all\">", "/assets/twitter/widgets", stylesheet))
				}
			}
		}

	}

	javascripts := FromCtx(topCtx, "theme.javascripts."+layoutName)
	if javascripts == nil && layoutName != "widgets" {
		javascripts = FromCtx(topCtx, "theme.javascripts.default")
	}
	if javascripts != nil {
		switch javascripts.(type) {
		case []interface{}:
			for _, _javascript := range javascripts.([]interface{}) {
				javascript := _javascript.(string)
				if strings.HasPrefix(javascript, "http://") || strings.HasPrefix(javascript, "https:") {
					assets = append(assets, fmt.Sprintf("<script src=\"%s\"></script>", javascript))
				} else {
					assets = append(assets, fmt.Sprintf("<script src=\"%s/%s\"></script>", theme_javascripts_path+"/"+javascript))
				}
			}
		case map[string]interface{}:
			for widgetName, _javascript := range javascripts.(map[string]interface{}) {
				javascript := widgetName + "/javascripts/" + _javascript.(string)
				if strings.HasPrefix(javascript, "http://") || strings.HasPrefix(javascript, "https:") {
					assets = append(assets, fmt.Sprintf("<script src=\"%s\"></script>", javascript))
				} else {
					assets = append(assets, fmt.Sprintf("<script src=\"%s/%s\"></script>", "/assets/twitter/widgets/"+javascript))
				}
			}
		}
	}

	rs := ""
	for _, str := range assets {
		rs += str + "\n"
	}
	return rs
}

func CopyResources(themeName string) {
	copyDir("others", "compiled")
	copyDir("media", "compiled/assets/media")
	copyDir("themes/"+themeName, "compiled/assets/"+themeName)
}

func copyDir(src string, target string) error {
	//log.Println("From", src, "To", target)
	fst, err := os.Stat(src)
	if err != nil {
		log.Println(err)
		return err
	}
	if !fst.IsDir() {
		return nil
	}
	finfos, err := ioutil.ReadDir(src)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, finfo := range finfos {
		if strings.HasPrefix(finfo.Name(), ".") {
			continue
		}
		//log.Println(finfo.Name())
		dst := target + "/" + finfo.Name()
		if finfo.IsDir() {
			copyDir(src+"/"+finfo.Name(), dst)
			continue
		}

		os.MkdirAll(filepath.Dir(dst), os.ModePerm)
		f, err := os.Open(src + "/" + finfo.Name())
		if err != nil {
			log.Println(err)
			continue
		}
		defer f.Close()
		f2, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			log.Println(err)
			continue
		}
		defer f2.Close()
		io.Copy(f2, f)
		f.Close()
		f2.Close()
	}
	return nil
}

func MakeSummary(post Mapper, lines int, topCtx mustache.Context) string {
	content := post["_content"].(*DocContent).Source
	r := bufio.NewReader(bytes.NewBufferString(content))
	dst := ""
	for lines > 0 {
		line, _ := r.ReadString('\n')
		dst += line
		lines--
		if lines == 0 {
			for "" != strings.Trim(line, "\r\n\t ") {
				line, _ = r.ReadString('\n')
				dst += line
			}
		}
	}
	str, err := mustache.RenderString(dst, topCtx)
	if err != nil {
		log.Println("BAD Mustache after Summary cut!")
		str, err = mustache.RenderString(dst, topCtx)
		if err != nil {
			log.Println("BAD Mustache Summary?", err)
			str = post["_content"].(*DocContent).Main
		}
	}
	mdParser := markdown.NewParser(&markdown.Extensions{Smart: true})
	buf := bytes.NewBuffer(nil)
	mdParser.Markdown(bytes.NewBufferString(str), markdown.ToHTML(buf))
	return buf.String()
}
