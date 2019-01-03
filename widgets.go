package gor

import (
	"fmt"
	"github.com/wendal/errors"
	"github.com/wendal/mustache"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	// 默认的挂件
	WidgetBuilders = make(map[string]WidgetBuilder)
)

const (
	Analytics_google = `
<script type="text/javascript">

  var _gaq = _gaq || [];
  var pluginUrl = '//www.google-analytics.com/plugins/ga/inpage_linkid.js';
  _gaq.push(['_require', 'inpage_linkid', pluginUrl]);
  _gaq.push(['_setAccount', '%s']);
  _gaq.push(['_trackPageview']);

  (function() {
    var ga = document.createElement('script'); ga.type = 'text/javascript'; ga.async = true;
    ga.src = ('https:' == document.location.protocol ? 'https://ssl' : 'http://www') + '.google-analytics.com/ga.js';
    var s = document.getElementsByTagName('script')[0]; s.parentNode.insertBefore(ga, s);
  })();

</script>`
	Comments_disqus = `
<div id="disqus_thread"></div>
<script>
    var disqus_developer = 1;
    var disqus_shortname = '%s'; // required: replace example with your forum shortname
    /* * * DON'T EDIT BELOW THIS LINE * * */
    (function() {
        var dsq = document.createElement('script'); dsq.type = 'text/javascript'; dsq.async = true;
        dsq.src = '//' + disqus_shortname + '.disqus.com/embed.js';
        (document.getElementsByTagName('head')[0] || document.getElementsByTagName('body')[0]).appendChild(dsq);
    })();
</script>
<noscript>Please enable JavaScript to view the <a href="http://disqus.com/?ref_noscript">comments powered by Disqus.</a></noscript>
<a href="http://disqus.com" class="dsq-brlink">blog comments powered by <span class="logo-disqus">Disqus</span></a>
`
	tpl_google_prettify = `
<script src="//cdnjs.cloudflare.com/ajax/libs/prettify/r298/prettify.min.js"></script>
<script>
  var pres = document.getElementsByTagName("pre");
  for (var i=0; i < pres.length; ++i) {
    pres[i].className = "prettyprint %s";
  }
  prettyPrint();
</script>
`
	Comments_duoshuo = `
	<!-- Duoshuo Comment BEGIN -->
	<div class="ds-thread"></div>
	<script type="text/javascript">
	var duoshuoQuery = {short_name:"%s"};//require,replace your short_name
	(function() {
					var ds = document.createElement('script');
					ds.type = 'text/javascript';ds.async = true;
					ds.src = '//static.duoshuo.com/embed.js';
					ds.charset = 'UTF-8';
					(document.getElementsByTagName('head')[0] 
					|| document.getElementsByTagName('body')[0]).appendChild(ds);
	})();
	</script>
	<!-- Duoshuo Comment END -->	
	`
	tpl_cnzz = `<script src="http://s25.cnzz.com/stat.php?id=%d&web_id=%d" language="JavaScript"></script>`

	tpl_uyan = `
<!-- UY BEGIN -->
<div id="uyan_frame"></div>
<script type="text/javascript" src="http://v2.uyan.cc/code/uyan.js?uid=%d"></script>
<!-- UY END -->
	`
)

type WidgetBuilder func(Mapper, mustache.Context) (Widget, error)

type Widget interface {
	Prepare(mapper Mapper, ctx mustache.Context) Mapper
}

func init() {
	WidgetBuilders["analytics"] = BuildAnalyticsWidget       //访问统计
	WidgetBuilders["comments"] = BuildCommentsWidget         //社会化评论
	WidgetBuilders["google_prettify"] = BuildGoogle_prettify // 代码高亮
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
		builderFunc := WidgetBuilders[info.Name()]
		if builderFunc == nil { // 看看是否符合自定义挂件的格式
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
		widget, err := builderFunc(cnf, topCtx)
		if err != nil {
			return err
		}
		widgets = append(widgets, widget)
		log.Println("Load widget from ", cnf_path)
		return nil
	})
	return widgets, assets, err
}

//-------------------------------------------------------------------------------------
type AnalyticsWidget Mapper

