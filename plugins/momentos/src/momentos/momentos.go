package momentos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.goblog.app/app/pkgs/bufferpool"
	"go.goblog.app/app/pkgs/htmlbuilder"
	"go.goblog.app/app/pkgs/plugintypes"
)

func GetPlugin() (plugintypes.SetConfig, plugintypes.UI2, plugintypes.UISummary, plugintypes.SetApp, plugintypes.Exec) {
	p := &plugin{}
	return p, p, p, p, p
}

type plugin struct {
	app       plugintypes.App
	config    map[string]any
	followers string
	following string
}

type Icon struct {
	URL string `json:"url"`
}

type Item struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon Icon   `json:"icon"`
	ID   string
}

type BridgyResponse struct {
	First struct {
		Items []Item `json:"items"`
		Next  string `json:"next"`
	} `json:"first"`
	Items []Item `json:"items"`
	Next  string `json:"next"`
}

func (p *plugin) SetConfig(config map[string]any) {
	p.config = config
}

func (p *plugin) Exec() {
	following, err := fetchBridgyData("https://fed.brid.gy/kandr3s.co/following")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	p.following = following

	followers, err := fetchBridgyData("https://fed.brid.gy/kandr3s.co/followers")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	p.followers = followers

}

// Posts Summaries
func (p *plugin) RenderSummaryForPost(rc plugintypes.RenderContext, post plugintypes.Post, doc *goquery.Document) {

	doc.Find(".h-entry").Each(func(i int, s *goquery.Selection) {
		sectionTitle := doc.Find(".post-section").Text()
		if len(sectionTitle) == 0 {
			return
		}
		sectionEmoji := string([]rune(sectionTitle)[0])
		permalink := doc.Find("a.permalink").AttrOr("href", "")
		hidden := doc.Find("p.hide")
		date := doc.Find(".post-footer time").Text()
		time := doc.Find(".post-footer time").AttrOr("datetime", "")
		postfooter := doc.Find(".post-footer")
		visibility := doc.Find(".visibility")
		linkSection := post.GetSection()

		// Read More link on Articles
		if linkSection == "articles" || linkSection == "listens" || linkSection == "reads" {
			doc.Find("div.e-content").AppendHtml(fmt.Sprintf(" <span class=readmore><a href=" + post.GetPath() + ">Read more...</a></span>"))
		}
		// Bookmarks/Links
		if linkSection == "bookmarks" || linkSection == "links" {
			external := post.GetFirstParameterValue("link")
			doc.Find("h2.p-name a.u-url").Each(func(i int, s *goquery.Selection) {
				s.SetAttr("href", external)
				s.SetAttr("target", "_blank")
				s.SetAttr("rel", "noopener noreferrer")
				s.SetAttr("class", "external")
			})
			doc.Find(".h-entry p a.u-bookmark-of").ParentsFiltered("p").Remove()
		}

		// Timeline Template
		doc.Find("article").PrependHtml(fmt.Sprintf("<span class=\"category\"><a href=\"/%s\" title=\"%v\">%s</a></span><div class=post-footer>%s<a href=%s class=\"date permalink u-url\"><time class=\"dt-published\" datetime=\"%s\">%s</time>➜</a></div>", post.GetSection(), sectionTitle, sectionEmoji, visibility.Text(), permalink, time, date))
		postfooter.Remove()
		hidden.Remove()
	})
}

func (p *plugin) SetApp(app plugintypes.App) {
	p.app = app
}

