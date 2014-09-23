package gor

import (
	"bytes"
	"fmt"
	"github.com/wendal/errors"
	"github.com/wendal/mustache"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type WidgetBuilder func(Mapper, mustache.Context) (Widget, error)

type Widget interface {
	Prepare(mapper Mapper, ctx mustache.Context) Mapper
}

func init() {
}

// 遍历目录,加载挂件
func LoadWidgets(topCtx mustache.Context) ([]Widget, string, error) {
	widgets := make([]Widget, 0)
	assets := ""

	err := filepath.Walk("widgets", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		cnf_path := path + "/config.yml"
		fst, err := os.Stat(cnf_path)
		if err != nil || fst.IsDir() {
			return nil //ignore
		}
		cnf, err := ReadYml(cnf_path)
		if err != nil {
			return errors.New(cnf_path + ":" + err.Error())
		}
		if cnf["layout"] != nil {
			widget_enable, ok := cnf["layout"].(bool)
			if ok && !widget_enable {
				log.Println("Disable >", cnf_path)
			}
		}

		log.Println("path>", path)

		widget, err := BuildWidget(path, cnf, topCtx)

		//log.Println("BuildWidget widget>", widget)
		//log.Println("BuildWidget err>", err)

		if err != nil {
			_widget, _assets, _err := BuildCustomWidget(info.Name(), path, cnf)
			if _err != nil {
				log.Println("NO WidgetBuilder >>", cnf_path, _err)
			}
			if _widget != nil {
				widgets = append(widgets, _widget)
				if _assets != nil {
					for _, asset := range _assets {
						assets += asset + "\n"
					}
				}
			}
			return nil
		}
		widgets = append(widgets, widget)
		log.Println("Load widget from ", cnf_path)
		return nil
	})

	//log.Println("LoadWidgets widgte > ", widgets)
	//log.Println("LoadWidgets assets > ", assets)
	return widgets, assets, err
}

//批量解析处理函数，解析所有包含 config.yml 文件的 widget 子目录
type MyWidget Mapper

func (self MyWidget) Prepare(mapper Mapper, topCtx mustache.Context) Mapper {
	if mapper["analytics"] != nil && !mapper["analytics"].(bool) {
		return nil
	}
	return Mapper(self)
}

func BuildWidget(path string, cnf Mapper, topCtx mustache.Context) (Widget, error) {
	layout := cnf.Layout()
	tracking := cnf[layout].(map[string]interface{})

	log.Println("BuildWidget >>", path, layout, tracking)

	if tracking == nil {
		return nil, errors.New(path + "Widget need config")
	}

	var doc bytes.Buffer
	if tmpl, err := template.ParseFiles(path + "/layouts/" + layout + ".tmpl"); err != nil {
		return nil, err
	} else {
		if err := tmpl.Execute(&doc, tracking); err != nil {
			return nil, err
		}
	}

	self := make(MyWidget)
	self[path] = doc.String()
	return self, nil
}

func PrapareWidgets(widgets []Widget, mapper Mapper, topCtx mustache.Context) mustache.Context {
	mappers := make([]interface{}, 0)
	for _, widget := range widgets {
		mr := widget.Prepare(mapper, topCtx)
		if mr != nil {
			for k, v := range mr {
				mapper[k] = v
			}
			mappers = append(mappers, mr)
		}
	}
	return mustache.MakeContexts(mappers...)
}

type CustomWidget struct {
	name   string
	layout *DocContent
	mapper Mapper
}

func (c *CustomWidget) Prepare(mapper Mapper, ctx mustache.Context) Mapper {
	return Mapper(map[string]interface{}{c.name: c.layout.Source})
}

func BuildCustomWidget(name string, dir string, cnf Mapper) (Widget, []string, error) {
	layoutName, ok := cnf["layout"]
	if !ok || layoutName == "" {
		log.Println("Skip Widget : " + dir)
		return nil, nil, nil
	}

	layoutFilePath := dir + "/layouts/" + layoutName.(string) + ".tmpl"
	f, err := os.Open(layoutFilePath)
	if err != nil {
		return nil, nil, errors.New("Fail to load Widget Layout" + dir + "\n" + err.Error())
	}
	defer f.Close()
	cont, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, nil, errors.New("Fail to load Widget Layout" + dir + "\n" + err.Error())
	}

	assets := []string{}
	for _, js := range cnf.GetStrings("javascripts") {
		path := "/assets/" + dir + "/javascripts/" + js
		assets = append(assets, fmt.Sprintf("<script type=\"text/javascript\" src=\"%s\"></script>", path))
	}
	for _, css := range cnf.GetStrings("stylesheets") {
		path2 := "/assets/" + dir + "/stylesheets/" + css
		assets = append(assets, fmt.Sprintf("<link href=\"%s\" type=\"text/css\" rel=\"stylesheet\" media=\"all\">", path2))
	}

	return &CustomWidget{name, &DocContent{string(cont), string(cont), nil}, cnf}, assets, nil

}
