package gor

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	URL "net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wendal/mustache"
)

// 构建PayLoad
func BuildPlayload(root string) (payload map[string]interface{}, err error) {
	//检查处理的根路径
	if root == "" {
		root = "."
	}
	root, err = filepath.Abs(root)
	root += "/"
	log.Println("root=", root)

	// 开始读取配置
	payload = make(Mapper)
	err = nil
	var cnf Mapper
	var site Mapper

	//-----------------------------------
	cnf, err = ReadYml(root + CONFIG_YAML)
	if err != nil {
		log.Println("Fail to read ", root+CONFIG_YAML, err)
		return
	}
	site, err = ReadYml(root + SITE_YAML)
	if err != nil {
		log.Println("Fail to read ", root+SITE_YAML, err)
		return
	}

	site["config"] = cnf
	payload["site"] = site
	payload["data"] = site // for v2

	// Check site config!
	themeName := cnf.String("theme")
	if themeName == "" { //必须有theme的设置
		err = errors.New("Miss theme config!")
		return
	}
	cnf["theme"] = themeName // 保证是string

	payload["layouts"] = LoadLayouts(root, themeName)

	production_url := cnf.String("production_url")
	if production_url == "" {
		err = errors.New("Miss production_url")
		return
	}
	if !strings.HasPrefix(production_url, "http://") && !strings.HasPrefix(production_url, "https://") {
		err = errors.New("production_url must start with https:// or http://")
		return
	}
	cnf["production_url"] = production_url

	// 域名保证是http/https开头,故,以下的除了,可以按https
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

	// 读取theme的配置
	//---------------------------------
	themeCnf, err := ReadYml(fmt.Sprintf("%s/themes/%s/theme.yml", root, themeName))
	if err != nil {
		log.Println("No such theme ?", themeName, err)
		return
	}
	payload["theme"] = themeCnf

	// 设置基础URL
	//-------------------------------
	urls := make(map[string]string)
	urls["media"] = basePath + "assets/media"
	urls["theme"] = basePath + "assets/" + themeName
	urls["theme_media"] = urls["theme"] + "/media"
	urls["theme_javascripts"] = urls["theme"] + "/javascripts"
	urls["theme_stylesheets"] = urls["theme"] + "/stylesheets"
	urls["base_path"] = basePath

	if site["urls"] != nil { //允许用户自定义基础URL,实现CDN等功能
		var site_url Mapper
		site_url = site["urls"].(map[string]interface{})
		for k, v := range site_url {
			urls[k] = v.(string)
		}
	}

	payload["urls"] = urls

	//---------------------------------
	// 检查非必填,但必须存在的配置信息, 如果没有就自动补齐
	var cnf_posts Mapper
	if cnf["posts"] == nil {
		cnf_posts = make(map[string]interface{})
		cnf["posts"] = cnf_posts
	} else {
		cnf_posts = cnf["posts"].(map[string]interface{})
	}
	if cnf_posts.Permalink() == "" {
		cnf_posts["permalink"] = "/:categories/:title/"
	}
	if cnf_posts.GetInt("summary_lines") < 5 {
		cnf_posts["summary_lines"] = 20
	}
	if cnf_posts.GetInt("latest") < 5 {
		cnf_posts["latest"] = 10
	}
	if cnf_posts.Layout() == "" {
		cnf_posts["layout"] = "post"
	}
	if cnf_posts["exclude"] == nil {
		cnf_posts["exclude"] = ""
	}

	var cnf_pages Mapper
	if cnf["pages"] == nil {
		cnf_pages = make(Mapper)
		cnf["pages"] = cnf_pages
	} else {
		cnf_pages = cnf["pages"].(map[string]interface{})
	}
	if cnf_pages.Layout() == "" {
		cnf_pages["layout"] = "page"
	}
	if cnf_pages["exclude"] == nil {
		cnf_pages["exclude"] = ""
	}
	if cnf_pages.Permalink() == "" {
		cnf_pages["permalink"] = "pretty"
	}

	//---------------------------------
	post_layout_default := cnf_posts.Layout()
	page_layout_default := cnf_pages.Layout()

	post_permalink_default := cnf_posts.Permalink()
	page_permalink_default := cnf_pages.Permalink()

	// 为Page和Post补齐layout和permalink
	//---------------------------------

	db := make(map[string]interface{})
	payload["db"] = db

	pages, err := LoadPages(root, cnf_pages.String("exclude"))
	if err != nil {
		return
	}
	db["pages"] = pages

	// 构建导航信息(page列表)
	navigation := make([]string, 0)

	for page_id, page := range pages {
		if page.Layout() == "" {
			page["layout"] = page_layout_default
		}
		if page.Permalink() == "" {
			page["permalink"] = page_permalink_default
		}

		page_url := ""
		switch {
		case strings.HasSuffix(page_id, "index.html"):
			page_url = page_id[0 : len(page_id)-len("index.html")]
		case strings.HasSuffix(page_id, "index.md"):
			page_url = page_id[0 : len(page_id)-len("index.md")]
		default:
			page_url = page_id[0 : len(page_id)-len(filepath.Ext(page_id))]
			if page["title"] == nil && !strings.HasSuffix(page_url, "/") {
				page["title"] = strings.Title(filepath.Base(page_url))
			}
		}
		if strings.HasPrefix(page_url, "/") {
			page["url"] = basePath + page_url[1:]
		} else {
			page["url"] = basePath + page_url
		}

		if page_id != "index.html" && page_id != "index.md" {
			navigation = append(navigation, page_id)
		}
	}
	if site["navigation"] == nil {
		db["navigation"] = navigation
	} else {
		db["navigation"] = AsStrings(site["navigation"])
	}

	dictionary, err := LoadPosts(root, cnf_posts["exclude"].(string))
	if err != nil {
		return
	}
	posts := make(map[string]interface{})
	posts["dictionary"] = dictionary
	db["posts"] = posts

	// for tags, and catalog
	// 配置Tag和catalog
	tags := make(map[string]*Tag)
	catalogs := make(map[string]*Catalog)
	chronological := make([]string, 0)
	collated := make(CollatedYears, 0)

	_collated := make(map[string]*CollatedYear)

	for id, post := range dictionary {
		chronological = append(chronological, id)

		if post.Layout() == "" {
			post["layout"] = post_layout_default
		}
		if post.Permalink() == "" {
			post["permalink"] = post_permalink_default
		}

		for _, _tag := range post.Tags() {
			tag := tags[_tag]
			if tag == nil {
				tag = &Tag{0, _tag, make([]string, 0), "/tags/#" + EncodePathInfo(_tag) + "-ref"}
				tags[_tag] = tag
			}
			tag.Count += 1
			tag.Posts = append(tag.Posts, id)
		}

		for _, _catalog := range post.Categories() {
			catalog := catalogs[_catalog]
			if catalog == nil {
				catalog = &Catalog{0, _catalog, make([]string, 0), "/categories/#" + EncodePathInfo(_catalog) + "-ref"}
				catalogs[_catalog] = catalog
			}
			catalog.Count += 1
			catalog.Posts = append(catalog.Posts, id)
		}

		_year, _month, _ := post["_date"].(time.Time).Date()
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

		CreatePostURL(db, basePath, post)
	}

	// 还得按时间分类
	for _, _yearc := range _collated {
		monthArray := make(CollatedMonths, 0)
		for _, _monthc := range _yearc.months {
			_monthc.Posts = SortPosts(dictionary, _monthc.Posts)
			monthArray = append(monthArray, _monthc)
		}
		sort.Sort(monthArray)
		_yearc.months = nil
		_yearc.Months = monthArray
		collated = append(collated, _yearc)
	}
	sort.Sort(collated)

	for _, catalog := range catalogs {
		catalog.Posts = SortPosts(dictionary, catalog.Posts)
	}
	for _, tag := range tags {
		tag.Posts = SortPosts(dictionary, tag.Posts)
	}

	posts["tags"] = tags
	posts["categories"] = catalogs
	posts["chronological"] = SortPosts(dictionary, chronological)
	posts["collated"] = collated

	return // ~_~ 哦也,搞定收工
}

