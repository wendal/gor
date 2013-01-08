package gor

import (
	"errors"
	"fmt"
	"github.com/wendal/mustache"
	"log"
	"os"
	"path/filepath"
)

var (
	WidgetBuilders = make(map[string]WidgetBuilder)
)

const (
	Analytics_google = `
<script>
    var _gaq=[['_setAccount','%s'],['_trackPageview']];
    (function(d,t){var g=d.createElement(t),s=d.getElementsByTagName(t)[0];
    g.src=('https:'==location.protocol?'//ssl':'//www')+'.google-analytics.com/ga.js';
    s.parentNode.insertBefore(g,s)}(document,'script'));
</script>`
	Comments_disqus = `
<div id="disqus_thread"></div>
<script>
    var disqus_developer = 1;
    var disqus_shortname = '%s'; // required: replace example with your forum shortname
    /* * * DON'T EDIT BELOW THIS LINE * * */
    (function() {
        var dsq = document.createElement('script'); dsq.type = 'text/javascript'; dsq.async = true;
        dsq.src = 'http://' + disqus_shortname + '.disqus.com/embed.js';
        (document.getElementsByTagName('head')[0] || document.getElementsByTagName('body')[0]).appendChild(dsq);
    })();
</script>
<noscript>Please enable JavaScript to view the <a href="http://disqus.com/?ref_noscript">comments powered by Disqus.</a></noscript>
<a href="http://disqus.com" class="dsq-brlink">blog comments powered by <span class="logo-disqus">Disqus</span></a>
`
	tpl_google_prettify = `
<script src="http://cdnjs.cloudflare.com/ajax/libs/prettify/188.0.0/prettify.js"></script>
<script>
  var pres = document.getElementsByTagName("pre");
  for (var i=0; i < pres.length; ++i) {
    pres[i].className = "prettyprint %s";
  }
  prettyPrint();
</script>
`
)

type WidgetBuilder func(Mapper, mustache.Context) (Widget, error)

type Widget interface {
	Prepare(mapper Mapper, ctx mustache.Context) Mapper
}

func init() {
	WidgetBuilders["analytics"] = BuildAnalyticsWidget
	WidgetBuilders["comments"] = BuildCommentsWidget
	WidgetBuilders["google_prettify"] = BuildGoogle_prettify
}

func LoadWidgets(topCtx mustache.Context) ([]Widget, error) {
	widgets := make([]Widget, 0)
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
		if builderFunc == nil {
			log.Println("NO WidgetBuilder >>", cnf_path)
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
	return widgets, err
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
	if cnf.Layout() != "google" {
		return nil, errors.New("AnalyticsWidget Only for Goolge yet")
	}
	google := cnf[cnf.Layout()].(map[string]interface{})
	tracking_id := google["tracking_id"]
	if tracking_id == nil {
		return nil, errors.New("AnalyticsWidget Of Google need tracking_id")
	}
	self := make(AnalyticsWidget)
	self["analytics"] = fmt.Sprintf(Analytics_google, tracking_id)
	return self, nil
}

//--------------------------------------------------------------------------------

type CommentsWidget Mapper

func (self CommentsWidget) Prepare(mapper Mapper, topCtx mustache.Context) Mapper {
	if mapper["comments"] != nil && !mapper["comments"].(bool) {
		return nil
	}
	return Mapper(self)
}

func BuildCommentsWidget(cnf Mapper, topCtx mustache.Context) (Widget, error) {
	if cnf.Layout() != "disqus" {
		return nil, errors.New("CommentsWidget Only for disqus yet")
	}
	disqus := cnf[cnf.Layout()].(map[string]interface{})
	short_name := disqus["short_name"]
	if short_name == nil {
		return nil, errors.New("CommentsWidget Of disqus need short_name")
	}
	self := make(CommentsWidget)
	self["comments"] = fmt.Sprintf(Comments_disqus, short_name)
	return self, nil
}

//-----------------------------------------------
type google_prettify Mapper

func (self google_prettify) Prepare(mapper Mapper, topCtx mustache.Context) Mapper {
	if mapper["google_prettify"] != nil && !mapper["google_prettify"].(bool) {
		return nil
	}
	return Mapper(self)
}

func BuildGoogle_prettify(cnf Mapper, topCtx mustache.Context) (Widget, error) {
	if cnf["linenums"].(bool) {
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
			mappers = append(mappers, mr)
		}
	}
	return mustache.MakeContexts(mappers...)
}
