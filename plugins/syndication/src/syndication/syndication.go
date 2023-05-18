package syndication

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	for _, link := range syndicationLinks {
		rym := "rateyourmusic.com"
		masto := "todon.org"
		twitter := "twitter.com"
		boxd := "boxd"
		micro := "micro.blog"
		bridgy := "brid.gy"
		if strings.Contains(link, bridgy) {
			hb.WriteElementOpen("a", "href", link, "rel", "syndication", "title", "This post is part of The Fediverse")
			hb.WriteElementClose("a")
		}
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

// Syndicate on Post Creation
func (p *plugin) PostCreated(post plugintypes.Post) {
	p.syndicate(post)
}

// Syndicate on Post Update
func (p *plugin) PostUpdated(post plugintypes.Post) {
	p.syndicate(post)
}

// Webmention Sender
func (p *plugin) syndicate(post plugintypes.Post) {
	source := p.app.GetBlogURL() + post.GetPath()
	syndicateTo := p.app.GetSyndicationTargets()
	targets := post.GetParameter("syndication")

	for _, target := range targets {
		if contains(syndicateTo, target) {
			endpoint, err := discoverWebmentionEndpoint(target)
			if err != nil {
				// TODO handle non-webmention Syndication Targets
				fmt.Println("Syndication plugin: Error discovering Webmention endpoint:", err)
				continue
			}

			data := map[string][]string{
				"source": {source},
				"target": {target},
			}

			resp, err := http.PostForm(endpoint, data)
			if err != nil {
				fmt.Println("Syndication plugin: Error sending webmention:", err)
				continue
			}

			fmt.Printf("Syndication plugin: Webmention sent. Result: %s", resp.Status)
		}
	}
}

// Check Syndication Target
func contains(array []string, target string) bool {
	for _, item := range array {
		if item == target {
			return true
		}
	}
	return false
}

// Discover Webmention Endpoint
func discoverWebmentionEndpoint(targetURL string) (string, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	endpoint, err := parseWebmentionEndpoint(resp.Body, targetURL) // Pass targetURL as baseURL argument
	if err != nil {
		return "", err
	}

	// Resolve the endpoint URL relative to the target URL
	resolvedURL, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	resolvedURL = u.ResolveReference(resolvedURL)

	return resolvedURL.String(), nil
}

// Parse response and find Webmention endpoint
func parseWebmentionEndpoint(body io.Reader, baseURL string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return "", err
	}

	endpoint := ""
	foundEndpoint := false

	doc.Find("link[rel=webmention]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && !foundEndpoint {
			u, err := url.Parse(href)
			if err != nil {
				return
			}

			if u.IsAbs() {
				endpoint = href
			} else {
				base, err := url.Parse(baseURL)
				if err != nil {
					return
				}

				endpoint = base.ResolveReference(u).String()
			}

			foundEndpoint = true
		}
	})

	return endpoint, nil
}
