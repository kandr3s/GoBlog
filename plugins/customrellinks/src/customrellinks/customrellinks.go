package customrellinks

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
	if customRelLinksAny, ok := p.config[blog]; ok {
		if customRelLinks, ok := customRelLinksAny.([]any); ok {
			buf := bufferpool.Get()
			defer bufferpool.Put(buf)
			hb := htmlbuilder.NewHtmlBuilder(buf)
			for _, customElementAny := range customRelLinks {
				if element, ok := customElementAny.(map[string]any); ok {
					name := unwrapToString(element["name"])
					link := unwrapToString(element["link"])
					// Microsub
					if name == "microsub" {
						hb.WriteElementOpen("link", "rel", "microsub", "href", link)
					}
					// Manifest
					if name == "manifest" {
						hb.WriteElementOpen("link", "rel", "manifest", "href", "/manifest.json")
					}
				}
			}
			doc.Find("head").AppendHtml(buf.String())
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