func (p *plugin) RenderWithDocument(rc plugintypes.RenderContext, doc *goquery.Document) {
	blog := rc.GetBlog()

	// Hide TTS Button
	doc.Find("article div.actions").Remove()

	// Hide Updated Times
	doc.Find("meta[itemprop='datePublished']").Remove()

	if blog == "" {
		fmt.Println("🔌 Momentos: blog is empty!")
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
					// Enable Microsub support
					if name == "microsub" {
						buf.Reset()
						link := element["link"]
						hb.WriteElementOpen("link", "rel", "microsub", "href", link)
						doc.Find("head").AppendHtml(buf.String())
					}
					// Manifest (Enable PWA)
					if name == "manifest" {
						buf.Reset()
						hb.WriteElementOpen("link", "rel", "manifest", "href", "/manifest.json")
						doc.Find("head").AppendHtml(buf.String())
					}
					// Blog image
					if name == "avatar" {
						buf.Reset()
						hb.WriteElementOpen("a", "href", "/", "rel", "home", "logo", blog, "translate", "no")
						hb.WriteElementOpen("img", "src", "/profile.jpg", "class", "avatar", "alt", blog, "title", blog)
						hb.WriteElementClose("img")
						hb.WriteElementClose("a")
						doc.Find("header").PrependHtml(buf.String())
					}
					// Custom Menus
					if name == "menus" {
						buf.Reset()
						// Main menu
						doc.Find("header nav:first-of-type").AddClass("menu").SetAttr("id", "menu")
						hb.WriteElementOpen("label", "for", "show-menu", "class", "show-menu menu-icon", "alt", "Menu icon", "title", "Menu")
						hb.WriteUnescaped("☰ ") //
						hb.WriteElementClose("label")
						hb.WriteElementOpen("input", "type", "checkbox", "id", "show-menu", "role", "button")

						doc.Find("header nav:first-of-type").BeforeHtml(buf.String())
						// User Menu
						buf.Reset()
						userMenu := doc.Find("header nav:not(.menu)")
						if userMenu.Length() > 0 {
							userMenu.AddClass("user-menu")
							textMap := map[string]string{
								"/editor":        "📄",
								"/notifications": "🔔",
								"/webmention":    "🌐",
								"/comment":       "💬",
								"/settings":      "⚙️",
								"/logout":        "❌",
							}
							userMenu.Find("a").Each(func(i int, s *goquery.Selection) {
								href, _ := s.Attr("href")
								if text, ok := textMap[href]; ok {
									s.SetText(text)
								}
							})
						}
					}
					// Enable Infinite Scroll
					if name == "infiniteScroll" {
						buf.Reset()
						hb.WriteElementOpen("script", "src", "/assets/js/pagination.js")
						hb.WriteElementClose("script")
						doc.Find("body").AppendHtml(buf.String())
					}
					// Display Social Icons
					// TODO add config fields.
					if name == "socialicons" {
						buf.Reset()
						hb.WriteUnescaped("<div class=\"social-icons\"><a class=\"social-link\" href=\"/@kandr3s\" title=\"Mastodon\"><i class=\"mastodon\"></i></a><a class=\"social-link\" rel=\"nofollow\" target=\"_blank\" href=\"/sonemic\"><i class=\"sonemic\"></i></a><a class=\"social-link\" rel=\"nofollow\" target=\"_blank\" href=\"/letterboxd\"><i class=\"letterboxd\"></i></a><a class=\"social-link\" rel=\"nofollow\" target=\"_blank\" href=\"/github\" title=\"GitHub\"><i class=\"github\"></i></a><a class=\"social-link\" rel=\"nofollow\" target=\"_blank\" href=\"/spotify\" title=\"Spotify\"><i class=\"spotify\"></i></a><a class=\"social-link\" rel=\"nofollow\" target=\"_blank\" href=\"https://app.element.io/#/user/@kandr3s:matrix.org\" title=\"Matrix\"><i class=\"matrix\"></i></a></div>")
						doc.Find("footer nav").AfterHtml(buf.String())
					}
					// IndieWeb Icons/Banner
					if name == "indiewebicons" {
						buf.Reset()
						hb.WriteElementOpen("p")
						hb.WriteElementOpen("img", "src", "/images/indieweb-icons.jpg", "usemap", "#image-map")

						hb.WriteElementOpen("map", "name", "image-map")

						hb.WriteElementOpen("area", "shape", "rect", "coords", "0,0,80,320", "href", "https://goblog.app")
						hb.WriteElementOpen("area", "shape", "rect", "coords", "80,0,160,320", "href", "https://www.w3.org/TR/webmention/")
						hb.WriteElementOpen("area", "shape", "rect", "coords", "160,0,240,320", "href", "https://microformats.org/")
						hb.WriteElementOpen("area", "shape", "rect", "coords", "240,0,320,320", "href", "https://indieweb.org/")

						hb.WriteElementClose("map")
						hb.WriteElementClose("p")
						doc.Find("footer").AppendHtml(buf.String())
					}
					// Hello message
					if name == "hello" && rc.GetPath() == "/" {
						if pathValue, ok := element["path"].(string); ok {
							hello, err := p.app.GetPost(pathValue)
							if err == nil {
								doc.Find(".h-feed").PrependHtml(fmt.Sprintf("<article class=intro id=hello>%v</article>", hello.GetContent()))
							}
						}
					}
				}
			}

		}
	}
	// Display Bridgy Following
	if rc.GetPath() == "/@kandr3s" {
		doc.Find("#following ul").AppendHtml(fmt.Sprintf(p.following))
		doc.Find("#followers ul").AppendHtml(fmt.Sprintf(p.followers))

	}

	// Display Section Title in Menu
	selectSection := doc.Find("main.h-feed h1.p-name").Text()

	if selectSection != "" {
		doc.Find("header label.show-menu").AppendHtml(selectSection)
		doc.Find("main.h-feed h1.p-name").AddClass("hide")
	}
}

func fetchBridgyData(url string) (string, error) {
	var itemsHTML strings.Builder

	for url != "" {
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		var paginatedResponse BridgyResponse
		err = json.NewDecoder(resp.Body).Decode(&paginatedResponse)
		if err != nil {
			return "", err
		}

		// Extract items
		var responseItems []Item
		if len(paginatedResponse.First.Items) > 0 {
			responseItems = paginatedResponse.First.Items
			url = paginatedResponse.First.Next
		} else {
			responseItems = paginatedResponse.Items
			url = paginatedResponse.Next
		}

		// Process items and format as HTML
		r := regexp.MustCompile(`https://([^/]+)/@(.+)`)
		for _, item := range responseItems {
			match := r.FindStringSubmatch(item.URL)
			if len(match) > 2 {
				domain := match[1]
				name := match[2]

				item.ID = fmt.Sprintf("@%s@%s", name, domain)

				// Format item as HTML string
				htmlString := fmt.Sprintf("<li style=margin:10px auto><img loading=lazy src=%s width=34px height=34px style=vertical-align:middle> %s (<a href=%s target=_blank class=external>%s</a>)</li>", item.Icon.URL, item.Name, item.URL, item.ID)
				itemsHTML.WriteString(htmlString)
			}
		}
	}

	return itemsHTML.String(), nil
}
