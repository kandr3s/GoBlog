package syndication

import (
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.goblog.app/app/pkgs/bufferpool"
	"go.goblog.app/app/pkgs/htmlbuilder"
	"go.goblog.app/app/pkgs/plugintypes"
)

type plugin struct {
	app           plugintypes.App
	parameterName string
}

func GetPlugin() (plugintypes.SetConfig, plugintypes.SetApp, plugintypes.UI2) {
	p := &plugin{}
	return p, p, p
}

func (p *plugin) SetApp(app plugintypes.App) {
	p.app = app
}

func (p *plugin) SetConfig(config map[string]any) {
	p.parameterName = "syndication" // default
	if configParameterAny, ok := config["parameter"]; ok {
		if configParameter, ok := configParameterAny.(string); ok {
			p.parameterName = configParameter // override default from config
		}
	}
}

func (p *plugin) RenderWithDocument(rc plugintypes.RenderContext, doc *goquery.Document) {
	post, err := p.app.GetPost(rc.GetPath())
	if err != nil || post == nil {
		return
	}
	syndicationLinks, ok := post.GetParameters()[p.parameterName]
	if !ok || len(syndicationLinks) == 0 {
		return
	}
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)
	hb := htmlbuilder.NewHtmlBuilder(buf)
	for _, link := range syndicationLinks {
		rym := "rateyourmusic.com"
		masto := "todon.org"
		twitter := "twitter.com"
		boxd := "boxd"
		micro := "micro.blog"
		if strings.Contains(link, rym) {
			hb.WriteElementOpen("span", "class", "syndication")
			hb.WriteElementOpen("a", "href", link, "class", "sonemic", "rel", "syndication", "title", "This post on RYM/Sonemic")
			hb.WriteElementOpen("img", "src", "/assets/icons/sonemic.svg", "style", "width: 1rem")
			hb.WriteElementClose("a")
			hb.WriteElementClose("span")
		}
		if strings.Contains(link, masto) {
			hb.WriteElementOpen("span", "class", "syndication")
			hb.WriteElementOpen("a", "href", link, "class", "mastodon", "rel", "syndication", "title", "This post on the Fediverse")
			hb.WriteElementOpen("img", "src", "/assets/icons/mastodon.svg", "style", "width: 1rem")
			hb.WriteElementClose("a")
			hb.WriteElementClose("span")
		}
		if strings.Contains(link, twitter) {
			hb.WriteElementOpen("span", "class", "syndication")
			hb.WriteElementOpen("a", "href", link, "class", "twitter", "rel", "syndication", "title", "This post on Twitter")
			hb.WriteElementOpen("img", "src", "/assets/icons/twitter.svg", "style", "width: 1rem")
			hb.WriteElementClose("a")
			hb.WriteElementClose("span")
		}
		if strings.Contains(link, boxd) {
			hb.WriteElementOpen("span", "class", "syndication")
			hb.WriteElementOpen("a", "href", link, "class", "letterboxd", "rel", "syndication", "title", "This post on Letterboxd")
			hb.WriteElementOpen("img", "src", "/assets/icons/letterboxd.svg", "style", "width: 1rem")
			hb.WriteElementClose("a")
			hb.WriteElementClose("span")
		}
		if strings.Contains(link, micro) {
			hb.WriteElementOpen("span", "class", "syndication")
			hb.WriteElementOpen("a", "href", link, "class", "microblog", "rel", "syndication", "title", "This post on Micro.blog")
			hb.WriteElementOpen("img", "src", "/assets/icons/microblog.svg", "style", "width: 1rem")
			hb.WriteElementClose("a")
			hb.WriteElementClose("span")
		} else {
			hb.WriteElementOpen("data", "value", link, "class", "u-syndication hide")
			hb.WriteElementClose("data")
		}
	}
	doc.Find("main.h-entry article").AppendHtml(buf.String())
}
