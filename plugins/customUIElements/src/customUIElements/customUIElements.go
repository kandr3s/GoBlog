package customUIElements

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
		fmt.Println("Custom UI Elements plugin: blog is empty!")
		return
	}
	if customUIElementsAny, ok := p.config[blog]; ok {
		if customUIElements, ok := customUIElementsAny.([]any); ok {
			buf := bufferpool.Get()
			defer bufferpool.Put(buf)
			hb := htmlbuilder.NewHtmlBuilder(buf)
			for _, customElements := range customUIElements {
				if element, ok := customElements.(map[string]any); ok {
					name := element["name"]
					if name == "microsub" {
						buf.Reset()
						link := element["link"]
						hb.WriteElementOpen("link", "rel", "microsub", "href", link)
						doc.Find("head").AppendHtml(buf.String())
					}
					if name == "manifest" {
						buf.Reset()
						hb.WriteElementOpen("link", "rel", "manifest", "href", "/manifest.json")
						doc.Find("head").AppendHtml(buf.String())
					}
					if name == "socialicons" {
						buf.Reset()
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
						hb.WriteElementClose("p")
						doc.Find(".show-menu.menu-icon").BeforeHtml(buf.String())
					}
					if name == "indiewebicons" {
						buf.Reset()
						hb.WriteElementOpen("p")
						hb.WriteElementOpen("a", "target", "_blank", "href", "https://indieweb.org/Webmention")
						hb.WriteElementOpen("img", "src", "/images/webmention_button.png", "title", "This website supports Webmentions")
						hb.WriteElementClose("a")
						hb.WriteElementOpen("a", "target", "_blank", "href", "https://microformats.org")
						hb.WriteElementOpen("img", "src", "/images/microformats_button.png", "title", "This website supports Webmentions")
						hb.WriteElementClose("a")
						hb.WriteElementOpen("a", "target", "_blank", "href", "https://indieweb.org/")
						hb.WriteElementOpen("img", "src", "/images/indieweb_button.png", "title", "This website supports Webmentions")
						hb.WriteElementClose("a")
						hb.WriteElementClose("p")
						doc.Find("footer").AppendHtml(buf.String())
					}
				}
			}

		}
	}
}

func unwrapToString(o any) (string, bool) {
	if o == nil {
		return "", false
	}
	s, ok := o.(string)
	return s, ok && s != ""
}
