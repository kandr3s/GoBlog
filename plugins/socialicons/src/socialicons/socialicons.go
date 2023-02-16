package socialicons

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"go.goblog.app/app/pkgs/bufferpool"
	"go.goblog.app/app/pkgs/htmlbuilder"
	"go.goblog.app/app/pkgs/plugintypes"
)

func GetPlugin() (plugintypes.SetConfig, plugintypes.UI2) {
	p := &plugin{}
	return p, p
}

type plugin struct {
	config map[string]any
}

func (p *plugin) SetConfig(config map[string]any) {
	p.config = config
}

func (p *plugin) RenderWithDocument(rc plugintypes.RenderContext, doc *goquery.Document) {
	blog := rc.GetBlog()
	if blog == "" {
		fmt.Println("social icons plugin: blog is empty!")
		return
	}
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)
	hb := htmlbuilder.NewHtmlBuilder(buf)
	// Social Links
	hb.WriteElementOpen("p", "class", "social-icons", "style", "margin: 0")
	hb.WriteElementOpen("span", "class", "social-link")
	hb.WriteElementOpen("a", "href", "/feeds")
	hb.WriteElementOpen("img", "src", "/assets/icons/rss.svg", "class", "rss-feed", "title", "Subscribe", "width", "18px", "height", "18px")
	hb.WriteElementClose("a")
	hb.WriteElementClose("span")
	hb.WriteElementOpen("span", "class", "social-link")
	hb.WriteElementOpen("a", "target", "_blank", "href", "https://matrix.to/#/@kandr3s:matrix.org")
	hb.WriteElementOpen("img", "src", "/assets/icons/matrix.svg", "title", "Chat on Matrix", "width", "18px", "height", "18px", "style", "background:white")
	hb.WriteElementClose("a")
	hb.WriteElementClose("span")
	hb.WriteElementOpen("span", "class", "social-link")
	hb.WriteElementOpen("a", "target", "_blank", "href", "https://micro.blog/kandr3s")
	hb.WriteElementOpen("img", "src", "/assets/icons/microblog.svg", "class", "microblog", "title", "Micro.blog", "width", "18px", "height", "18px")
	hb.WriteElementClose("a")
	hb.WriteElementClose("span")
	hb.WriteElementOpen("span", "class", "social-link")
	hb.WriteElementOpen("a", "href", "/@kandr3s")
	hb.WriteElementOpen("img", "src", "/assets/icons/mastodon.svg", "class", "mastodon", "title", "On the Fediverse", "width", "18px", "height", "18px")
	hb.WriteElementClose("a")
	hb.WriteElementClose("span")
	hb.WriteElementOpen("span", "class", "social-link")
	hb.WriteElementOpen("a", "target", "_blank", "href", "https://letterboxd.com/kandr3s")
	hb.WriteElementOpen("img", "src", "/assets/icons/letterboxd.svg", "class", "letterboxd", "title", "Letterboxd", "width", "18px", "height", "18px")
	hb.WriteElementClose("a")
	hb.WriteElementClose("span")
	hb.WriteElementOpen("span", "class", "social-link")
	hb.WriteElementOpen("a", "target", "_blank", "href", "https://rateyourmusic.com/~kandr3s")
	hb.WriteElementOpen("img", "src", "/assets/icons/sonemic.svg", "class", "sonemic", "title", "RYM/Sonemic", "width", "18px", "height", "18px")
	hb.WriteElementClose("a")
	hb.WriteElementClose("span")
	hb.WriteElementOpen("span", "class", "social-link")
	hb.WriteElementOpen("a", "target", "_blank", "href", "https://github.com/kandr3s")
	hb.WriteElementOpen("img", "src", "/assets/icons/github-alt.svg", "class", "github", "title", "GitHub", "width", "18px", "height", "18px")
	hb.WriteElementClose("a")
	hb.WriteElementClose("span")
	//hb.WriteElementOpen("span", "class", "social-link")
	//hb.WriteElementOpen("a", "target", "_blank", "href", "https://twitter.com/kandr3s")
	//hb.WriteElementOpen("img", "src", "/assets/icons/twitter.svg", "class", "twitter", "title", "Twitter", "width", "18px", "height", "18px")
	//hb.WriteElementClose("a")
	//hb.WriteElementClose("span")
	hb.WriteElementClose("p")
	doc.Find("header").AppendHtml(buf.String())
}