// 载入所有Page
func LoadPages(root string, exclude string) (pages map[string]Mapper, err error) {
	pages = make(map[string]Mapper)
	err = nil
	var _exclude *regexp.Regexp
	if exclude != "" {
		_exclude, err = regexp.Compile(exclude)
		if err != nil {
			err = errors.New("BAD pages exclude regexp : " + exclude + "\t" + err.Error())
			return
		}
	}
	err = filepath.Walk(root+"pages/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		if _exclude != nil && _exclude.Match([]byte(path[len(root+"pages/"):])) {
			return nil
		}
		page, err := LoadPage(root, path)
		if err != nil {
			return err
		}
		pages[page.Id()] = page
		return nil
	})
	return
}

// 载入特定的一个Page文件
func LoadPage(root string, path string) (ctx Mapper, err error) {
	ctx, err = ReadMuPage(path)
	if err != nil {
		return
	}
	ctx["id"] = path[len(root+"pages/"):]
	return
}

// 载入所有的Post
func LoadPosts(root string, exclude string) (posts map[string]Mapper, err error) {
	posts = make(map[string]Mapper)
	err = nil
	var _exclude *regexp.Regexp
	if exclude != "" {
		_exclude, err = regexp.Compile(exclude)
		if err != nil {
			err = errors.New("BAD pages exclude regexp : " + exclude + "\t" + err.Error())
			return
		}
	}
	err = filepath.Walk(root+"posts/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		if _exclude != nil && _exclude.Match([]byte(path[len(root+"posts/"):])) {
			return nil
		}
		post, err := LoadPost(root, path)
		if err != nil {
			return err
		}
		posts[post.Id()] = post
		return nil
	})
	return
}

