package gor

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/wendal/errors"
	"github.com/wendal/mustache"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// 编译整个网站
func Compile() error {
	var ctx mustache.Context // 渲染上下文
	var docCont *DocContent  // 文档内容,仅作为变量声明
	var str string           // 仅声明,以减少不一样的编译错误
	var err error            // 仅声明

	var layouts map[string]Mapper

	payload, err := BuildPlayload("./") // payload,核心上下文的主要部分,不可变
	if err != nil {
		log.Println("Build PayLoad FAIL!!")
		return err
	}

	payload_ctx := mustache.MakeContextDir(payload, ".tmp_partials/")
	themeName := FromCtx(payload_ctx, "site.config.theme").(string)

	if FromCtx(payload_ctx, "site.config.markdown.toc_title") != nil {
		TOC_TITLE = FromCtx(payload_ctx, "site.config.markdown.toc_title").(string)
	}

	os.Remove(".tmp_partials")
	copyDir("partials", ".tmp_partials")
	copyDir("themes/"+themeName+"/partials", ".tmp_partials")

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

	//mdParser = markdown.NewParser(&markdown.Extensions{Smart: true})
	helpers := make(map[string]mustache.SectionRenderFunc)
	ctxHelpers := make(map[string]func(interface{}) interface{})

	dynamicMapper := make(Mapper)
	topCtx := mustache.MakeContexts(payload_ctx, helpers, ctxHelpers, dynamicMapper)

	widgets, widget_assets, err := LoadWidgets(topCtx)
	if err != nil {
		return err
	}

	//log.Println(">>>", payload_ctx.Dir(), "?>", topCtx.Dir())

	BaiscHelpers(payload, helpers, topCtx)
	CtxHelpers(payload, ctxHelpers, topCtx)
	layouts = payload["layouts"].(map[string]Mapper)

	if len(widgets) > 0 {
		widget_assets += PrapareAssets(themeName, "widgets", topCtx)
	}

	base_path := payload["urls"].(map[string]string)["base_path"]
	CopyResources(base_path, themeName)

	// Render Pages
	pages := payload["db"].(map[string]interface{})["pages"].(map[string]Mapper)
	for id, page := range pages {
		docCont = page["_content"].(*DocContent)
		//top := make(map[string]interface{})
		dynamicMapper["current_page_id"] = id
		dynamicMapper["page"] = page
		dynamicMapper["assets"] = PrapareAssets(themeName, page.Layout(), topCtx) + widget_assets
		widgetCtx := PrapareWidgets(widgets, page, topCtx)
		ctx = mustache.MakeContexts(page, dynamicMapper, topCtx, widgetCtx)
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
		//top := make(map[string]interface{})
		dynamicMapper["current_page_id"] = id
		dynamicMapper["page"] = post
		dynamicMapper["assets"] = PrapareAssets(themeName, post.Layout(), topCtx) + widget_assets
		docCont = post["_content"].(*DocContent)
		widgetCtx := PrapareWidgets(widgets, post, topCtx)
		ctx = mustache.MakeContexts(post, dynamicMapper, topCtx, widgetCtx)

		str, err = RenderInLayout(docCont.Main, post.Layout(), layouts, ctx)
		if err != nil {
			return errors.New(id + ">" + err.Error())
		}

		WriteTo(post.Url(), str)
	}

	//我们还得把分页给解决了哦
	if paginatorCnf := FromCtx(topCtx, "site.config.paginator"); paginatorCnf != nil {
		var pgCnf Mapper
		pgCnf = paginatorCnf.(map[string]interface{})
		if _, ok := layouts[pgCnf.String("layout")]; ok {
			log.Println("Enable paginator")
			renderPaginator(pgCnf, layouts, topCtx, widgets)
		} else {
			log.Println("Layout Not Found", pgCnf.String("layout"))
		}
	}

	if Plugins != nil {
		for _, plugin := range Plugins {
			plugin.Exec(topCtx)
		}
	}

	log.Println("Done")
	return nil
}

func RenderInLayout(content string, layoutName string, layouts map[string]Mapper, ctx mustache.Context) (string, error) {
	//log.Println("Render Layout", layoutName, ">>", content, "<<END")
	ctx2 := make(map[string]string)
	ctx2["content"] = content
	layout := layouts[layoutName]
	if layout == nil {
		return "", errors.New("Not such Layout : " + layoutName)
	}
	//log.Println(layoutName, layout["_content"])
	buf := &bytes.Buffer{}
	err := layout["_content"].(*DocContent).TPL.Render(mustache.MakeContexts(ctx2, ctx), buf)
	if err != nil {
		return content, err
	}
	if layout.Layout() != "" {
		return RenderInLayout(buf.String(), layout.Layout(), layouts, ctx)
	}
	return buf.String(), nil
}

