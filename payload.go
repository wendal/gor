package gor

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func BuildPlayload() (payload map[string]interface{}, err error) {
	payload = make(Mapper)
	err = nil

	//-----------------------------------
	cnf, err := ReadConfig(".")
	if err != nil {
		return
	}
	site, err := ReadYml("./site.yml")
	if err != nil {
		return
	}
	site["config"] = cnf
	payload["site"] = site

	// Check site config!
	if cnf["theme"] == nil {
		err = errors.New("Miss theme config!")
		return
	}
	production_url := cnf["production_url"]
	if production_url == nil {
		err = errors.New("Miss production_url")
		return
	}
	rootUrl := production_url.(string)
	if !strings.HasPrefix(rootUrl, "http://") && !strings.HasPrefix(rootUrl, "https://") {
		err = errors.New("production_url must start with https:// or http://")
		return
	}
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

	//---------------------------------
	themeCnf, err := ReadYml("./themes/" + cnf["theme"].(string) + "/theme.yml")
	if err != nil {
		return
	}
	payload["theme"] = themeCnf

	//-------------------------------
	urls := make(map[string]string)
	urls["media"] = basePath + "assets/media"
	urls["theme"] = basePath + "assets/" + cnf["theme"].(string)
	urls["theme_media"] = urls["theme"] + "/media"
	urls["theme_javascripts"] = urls["theme"] + "/javascripts"
	urls["theme_stylesheets"] = urls["theme"] + "/stylesheets"
	urls["base_path"] = basePath

	payload["urls"] = urls

	//---------------------------------
	// set default configs
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
	if cnf_posts.GetInt("lastest") < 5 {
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

	//---------------------------------

	db := make(map[string]interface{})
	payload["db"] = db

	pages, err := LoadPages(cnf_pages.GetString("exclude"))
	if err != nil {
		return
	}
	db["pages"] = pages

	navigation := make([]string, 0)

	for page_id, page := range pages {
		if page.Layout() == "" {
			page["layout"] = page_layout_default
		}
		if page.Permalink() == "" {
			page["permalink"] = page_permalink_default
		}

		//TODO create page URL
		page_url := ""
		switch {
		case strings.HasSuffix(page_id, "index.html"):
			page_url = page_id[0 : len(page_id)-len("index.html")]
		case strings.HasSuffix(page_id, "index.md"):
			page_url = page_id[0 : len(page_id)-len("index.md")]
		default:
			page_url = page_id[0 : len(page_id)-len(filepath.Ext(page_id))]
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
	db["navigation"] = navigation

	dictionary, err := LoadPosts(cnf_posts["exclude"].(string))
	if err != nil {
		return
	}
	posts := make(map[string]interface{})
	posts["dictionary"] = dictionary
	db["posts"] = posts

	// for tags, and catalog
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
				tag = &Tag{0, _tag, make([]string, 0), "/tag#" + _tag + "-ref"}
				tags[_tag] = tag
			}
			tag.Count += 1
			tag.Posts = append(tag.Posts, id)
		}

		for _, _catalog := range post.Categories() {
			catalog := catalogs[_catalog]
			if catalog == nil {
				catalog = &Catalog{0, _catalog, make([]string, 0), "/catalogs#" + _catalog + "-ref"}
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

	for _, _yearc := range _collated {
		monthArray := make(CollatedMonths, 0)
		for _, _monthc := range _yearc.months {
			// TODO Sort Posts
			monthArray = append(monthArray, _monthc)
		}
		sort.Sort(monthArray)
		_yearc.months = nil
		_yearc.Months = monthArray
		collated = append(collated, _yearc)
	}
	sort.Sort(collated)

	posts["tags"] = tags
	posts["catalogs"] = catalogs
	posts["chronological"] = chronological
	posts["collated"] = collated

	return
}

func LoadPages(exclude string) (pages map[string]Mapper, err error) {
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
	err = filepath.Walk("pages/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		if _exclude != nil && _exclude.Match([]byte(path[len("pages/"):])) {
			return nil
		}
		page, err := LoadPage(path)
		if err != nil {
			return err
		}
		pages[page.Id()] = page
		return nil
	})
	return
}

func LoadPage(path string) (ctx Mapper, err error) {
	ctx, err = ReadMuPage(path)
	if err != nil {
		return
	}
	ctx["id"] = path[len("pages/"):]
	return
}

func LoadPosts(exclude string) (posts map[string]Mapper, err error) {
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
	err = filepath.Walk("posts/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		if _exclude != nil && _exclude.Match([]byte(path[len("posts/"):])) {
			return nil
		}
		post, err := LoadPost(path)
		if err != nil {
			return err
		}
		posts[post.Id()] = post
		return nil
	})
	return
}

func LoadPost(path string) (ctx Mapper, err error) {
	ctx, err = ReadMuPage(path)
	if err != nil {
		return
	}
	if ctx["date"] == nil {
		err = errors.New("Miss date! >> " + path + " " + err.Error())
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

	ctx["id"] = path
	ctx["_date"] = date

	ctx["categories"] = ctx.Categories()
	ctx["tags"] = ctx.Tags()

	return
}

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
	ctx["content"] = &DocContent{string(d)}
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

type DocContent struct {
	Source string `json:"-"`
}

type CollatedYear struct {
	Year   string
	Months []*CollatedMonth
	months map[string]*CollatedMonth `json:"-"`
}

type CollatedMonth struct {
	Month  string
	_month time.Month `json:"-"`
	Posts  []string
}

type CollatedYears []*CollatedYear

func (c CollatedYears) Len() int {
	return len(c)
}

func (c CollatedYears) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c CollatedYears) Less(i, j int) bool {
	return c[i].Year > c[j].Year
}

type CollatedMonths []*CollatedMonth

func (c CollatedMonths) Len() int {
	return len(c)
}

func (c CollatedMonths) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c CollatedMonths) Less(i, j int) bool {
	return c[i]._month > c[j]._month
}

func CreatePostURL(db map[string]interface{}, basePath string, post map[string]interface{}) {
	url := post["permalink"].(string)
	if strings.Contains(url, ":") {
		year, month, day := post["_date"].(time.Time).Date()
		url = strings.Replace(url, ":year", fmt.Sprintf("%v", year), -1)
		url = strings.Replace(url, ":month", fmt.Sprintf("%02d", month), -1)
		url = strings.Replace(url, ":day", fmt.Sprintf("%02d", day), -1)
		url = strings.Replace(url, ":title", fmt.Sprintf("%v", post["title"]), -1)
		url = strings.Replace(url, ":filename", filepath.Dir(post["id"].(string)), -1)
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
		post["url"] = basePath + url[1:]
	}
}