// 载入特定的Post
func LoadPost(root string, path string) (ctx Mapper, err error) {
	ctx, err = ReadMuPage(path)
	if err != nil {
		return
	}
	if ctx["date"] == nil {
		err = errors.New("Miss date! >> " + path)
		return
	}
	if ctx["title"] == "" {
		err = errors.New("Miss title! >> " + path)
		return
	}
	var date time.Time
	date, err = time.Parse("2006-01-02", ctx["date"].(string))
	if err != nil {
		date2, err2 := time.Parse("2006-01-02 15:04:05", ctx["date"].(string))
		if err2 != nil {
			err = errors.New("BAD date >>" + path + " " + err.Error() + " " + err2.Error())
			return
		}
		date = date2
		err = nil
	}

	ctx["id"] = path[len(root):]
	ctx["_date"] = date

	ctx["categories"] = ctx.Categories()
	if len(ctx.Categories()) == 0 {
		ctx["categories"] = []string{"default"} // Set default catalog
	}
	ctx["tags"] = ctx.Tags()

	return
}

// 读取包含元数据的文件,返回ctx(包含文本)
func ReadMuPage(path string) (ctx map[string]interface{}, err error) {
	//log.Println("Read", path)
	err = nil
	ctx = nil

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	br := bufio.NewReader(f)
	line, err := br.ReadString('\n')
	if err != nil {
		return
	}
	if !strings.HasPrefix(line, "---") {
		err = errors.New("Not Start with ---   : " + path)
		return
	}

	buf := bytes.NewBuffer(nil)

	for {
		line, err = br.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return
			}
		}
		if strings.HasPrefix(line, "---") {
			break
		}
		buf.WriteString(line)
	}
	buf = bytes.NewBuffer(buf.Bytes())

	ctx, err = ReadYmlReader(buf)
	if err != nil {
		err = errors.New(path + " --> " + err.Error())
		return
	}

	d, err := ioutil.ReadAll(br)
	if err != nil {
		err = errors.New(path + " --> " + err.Error())
		return
	}
	ctx["_content"] = &DocContent{string(d), "", nil}
	return
}

type Tag struct {
	Count int      `json:"count"`
	Name  string   `json:"name"`
	Posts []string `json:"posts"`
	Url   string   `json:"url"`
}

type Catalog struct {
	Count int      `json:"count"`
	Name  string   `json:"name"`
	Posts []string `json:"posts"`
	Url   string   `json:"url"`
}

func AsStrings(v interface{}) (strs []string) {
	strs = make([]string, 0)
	if v == nil {
		return
	}
	switch v.(type) {
	case string:
		strs = []string{v.(string)}
	case []interface{}:
		for _, v2 := range v.([]interface{}) {
			strs = append(strs, fmt.Sprintf("%v", v2))
		}
	case []string:
		strs = v.([]string)
	default:
		log.Println(">>", v)
	}
	//log.Println("##", strs)
	return
}

