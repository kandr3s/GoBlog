package discoteca

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"go.goblog.app/app/pkgs/bufferpool"
	"go.goblog.app/app/pkgs/htmlbuilder"
	"go.goblog.app/app/pkgs/plugintypes"
)

type plugin struct {
	app           plugintypes.App
	parameterName string
}

type iTunesResponse struct {
	Results []struct {
		ArtworkUrl100 string `json:"artworkUrl100"`
	} `json:"results"`
}

func GetPlugin() (plugintypes.SetConfig, plugintypes.SetApp, plugintypes.UIPost, plugintypes.UISummary) {
	p := &plugin{}
	return p, p, p, p
}

func (p *plugin) SetApp(app plugintypes.App) {
	p.app = app
}

func (p *plugin) SetConfig(config map[string]interface{}) {
	p.parameterName = "section" // default
	if configParameter, ok := config["parameter"].(string); ok {
		p.parameterName = configParameter // override default from config
	}
}

func (p *plugin) RenderPost(rc plugintypes.RenderContext, post plugintypes.Post, doc *goquery.Document) {
	section := post.GetSection()
	if section == "listens" {
		params := post.GetParameters()
		var album, artist, year, rating, albumstyle string
		if albumSlice, ok := params["album"]; ok {
			album = albumSlice[0]
		} else {
			return
		}
		if artistSlice, ok := params["artist"]; ok {
			artist = artistSlice[0]
		}
		if yearSlice, ok := params["year"]; ok {
			year = yearSlice[0]
		}
		if ratingSlice, ok := params["rating"]; ok {
			rating = ratingSlice[0]
		}
		if albumstyleSlice, ok := params["albumstyle"]; ok {
			albumstyle = albumstyleSlice[0]
		}
		albumartSlice, albumartOk := params["albumart"]
		artworkURL := ""

		// Get AlbumArt from Itunes
		if !albumartOk {
			// Build the URL for the iTunes Search API
			baseURL := "https://itunes.apple.com/search"
			query := url.Values{}
			query.Add("term", fmt.Sprintf("%s %s", artist, album))
			query.Add("entity", "album")
			query.Add("limit", "1")
			url := fmt.Sprintf("%s?%s", baseURL, query.Encode())

			// Send a GET request to the iTunes Search API
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("Error fetching album cover:", err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			// Decode the JSON response
			var iTunesResp iTunesResponse
			if err := json.NewDecoder(resp.Body).Decode(&iTunesResp); err != nil {
				fmt.Println("Error decoding iTunes response:", err)
				os.Exit(1)
			}

			// Print the URL of the album cover
			if len(iTunesResp.Results) > 0 {
				artworkURL = iTunesResp.Results[0].ArtworkUrl100
				re := regexp.MustCompile("100x100bb\\.jpg")
				artworkURL = re.ReplaceAllString(artworkURL, "316x316bb.webp")
			} else {
				fmt.Println("No results found")
			}

			// Build Listens HTML Template
			buf := bufferpool.Get()
			defer bufferpool.Put(buf)
			hb := htmlbuilder.NewHtmlBuilder(buf)
			hb.WriteElementOpen("p", "class", "p-summary hide")
			hb.WriteUnescaped("🎵 Listened to " + album + " by " + artist)
			hb.WriteElementClose("p")
			hb.WriteElementOpen("div", "class", "e-content")
			hb.WriteElementOpen("div", "class", "album-details")
			hb.WriteElementOpen("p", "class", "vinyl-case")
			hb.WriteElementOpen("img", "src", artworkURL, "class", "album-art", "loading", "lazy", "width", "150px", "height", "150px")
			hb.WriteElementClose("p")

			hb.WriteElementOpen("div", "class", "album-info")
			hb.WriteElementOpen("p")
			hb.WriteElementOpen("img", "src", "/assets/icons/ratings/"+rating+".png", "alt", "Album rating", "title", "Album rating")
			hb.WriteElementClose("p")
			hb.WriteElementOpen("p", "class", "album-title")
			hb.WriteUnescaped(album)
			hb.WriteElementClose("p")
			hb.WriteElementOpen("p", "class", "album-artist")
			hb.WriteUnescaped(artist)
			hb.WriteElementOpen("p", "class", "album-meta")
			hb.WriteUnescaped(year + " · " + albumstyle)
			hb.WriteElementClose("p")
			hb.WriteElementClose("p")
			hb.WriteElementClose("div")

			doc.Find(".e-content").PrependHtml(buf.String())
		} else {
			artworkURL = albumartSlice[0]
		}
	}
}

func (p *plugin) RenderSummaryForPost(rc plugintypes.RenderContext, post plugintypes.Post, doc *goquery.Document) {
	section := post.GetSection()
	if section == "listens" {
		params := post.GetParameters()
		var album, artist, year, rating, albumstyle string
		if albumSlice, ok := params["album"]; ok {
			album = albumSlice[0]
		} else {
			return
		}
		if artistSlice, ok := params["artist"]; ok {
			artist = artistSlice[0]
		}
		if yearSlice, ok := params["year"]; ok {
			year = yearSlice[0]
		}
		if ratingSlice, ok := params["rating"]; ok {
			rating = ratingSlice[0]
		}
		if albumstyleSlice, ok := params["albumstyle"]; ok {
			albumstyle = albumstyleSlice[0]
		}
		albumartSlice, albumartOk := params["albumart"]
		artworkURL := ""

		// Get AlbumArt from iTunes
		if !albumartOk {
			// Build the URL for the iTunes Search API
			baseURL := "https://itunes.apple.com/search"
			query := url.Values{}
			query.Add("term", fmt.Sprintf("%s %s", artist, album))
			query.Add("entity", "album")
			query.Add("limit", "1")
			url := fmt.Sprintf("%s?%s", baseURL, query.Encode())

			// Send a GET request to the iTunes Search API
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("Error fetching album cover:", err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			// Decode the JSON response
			var iTunesResp iTunesResponse
			if err := json.NewDecoder(resp.Body).Decode(&iTunesResp); err != nil {
				fmt.Println("Error decoding iTunes response:", err)
				os.Exit(1)
			}

			// Print the URL of the album cover
			if len(iTunesResp.Results) > 0 {
				artworkURL = iTunesResp.Results[0].ArtworkUrl100
				re := regexp.MustCompile("100x100bb\\.jpg")
				artworkURL = re.ReplaceAllString(artworkURL, "316x316bb.webp")
			} else {
				fmt.Println("No results found")
			}

			// Build Listens HTML Template
			buf := bufferpool.Get()
			defer bufferpool.Put(buf)
			hb := htmlbuilder.NewHtmlBuilder(buf)
			hb.WriteElementOpen("p", "class", "p-summary hide")
			hb.WriteUnescaped("🎵 Listened to " + album + " by " + artist)
			hb.WriteElementClose("p")
			hb.WriteElementOpen("div", "class", "e-content")
			hb.WriteElementOpen("div", "class", "album-details")
			hb.WriteElementOpen("p", "class", "vinyl-case")
			hb.WriteElementOpen("img", "src", artworkURL, "class", "album-art", "loading", "lazy", "width", "150px", "height", "150px")
			hb.WriteElementClose("p")

			hb.WriteElementOpen("div", "class", "album-info")
			hb.WriteElementOpen("p")
			hb.WriteElementOpen("img", "src", "/assets/icons/ratings/"+rating+".png", "alt", "Album rating", "title", "Album rating")
			hb.WriteElementClose("p")
			hb.WriteElementOpen("p", "class", "album-title")
			hb.WriteUnescaped(album)
			hb.WriteElementClose("p")
			hb.WriteElementOpen("p", "class", "album-artist")
			hb.WriteUnescaped(artist)
			hb.WriteElementOpen("p", "class", "album-meta")
			hb.WriteUnescaped(year + " · " + albumstyle)
			hb.WriteElementClose("p")
			hb.WriteElementClose("p")
			hb.WriteElementClose("div")

			doc.Find(".e-content").PrependHtml(buf.String())
		} else {
			artworkURL = albumartSlice[0]
		}
	}
}
