package syndication

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
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

func GetPlugin() (plugintypes.SetConfig, plugintypes.SetApp, plugintypes.UI2, plugintypes.PostCreatedHook, plugintypes.PostUpdatedHook) {
	p := &plugin{}
	return p, p, p, p, p
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

// TODO move customization to UI/Momentos plugin
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
	hb.WriteElementOpen("div", "class", "syndication")
	hb.WriteElementOpen("strong")
	hb.WriteUnescaped("Also on: ")
	hb.WriteElementClose("strong")
	for i, link := range syndicationLinks {
		if i > 0 {
			hb.WriteUnescaped(" &bull; ")
		}

		switch {
		case strings.Contains(link, "brid.gy"):
			hb.WriteElementOpen("data", "value", link, "class", "u-syndication hide")
			hb.WriteElementClose("data")
			hb.WriteElementOpen("link", "rel", "alternate", "type", "application/activity+json", "href", "https://fed.brid.gy/r/"+rc.GetURL())
			hb.WriteElementOpen("i", "class", "fediverse", "title", "The Fediverse")
			hb.WriteElementClose("i")
			// Post Summaries
			summary := post.GetFirstParameterValue("summary")
			if summary != "" {
				hb.WriteElementOpen("data", "value", summary, "class", "p-summary hide")
				hb.WriteElementClose("data")
			}
		case strings.Contains(link, "rateyourmusic.com"):
			hb.WriteElementOpen("i", "class", "sonemic")
			hb.WriteElementClose("i")
			hb.WriteElementOpen("a", "href", link, "rel", "syndication", "class", "u-syndication", "target", "_blank", "title", "This post on RYM/Sonemic")
			hb.WriteUnescaped("RYM/Sonemic")
			hb.WriteElementClose("a")
		case strings.Contains(link, "todon"):
			hb.WriteElementOpen("i", "class", "mastodon")
			hb.WriteElementClose("i")
			hb.WriteElementOpen("a", "href", link, "rel", "syndication", "class", "u-syndication", "target", "_blank", "title", "This post on Mastodon")
			hb.WriteUnescaped("Mastodon")
			hb.WriteElementClose("a")
		// case strings.Contains(link, "twitter.com"):
		// 	hb.WriteElementOpen("a", "href", link, "rel", "syndication", "class", "u-syndication", "target", "_blank", "title", "This post on Twitter")
		// 	hb.WriteElementOpen("img", "src", "/assets/icons/twitter.svg", "alt", "Twitter logo", "style", "width: 1rem")
		// 	hb.WriteUnescaped("Twitter")
		// 	hb.WriteElementClose("a")
		case strings.Contains(link, "boxd"):
			hb.WriteElementOpen("i", "class", "letterboxd")
			hb.WriteElementClose("i")
			hb.WriteElementOpen("a", "href", link, "class", "u-syndication", "rel", "syndication", "target", "_blank", "title", "This post on Letterboxd")
			hb.WriteUnescaped("Letterboxd")
			hb.WriteElementClose("a")
		case strings.Contains(link, "micro.blog"):
			jsonURL := "https://micro.blog/webmention?target=https://kandr3s.co" + rc.GetPath()
			homePageURL := ""
			// Fetch the JSON file
			resp, err := http.Get(jsonURL)
			if err != nil {
				fmt.Printf("Failed to fetch JSON file: %v\n", err)
				return
			}
			defer resp.Body.Close()

			// Read the JSON data
			jsonData, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Failed to read JSON data: %v\n", err)
				return
			}

			// Parse the JSON data
			var data map[string]interface{}
			err = json.Unmarshal(jsonData, &data)
			if err != nil {
				hb.WriteElementOpen("a", "href", link, "rel", "syndication", "class", "u-syndication", "target", "_blank", "title", "This post on Micro.blog")
			} else {
				homePageURL, ok := data["home_page_url"].(string)
				if !ok {
					fmt.Println("Failed to extract home_page_url from JSON data")
					return
				}
				hb.WriteElementOpen("i", "class", "microblog")
				hb.WriteElementClose("i")
				hb.WriteElementOpen("a", "href", homePageURL, "rel", "syndication", "class", "u-syndication", "target", "_blank", "title", "This post on Micro.blog")

			}
			hb.WriteUnescaped("Micro.blog")
			hb.WriteElementClose("a")
		case strings.Contains(link, "spotify.com"):
			hb.WriteElementOpen("i", "class", "spotify")
			hb.WriteElementClose("i")
			hb.WriteElementOpen("a", "href", link, "rel", "syndication", "class", "u-syndication", "target", "_blank", "title", "Listen on Spotify")
			hb.WriteUnescaped("Spotify")
			hb.WriteElementClose("a")
		case strings.Contains(link, "utube"):
			hb.WriteElementOpen("i", "class", "youtube")
			hb.WriteElementClose("i")
			hb.WriteElementOpen("a", "href", link, "rel", "syndication", "class", "u-syndication", "target", "_blank", "title", "Listen on YouTube")
			hb.WriteUnescaped("Youtube")
			hb.WriteElementClose("a")
		default:
			hb.WriteElementOpen("data", "value", link, "class", "u-syndication hide")
			hb.WriteElementClose("data")
		}
	}

	hb.WriteElementClose("div")
	doc.Find("main.h-entry article").AppendHtml(buf.String())
}

// Syndicate on Post Creation
func (p *plugin) PostCreated(post plugintypes.Post) {
	p.syndicate(post)
}

// Syndicate on Post Update
func (p *plugin) PostUpdated(post plugintypes.Post) {
	// Check if Post has Webmention enabled
	webmention := post.GetFirstParameterValue("webmention")
	if webmention != "false" {
		p.syndicate(post)
	} else {
		fmt.Println("🔌 Syndication: Webmentions disabled on this post.")
	}
}

// Webmention Sender
func (p *plugin) syndicate(post plugintypes.Post) {
	source := p.app.GetBlogURL() + post.GetPath()
	syndicationTargets := p.app.GetSyndicationTargets()
	syndicationParam := post.GetParameter("syndication")

	if syndicationParam == nil {
		return
	}

	for _, syndicationLink := range syndicationParam {
		for _, target := range syndicationTargets {
			regex := regexp.MustCompile(`^(https?://[^/]+)`)
			match := regex.FindStringSubmatch(target)

			if len(match) >= 2 {
				webmentionEndpoint := match[1] + "/webmention"
				if strings.Contains(target, syndicationLink) {
					data := map[string][]string{
						"source": {source},
						"target": {target},
					}

					resp, err := http.PostForm(webmentionEndpoint, data)
					if err != nil {
						fmt.Println("🔌 Syndication: Error sending webmention:", err)
						continue
					}

					fmt.Printf("🔌 Syndication: Webmention sent to '%s'. Result: %s\n", webmentionEndpoint, resp.Status)
					if err != nil {
						fmt.Println("🔌 Syndication: Error sending Webmention:", err)
						fmt.Println("Sending Ping instead...")
						req, err := http.NewRequest("POST", webmentionEndpoint, nil)
						if err != nil {
							fmt.Println("Error sending ping request:", err)
							return
						}
						fmt.Printf("Ping sent to '%s' - Result: %s", webmentionEndpoint, req.Response.Status)
						continue
					}
				}
			}
		}
	}
}