func BaiscHelpers(payload Mapper, helpers map[string]mustache.SectionRenderFunc, topCtx mustache.Context) {
	var err error
	chronological := FromCtx(topCtx, "db.posts.chronological").([]string)
	latest_size := int(FromCtx(topCtx, "site.config.posts.latest").(int64))
	dict := FromCtx(topCtx, "db.posts.dictionary").(map[string]Mapper)
	summary_lines := int(FromCtx(topCtx, "site.config.posts.summary_lines").(int64))
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

		current_page_id := FromCtx(ctx, "current_page_id")
		for id, page := range pages {
			top := map[string]interface{}{}
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
		names := make([]string, 0)[0:0]
		for name, _ := range categories {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			categorie := categories[name]
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
		names := make([]string, 0)[0:0]
		for name, _ := range tags {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			tag := tags[name]
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

		//current_page
		current_page_id := FromCtx(topCtx, "current_page_id")
		//log.Println("current_page_id", current_page_id)

		//log.Println(in)
		ids, ok := in.([]interface{})
		if !ok {
			log.Println("Not String Array?")
			return false
		}

		_pages := make([]Mapper, 0)
		for _, id := range ids {
			p := pages[id.(string)]
			if current_page_id != nil {
				p["is_active_page"] = id == current_page_id.(string)
				//log.Println("is_active_page", id == current_page_id.(string))
			}
			_pages = append(_pages, p)
		}
		return _pages
	}
	ctxHelper["to_posts"] = func(in interface{}) interface{} {
		ids, ok := in.([]string)
		if !ok {
			return false
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
				return false
			}
			post = dict[id]
		}
		index := post_id_map[post.Id()]
		if index == post_list_sz-1 {
			return false
		}
		return dict[chronological[index+1]]
	}
	ctxHelper["previous"] = func(in interface{}) interface{} {
		post, ok := in.(Mapper)
		if !ok {
			id, ok := in.(string)
			if !ok {
				return false
			}
			post = dict[id]
		}
		index := post_id_map[post.Id()]
		//log.Println(index, post.Id())
		if index == 0 {
			return false
		}
		//log.Println("Post has previous", index, post.Id())
		return dict[chronological[index-1]]
	}

	ctxHelper["to_categories"] = func(in interface{}) interface{} {
		ids, ok := in.([]string)
		if !ok {
			log.Println("BAD to_categories")
			return false
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
			return false
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
	//mdParser := markdown.NewParser(&markdown.Extensions{Smart: true})
	str, err := mustache.RenderString(content, ctx)
	if err != nil {
		log.Println("Error When Parse >> " + id)
		return str, err
	}
	if strings.HasSuffix(id, ".md") || strings.HasSuffix(id, ".markdown") {
		//log.Println("R: MD : " + id)
		str = MarkdownToHtml(str)
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

	url = DecodePathInfo(url)

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
					assets = append(assets, fmt.Sprintf("<link href=\"%s/%s\" type=\"text/css\" rel=\"stylesheet\" media=\"all\">", "/assets/"+theme+"/widgets", stylesheet))
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
					assets = append(assets, fmt.Sprintf("<script src=\"%s/%s\"></script>", theme_javascripts_path, javascript))
				}
			}
		case map[string]interface{}:
			for widgetName, _javascript := range javascripts.(map[string]interface{}) {
				javascript := widgetName + "/javascripts/" + _javascript.(string)
				if strings.HasPrefix(javascript, "http://") || strings.HasPrefix(javascript, "https:") {
					assets = append(assets, fmt.Sprintf("<script src=\"%s\"></script>", javascript))
				} else {
					assets = append(assets, fmt.Sprintf("<script src=\"%s/%s\"> </script>", "/assets/"+theme+"/widgets/", javascript))
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

func CopyResources(base_path string, themeName string) {
	copyDir("others", "compiled"+base_path)
	copyDir("media", "compiled"+base_path+"assets/media")
	copyDir("themes/"+themeName, "compiled"+base_path+"/assets/"+themeName)
	copyDir("widgets", "compiled"+base_path+"/assets/widgets")
}

func copyDir(src string, target string) error {
	//log.Println("From", src, "To", target)
	fst, err := os.Stat(src)
	if err != nil {
		//log.Println(err)
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
		if finfo.Name() == "config.yml" {
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
		f2, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
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
	readUntil := ""
	for lines > 0 {
		line, _ := r.ReadString('\n')
		dst += line
		lines--
		if strings.Trim(line, "\r\n\t ") == "```" {
			if readUntil == "" {
				readUntil = "```"
			} else {
				readUntil = ""
			}
		}
		if lines == 0 {
			var err error
			for readUntil != strings.Trim(line, "\r\n\t ") {
				line, err = r.ReadString('\n')
				dst += line
				if err != nil {
					break
				}
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
	str = strings.Replace(str, TOC_MARKUP, "", 1)
	return MarkdownToHtml(str)
}

func renderPaginator(pgCnf Mapper, layouts map[string]Mapper, topCtx mustache.Context, widgets []Widget) {
	base_path := FromCtx(topCtx, "urls.base_path").(string)
	summary_lines := int(FromCtx(topCtx, "site.config.posts.summary_lines").(int64))
	per_page := pgCnf.Int("per_page")
	if per_page < 2 {
		per_page = 2
	} else if per_page > 100 {
		per_page = 100
	}
	namespace := pgCnf.String("namespace")
	if namespace == "" {
		namespace = "/posts/"
	}
	layout := pgCnf.String("layout")

	posts_ctx := make(Mapper)

	chronological, _ := FromCtx(topCtx, "db.posts.chronological").([]string)
	dictionary, _ := FromCtx(topCtx, "db.posts.dictionary").(map[string]Mapper)
	siteTitle, _ := FromCtx(topCtx, "site.title").(string)

	page_count := len(chronological)/per_page + 1
	if len(chronological)%per_page == 0 {
		page_count--
	}

	paginator_navigation := make([]Mapper, page_count)
	for i := 0; i < len(paginator_navigation); i++ {
		pn := make(Mapper)
		pn["page_number"] = i + 1
		pn["name"] = fmt.Sprintf("%d", i+1)
		pn["url"] = fmt.Sprintf("%s%d/", base_path+namespace, i+1)
		pn["is_active_page"] = false
		paginator_navigation[i] = pn
	}

	posts_ctx["paginator_navigation"] = paginator_navigation

	one_page := make([]Mapper, 0)
	current_page_number := 0
	log.Println("Total posts: ", len(chronological))
	for i, post_id := range chronological {
		if i != 0 && i%per_page == 0 {
			current_page_number++
			//log.Printf("Rendering page #%d with %d posts", current_page_number, len(one_page))
			posts_ctx["current_page_number"] = current_page_number
			posts_ctx["paginator"] = one_page
			if current_page_number >= 2 {
				paginator_navigation[current_page_number-2]["is_active_page"] = false
			}
			paginator_navigation[current_page_number-1]["is_active_page"] = true
			widgetCtx := PrapareWidgets(widgets, make(Mapper), topCtx)
			renderOnePager(paginator_navigation[current_page_number-1].String("url"), layout, layouts,
				mustache.MakeContexts(map[string]interface{}{"posts": posts_ctx,
					"page": map[string]interface{}{"title": fmt.Sprintf("%s Page %d", siteTitle, current_page_number)}}, topCtx, widgetCtx))
			one_page = one_page[0:0]
		}
		post := dictionary[post_id]
		post["summary"] = MakeSummary(post, summary_lines, topCtx)
		one_page = append(one_page, post)
	}
	if len(one_page) > 0 {
		current_page_number++
		//log.Printf("Rendering page #%d with %d post(s)", current_page_number, len(one_page))
		posts_ctx["current_page_number"] = current_page_number
		posts_ctx["paginator"] = one_page
		if current_page_number >= 2 {
			paginator_navigation[current_page_number-2]["is_active_page"] = false
		}
		paginator_navigation[current_page_number-1]["is_active_page"] = true
		m := make(Mapper)
		widgetCtx := PrapareWidgets(widgets, m, topCtx)
		renderOnePager(paginator_navigation[current_page_number-1].String("url"), layout, layouts,
			mustache.MakeContexts(map[string]interface{}{"posts": posts_ctx,
				"page": map[string]interface{}{"title": fmt.Sprintf("%s Page %d", siteTitle, current_page_number)}}, topCtx, widgetCtx))
	}
}

func renderOnePager(url string, layoutName string, layouts map[string]Mapper, ctx mustache.Context) {
	str, err := RenderInLayout("", layoutName, layouts, ctx)
	if err != nil {
		log.Println("ERR: Pager ", url, err)
		return
	}
	if strings.HasSuffix(url, "/") {
		url += "/index.html"
	}
	WriteTo(url, str)
}
