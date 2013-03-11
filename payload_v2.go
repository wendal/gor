package gor

import (
	//"github.com/wendal/errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

//基于Struct的Payload,也就是初始加载的阶段

func MakePayLoad(root string) (webSite *WebSite, err error) {
	webSite = &WebSite{}
	err = nil

	//检查处理的根路径
	if root == "" {
		root = "."
	}
	root, err = filepath.Abs(root)
	if err != nil {
		panic(err)
	}
	root += "/"
	webSite.Root = root
	D("root dir >", root)

	webSite.LoadMainConfig()
	webSite.CheckMainConfig()
	webSite.MakeBasicURLs()

	webSite.FixPostPageConfigs()

	webSite.LoadPages()
	webSite.LoadPosts()

	return
}

// 读入核心配置
func (webSite *WebSite) LoadMainConfig() {

	cnf, err := ReadYml(webSite.Root + CONFIG_YAML)
	if err != nil {
		panic(err)
	}
	ToStruct(cnf, reflect.ValueOf(&webSite.TopCnf))
	D("config.yml", webSite.TopCnf)

	siteCnf, err := ReadYml(webSite.Root + SITE_YAML)
	if err != nil {
		panic(err)
	}
	ToStruct(siteCnf, reflect.ValueOf(&webSite.SiteCnf))
	D("site.yml", webSite.SiteCnf)
}

func (webSite *WebSite) CheckMainConfig() {

	themeName := webSite.TopCnf.Theme
	if themeName == "" { //必须有theme的设置
		panic("Miss theme config!")
	}
	// 载入theme的设置
	themeCnf, err := ReadYml(fmt.Sprintf("%s/themes/%s/theme.yml", webSite.Root, themeName))
	if err != nil {
		panic("No such theme ? " + themeName + " " + err.Error())
	}
	ToStruct(themeCnf, reflect.ValueOf(&webSite.ThemeCnf))

	webSite.Layouts = LoadLayouts(webSite.Root, themeName)
	if webSite.Layouts == nil || len(webSite.Layouts) == 0 {
		panic("Theme without any layout!!")
	}

	production_url := webSite.TopCnf.Production_url
	if production_url == "" {
		panic("Miss production_url")
	}
	if !strings.HasPrefix(production_url, "http://") && !strings.HasPrefix(production_url, "https://") {
		panic("production_url must start with https:// or http://")
	}

	// 域名保证是http/https开头,故,以下的处理,可以按https
	rootUrl := production_url
	pos := strings.Index(rootUrl[len("https://"):], "/")
	basePath := ""
	if pos == -1 {
		basePath = "/"
	} else {
		basePath = rootUrl[len("https://")+pos:]
		if !strings.HasSuffix(basePath, "/") {
			basePath += "/"
		}
	}
	webSite.RootURL = rootUrl
	webSite.BasePath = basePath
}

func (webSite *WebSite) MakeBasicURLs() {
	urls := make(map[string]string)
	urls["media"] = webSite.BasePath + "assets/media"
	urls["theme"] = webSite.BasePath + "assets/" + webSite.TopCnf.Theme
	urls["theme_media"] = urls["theme"] + "/media"
	urls["theme_javascripts"] = urls["theme"] + "/javascripts"
	urls["theme_stylesheets"] = urls["theme"] + "/stylesheets"
	urls["base_path"] = webSite.BasePath

	//需要重写?
	/*
		if site["urls"] != nil { //允许用户自定义基础URL,实现CDN等功能
			var site_url Mapper
			site_url = site["urls"].(map[string]interface{})
			for k, v := range site_url {
				urls[k] = v.(string)
			}
		}
	*/
	webSite.BaiseURLs = urls
}

func (webSite *WebSite) FixPostPageConfigs() {

	postsCnf := webSite.TopCnf.Posts
	if postsCnf.Permalink == "" {
		postsCnf.Permalink = "/:categories/:title/"
	}
	if postsCnf.Summary_lines < 5 {
		postsCnf.Summary_lines = 20
	}
	if postsCnf.Latest < 5 {
		postsCnf.Latest = 5
	}
	if postsCnf.Layout == "" {
		postsCnf.Layout = "post"
	}

	pagesCnf := webSite.TopCnf.Pages
	if pagesCnf.Layout == "" {
		pagesCnf.Layout = "page"
	}
	if pagesCnf.Permalink == "" {
		pagesCnf.Permalink = "pretty"
	}
}

func (webSite *WebSite) LoadPages() {
	pagesCnf := webSite.TopCnf.Pages
	pages, err := LoadPages(webSite.Root, pagesCnf.Exclude)
	if err != nil {
		return
	}
	// 构建导航信息(page列表),及整理page的配置信息
	navigation := make([]string, 0)
	webSite.Pages = make(map[string]PageBean)
	for page_id, page := range pages {
		pageBean := PageBean{}
		ToStruct(page, reflect.ValueOf(&pageBean))
		webSite.Pages[page_id] = pageBean

		if pageBean.Layout == "" {
			pageBean.Layout = pagesCnf.Layout
		}
		if pageBean.Permalink == "" {
			pageBean.Permalink = pagesCnf.Permalink
		}

		page_url := ""
		switch {
		case strings.HasSuffix(page_id, "index.html"):
			page_url = page_id[0 : len(page_id)-len("index.html")]
		case strings.HasSuffix(page_id, "index.md"):
			page_url = page_id[0 : len(page_id)-len("index.md")]
		default:
			page_url = page_id[0 : len(page_id)-len(filepath.Ext(page_id))]
			if pageBean.Title == "" && !strings.HasSuffix(page_url, "/") {
				pageBean.Title = strings.Title(filepath.Base(page_url))
			}
		}
		if strings.HasPrefix(page_url, "/") {
			pageBean.Url = webSite.BasePath + page_url[1:]
		} else {
			pageBean.Url = webSite.BasePath + page_url
		}

		if page_id != "index.html" && page_id != "index.md" {
			navigation = append(navigation, page_id)
		}
	}
	if webSite.SiteCnf.Navigation == nil || len(webSite.SiteCnf.Navigation) == 0 {
		webSite.SiteCnf.Navigation = navigation
	}
}

func (webSite *WebSite) LoadPosts() {
	postsCnf := webSite.TopCnf.Posts
	posts, err := LoadPosts(webSite.Root, postsCnf.Exclude)
	if err != nil {
		return
	}
	webSite.Posts = make(map[string]PostBean)
	for post_id, _post := range posts {
		postBean := PostBean{}
		ToStruct(_post, reflect.ValueOf(&postBean))
		webSite.Posts[post_id] = postBean

		if postBean.Layout == "" {
			postBean.Layout = postsCnf.Layout
		}
		if postBean.Permalink == "" {
			postBean.Permalink = postsCnf.Permalink
		}

		if postBean.Tags == nil {
			postBean.Tags = []string{}
		}
		if postBean.Categories == nil {
			postBean.Categories = []string{}
		}
	}

	// 整理post
	tags := make(map[string]*Tag)
	catalogs := make(map[string]*Catalog)
	chronological := make([]string, 0)
	collated := make(CollatedYears, 0)

	_collated := make(map[string]*CollatedYear)

	for id, post := range webSite.Posts {
		chronological = append(chronological, id)

		for _, _tag := range post.Tags {
			tag := tags[_tag]
			if tag == nil {
				tag = &Tag{0, _tag, make([]string, 0), "/tags/#" + EncodePathInfo(_tag) + "-ref"}
				tags[_tag] = tag
			}
			tag.Count += 1
			tag.Posts = append(tag.Posts, id)
		}

		for _, _catalog := range post.Categories {
			catalog := catalogs[_catalog]
			if catalog == nil {
				catalog = &Catalog{0, _catalog, make([]string, 0), "/categories/#" + EncodePathInfo(_catalog) + "-ref"}
				catalogs[_catalog] = catalog
			}
			catalog.Count += 1
			catalog.Posts = append(catalog.Posts, id)
		}

		_year, _month, _ := post._Date.Date()
		year := fmt.Sprintf("%v", _year)
		month := _month.String()

		_yearc := _collated[year]
		if _yearc == nil {
			_yearc = &CollatedYear{year, make([]*CollatedMonth, 0), make(map[string]*CollatedMonth)}
			_collated[year] = _yearc
		}
		_monthc := _yearc.months[month]
		if _monthc == nil {
			_monthc = &CollatedMonth{month, _month, []string{}}
			_yearc.months[month] = _monthc
			//log.Println("Add>>", year, month, post["id"])
		}
		_monthc.Posts = append(_monthc.Posts, id)

		post_map := make(Mapper)
		post_map["title"] = post.Title
		post_map["_date"] = post._Date
		post_map["id"] = post.Id
		post_map["categories"] = post.Categories
		post_map["permalink"] = post.Permalink
		CreatePostURL(nil, webSite.BasePath, post_map)
		post.Url = post_map["url"].(string)
	}
	_ = collated
	// TODO 需要重写排序方法
	/*

		sort.Sort(collated)

		for _, catalog := range catalogs {
			catalog.Posts = SortPosts(webSite.Posts, catalog.Posts)
		}
		for _, tag := range tags {
			tag.Posts = SortPosts(webSite.Posts, tag.Posts)
		}

		webSite.Tags = tags
		webSite.Catalogs = catalogs
		webSite.Chronological = SortPosts(webSite.Posts, chronological)
		webSite.Collated = collated
	*/
}
