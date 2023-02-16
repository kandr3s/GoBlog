package letterboxd

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

func GetPlugin() (plugintypes.SetConfig, plugintypes.SetApp, plugintypes.UI) {
	p := &plugin{}
	return p, p, p
}

func (p *plugin) SetApp(app plugintypes.App) {
	p.app = app
}

func (p *plugin) SetConfig(config map[string]any) {
	p.parameterName = "syndication" // default
}

func (p *plugin) Render(rc plugintypes.RenderContext, rendered io.Reader, modified io.Writer) {
	def := func() {
		_, _ = io.Copy(modified, rendered)
	}
	post, err := p.app.GetPost(rc.GetPath())
	if err != nil || post == nil {
		def()
		return
	}
	letterboxdLink, ok := post.GetParameters()[p.parameterName]
	if !ok || len(letterboxdLink) == 0 {
		def()
		return
	}
	doc, err := goquery.NewDocumentFromReader(rendered)
	if err != nil {
		fmt.Println("letterboxd plugin: " + err.Error())
		def()
		return
	}
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)
	hb := htmlbuilder.NewHtmlBuilder(buf)
	for _, link := range letterboxdLink {
		boxd := "boxd"
		if strings.Contains(link, boxd) {
			doc.Find("main.h-entry article div.e-content p img").AddClass("u-photo")
		} else {
			break
		}
	}
	_ = goquery.Render(modified, doc.Selection)
}