func (self AnalyticsWidget) Prepare(mapper Mapper, topCtx mustache.Context) Mapper {
	if mapper["analytics"] != nil && !mapper["analytics"].(bool) {
		return nil
	}
	return Mapper(self)
}

func BuildAnalyticsWidget(cnf Mapper, topCtx mustache.Context) (Widget, error) {
	switch cnf.Layout() {
	case "google": // 鼎鼎大名的免费,但有点拖慢加载速度,原因你懂的
		google := cnf[cnf.Layout()].(map[string]interface{})
		tracking_id := google["tracking_id"]
		if tracking_id == nil {
			return nil, errors.New("AnalyticsWidget Of Google need tracking_id")
		}
		self := make(AnalyticsWidget)
		self["analytics"] = fmt.Sprintf(Analytics_google, tracking_id)
		return self, nil
	case "cnzz": //免费,而且很快,但强制嵌入一个反向链接,靠!
		cnzz := cnf[cnf.Layout()].(map[string]interface{})
		tracking_id := cnzz["tracking_id"]
		if tracking_id == nil {
			return nil, errors.New("AnalyticsWidget Of CNZZ need tracking_id")
		}
		self := make(AnalyticsWidget)
		self["analytics"] = fmt.Sprintf(tpl_cnzz, tracking_id, tracking_id)
		return self, nil
	}
	// 其他的尚不支持, 如果需要,请报个issue吧
	return nil, errors.New("AnalyticsWidget Only for Goolge/CNZZ yet")

}

//--------------------------------------------------------------------------------
// 社会化屏幕
type CommentsWidget Mapper

func (self CommentsWidget) Prepare(mapper Mapper, topCtx mustache.Context) Mapper {
	if mapper["comments"] != nil && !mapper["comments"].(bool) {
		log.Println("Disable comments")
		return nil
	}
	return Mapper(self)
}

func BuildCommentsWidget(cnf Mapper, topCtx mustache.Context) (Widget, error) {
	log.Println("Comments >>", cnf.Layout())
	switch cnf.Layout() {
	case "disqus":
		disqus := cnf[cnf.Layout()].(map[string]interface{})
		short_name := disqus["short_name"]
		if short_name == nil {
			return nil, errors.New("CommentsWidget Of disqus need short_name")
		}
		self := make(CommentsWidget)
		self["comments"] = fmt.Sprintf(Comments_disqus, short_name)
		return self, nil
	case "uyan" :
		uyan := cnf[cnf.Layout()].(map[string]interface{})
		uid := uyan["uid"]
		self := make(CommentsWidget)
		self["comments"] = fmt.Sprintf(tpl_uyan, uid)
		return self, nil
	case "duoshuo":
		duoshuo := cnf[cnf.Layout()].(map[string]interface{})
		short_name := duoshuo["short_name"]
		if short_name == nil {
			return nil, errors.New("CommentsWidget Of duoshuo need short_name")
		}
		self := make(CommentsWidget)
		self["comments"] = fmt.Sprintf(Comments_duoshuo, short_name)
		return self, nil
	}
	// 其他的,想不到还有啥,哈哈,需要其他的就报个issue吧
	return nil, errors.New("CommentsWidget Only for disqus yet")
}

//-----------------------------------------------
// 代码高亮
type google_prettify Mapper

func (self google_prettify) Prepare(mapper Mapper, topCtx mustache.Context) Mapper {
	if mapper["google_prettify"] != nil && !mapper["google_prettify"].(bool) {
		return nil
	}
	return Mapper(self)
}

func BuildGoogle_prettify(cnf Mapper, topCtx mustache.Context) (Widget, error) {
	if enable, ok := cnf["linenums"].(bool); ok && enable { //是否显示行号
		self := make(google_prettify)
		self["google_prettify"] = fmt.Sprintf(tpl_google_prettify, "linenums")
		return self, nil
	}
	self := make(google_prettify)
	self["google_prettify"] = fmt.Sprintf(tpl_google_prettify, "")
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

	layoutFilePath := dir + "/layouts/" + layoutName.(string) + ".html"
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