// 转为URL友好的路径
func EncodePathInfo(pathinfo string) string {
	pathinfo = strings.Replace(pathinfo, " ", "-", -1)
	pathinfo = strings.Replace(pathinfo, ":", "-", -1)
	return URL.QueryEscape(pathinfo)
}

// 解码URL编码的路径信息
func DecodePathInfo(pathinfo string) string {
	pathinfo2, err := URL.QueryUnescape(pathinfo)
	if err != nil {
		log.Println("DecodePathInfo Fail", err)
		return pathinfo
	}
	return pathinfo2
}

// 创建permalink的配置生产路径(不限于Post)
func CreatePostURL(db map[string]interface{}, basePath string, post map[string]interface{}) {
	var url string
	switch post["permalink"].(type) {
	case int64:
		url = strconv.FormatInt(post["permalink"].(int64), 10)
	default:
		url = post["permalink"].(string)
	}
	if strings.Contains(url, ":") {
		year, month, day := post["_date"].(time.Time).Date()
		url = strings.Replace(url, ":year", fmt.Sprintf("%v", year), -1)
		url = strings.Replace(url, ":month", fmt.Sprintf("%02d", month), -1)
		url = strings.Replace(url, ":day", fmt.Sprintf("%02d", day), -1)
		url = strings.Replace(url, ":title", EncodePathInfo(fmt.Sprintf("%v", post["title"])), -1)
		filename := filepath.Base(post["id"].(string))
		ext := filepath.Ext(filename)
		url = strings.Replace(url, ":filename", EncodePathInfo(filename[0:len(filename)-len(ext)]), -1)
		if len(post["categories"].([]string)) > 0 {
			url = strings.Replace(url, ":categories", post["categories"].([]string)[0], -1)
		} else {
			url = strings.Replace(url, ":categories", "", -1)
		}

		url = strings.Replace(url, ":i_month", fmt.Sprintf("%d", month), -1)
		url = strings.Replace(url, ":i_day", fmt.Sprintf("%d", day), -1)
	}

	if strings.HasPrefix(url, "/") {
		post["url"] = basePath + url[1:]
	} else {
		post["url"] = basePath + url
	}
}

type Posts []Mapper

func (p Posts) Len() int {
	return len(p)
}

func (p Posts) Less(i, j int) bool {
	p1_time := p[i]["_date"].(time.Time)
	p2_time := p[j]["_date"].(time.Time)
	if p1_time.Unix() != p2_time.Unix() {
		return p1_time.After(p2_time)
	}

	return p[i].Id() > p[j].Id()
}

func (p Posts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func SortPosts(dict map[string]Mapper, post_ids []string) []string {
	posts := make(Posts, 0)
	for _, post_id := range post_ids {
		posts = append(posts, dict[post_id])
	}
	sort.Sort(posts)
	post_ids = post_ids[0:0]
	for _, post := range posts {
		post_ids = append(post_ids, post.Id())
	}
	return post_ids
}

func LoadLayouts(root string, theme string) map[string]Mapper {
	layouts := make(map[string]Mapper)
	var layout Mapper
	log.Println(">>>", root+"themes/"+theme+"/layouts/")
	filepath.Walk(root+"themes/"+theme+"/layouts/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := filepath.Base(path)
		if strings.HasPrefix(filename, ".") {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		buf := make([]byte, 3)
		_, err = f.Read(buf)
		if err != nil {
			return err
		}
		f.Seek(0, os.SEEK_SET)
		if string(buf) == "---" {
			layout, err = ReadMuPage(path)
			if err != nil {
				return err
			}
			tpl, err := mustache.Parse(bytes.NewBufferString(layout["_content"].(*DocContent).Source))
			if err != nil {
				log.Println("Bad Layout", path, err)
				return err
			}
			layout["_content"].(*DocContent).TPL = tpl
			layout["_content"].(*DocContent).Source = ""
		} else {
			layout = make(map[string]interface{})
			tpl, err := mustache.Parse(f)
			if err != nil {
				return err
			}
			layout["_content"] = &DocContent{"", "", tpl}
		}
		layoutName := filename[0 : len(filename)-len(filepath.Ext(filename))]
		layouts[layoutName] = layout
		log.Println("Load Layout : " + layoutName)
		return nil
	})
	return layouts
}
