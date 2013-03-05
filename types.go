package gor

import (
	"fmt"
	"github.com/wendal/mustache"
	"log"
	"time"
)

// 最重要的封装类之一
// Golang是强静态语言,无法动态添加/删除属性, 而元数据(map[string]interface{})允许包含用户自定义的key
// 所以只能使用Mapper这类封装部分常用Getter
type Mapper map[string]interface{}

func (m Mapper) Get(key string) interface{} {
	return m[key]
}

func (m Mapper) GetString(key string) string {
	val := m[key]
	if val == nil {
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", val)
}

func (m *Mapper) String(key string) string {
	return m.GetString(key)
}

// goyaml2获取string的机制决定了string肯定trim了的. 但依赖这个特性,靠谱不?
/*
func (m *Mapper) StringTrim(key string) string {
	str := m.GetString(key)
	return strings.Trim(str, " \t\n\r")
}
*/

func (m *Mapper) Int64(key string) int64 {
	return m.GetInt(key)
}

func (m *Mapper) Int(key string) int {
	return int(m.GetInt(key))
}

func (m Mapper) GetInt(key string) int64 {
	val := m[key]
	if val == nil {
		return 0
	}
	i, ok := val.(int64)
	if ok {
		return i
	}
	i2, ok := val.(int)
	if ok {
		return int64(i2)
	}
	return 0
}

func (m Mapper) Id() string {
	return m.GetString("id")
}

func (m Mapper) Url() string {
	return m.GetString("url")
}

func (m Mapper) Layout() string {
	return m.GetString("layout")
}

func (m Mapper) Permalink() string {
	return m.GetString("permalink")
}

func (m Mapper) Tags() []string {
	return m.GetStrings("tags")
}

func (m Mapper) Categories() []string {
	return m.GetStrings("categories")
}

// 由于是类型不可预知,所以需要自行封装为[]string
func (m Mapper) GetStrings(key string) (strs []string) {
	v := m[key]
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
	return
}

type DocContent struct {
	Source string             `json:"-"`
	Main   string             `json:"-"`
	TPL    *mustache.Template `json:"-"`
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

type WebSite struct {
	Root     string
	SiteCnf  SiteConfig
	TopCnf   TopConfig
	Posts    map[string]PostBean
	Pages    map[string]PageBean
	ThemeCnf ThemeConfig
	Layouts  map[string]Mapper

	RootURL   string
	BasePath  string
	BaiseURLs map[string]string

	Tags          map[string]*Tag
	Catalogs      map[string]*Catalog
	Chronological []string
	Collated      CollatedYears
}

type SiteConfig struct {
	Title      string
	Tagline    string
	Author     Mapper
	Navigation []string
	//Urls       map[string]interface{} // for user custom
}

type TopConfig struct {
	Theme          string
	Production_url string
	Posts          PostConfig
	Pages          PageConfig
	Paginator      PaginatorConfig
}

type PostConfig struct {
	Permalink     string
	Summary_lines int
	Latest        int
	Layout        string
	Exclude       string
}

type PageConfig struct {
	Permalink string
	Layout    string
	Exclude   string
}

type PaginatorConfig struct {
	Namespace string
	Per_page  int
	Root_page string
	Layout    string
}

type PostBean struct {
	Id         string
	Title      string
	Date       string
	Layout     string
	Permalink  string
	Categories []string
	Tags       []string
	Url        string
	_Date      time.Time
	_Meta      map[string]interface{}
}
type PageBean struct {
	Id         string
	Title      string
	Date       time.Time
	Layout     string
	Permalink  string
	Categories []string
	Tags       []string
	Url        string
	_Meta      map[string]interface{}
}

type ThemeConfig struct{}
