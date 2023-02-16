package indiewebicons

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
	// IndieWeb Icons
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
