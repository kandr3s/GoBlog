package discoteca

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.goblog.app/app/pkgs/bufferpool"
	"go.goblog.app/app/pkgs/htmlbuilder"
	"go.goblog.app/app/pkgs/plugintypes"
)

type plugin struct {
	app      plugintypes.App
	artwork  string
	albumart string
}

type iTunesResponse struct {
	Results []struct {
		ArtworkUrl100 string `json:"artworkUrl100"`
	} `json:"results"`
}

func GetPlugin() (plugintypes.SetApp, plugintypes.UIPost, plugintypes.UISummary) {
	p := &plugin{}
	return p, p, p
}

func (p *plugin) SetApp(app plugintypes.App) {
	p.app = app
}

func (p *plugin) RenderPost(rc plugintypes.RenderContext, post plugintypes.Post, doc *goquery.Document) {
	section := post.GetSection()
	if section == "listens" {
		album := post.GetFirstParameterValue("album")
		artist := post.GetFirstParameterValue("artist")
		year := post.GetFirstParameterValue("year")
		rating := post.GetFirstParameterValue("rating")
		albumstyle := post.GetFirstParameterValue("albumstyle")
		artworkURL := post.GetFirstParameterValue("albumart")

		if album == "" {
			return
		} else {
			// Get AlbumArt from Itunes
			if artworkURL == "" {
				p.fetchAlbumArt(album, artist, year)
				p.app.SetPostParameter(post.GetPath(), "albumart", []string{p.albumart})
				p.app.PurgeCache()
			}
			if p.albumart != "" {
				artworkURL = p.albumart
			}
		}

		// Build Listens HTML Template
		doc.Find(".p-name").Remove()
		buf := bufferpool.Get()
		defer bufferpool.Put(buf)
		hb := htmlbuilder.NewHtmlBuilder(buf)
		hb.WriteElementOpen("div", "class", "album-details")
		hb.WriteElementOpen("p", "class", "vinyl-case")
		hb.WriteElementOpen("img", "src", artworkURL, "class", "album-art", "loading", "lazy", "width", "150px", "height", "150px")
		hb.WriteElementClose("p")

		hb.WriteElementOpen("div", "class", "album-info")
		if rating != "" {
			hb.WriteElementOpen("p")
			hb.WriteElementOpen("img", "src", "/assets/icons/ratings/"+rating+".png", "alt", "Album rating", "title", "Album rating")
			hb.WriteElementClose("p")
		}
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
		hb.WriteElementClose("div")

		doc.Find(".e-content").PrependHtml(buf.String())
	}
}

func (p *plugin) RenderSummaryForPost(rc plugintypes.RenderContext, post plugintypes.Post, doc *goquery.Document) {
	section := post.GetSection()
	if section == "listens" {
		album := post.GetFirstParameterValue("album")
		artist := post.GetFirstParameterValue("artist")
		year := post.GetFirstParameterValue("year")
		rating := post.GetFirstParameterValue("rating")
		albumstyle := post.GetFirstParameterValue("albumstyle")
		artworkURL := post.GetFirstParameterValue("albumart")

		if album == "" {
			return
		} else {
			if artworkURL == "" {
				p.fetchAlbumArt(album, artist, year)
				p.app.SetPostParameter(post.GetPath(), "albumart", []string{p.albumart})
				p.app.PurgeCache()
			}
			// Build Listens HTML Template
			doc.Find(".p-name").Remove()
			doc.Find(".e-content").PrependHtml(fmt.Sprintf("<p class=p-summary>%s</p><div class=album-details><p class=vinyl-case><img src=%s class=album-art loading=lazy width=150px height=150px /></p><div class=album-info><p><img src=/assets/icons/ratings/%s.png alt=Rating><p class=album-title>%s</p><p class=album-artist>%s</p><p class=album-meta>%s · %s</p></div>", post.GetTitle(), artworkURL, rating, album, artist, year, albumstyle))
		}
	}
}

func (p *plugin) fetchAlbumArt(album, artist, year string) {
	// Build iTunes API Request
	baseURL := "https://itunes.apple.com/search"
	query := url.Values{}
	query.Add("term", fmt.Sprintf("%s %s", artist, album))
	query.Add("entity", "album")
	query.Add("limit", "1")
	url := fmt.Sprintf("%s?%s", baseURL, query.Encode())

	// Send iTunes API request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching album cover:", err)
		return
	}
	defer resp.Body.Close()

	// Decode iTunes API response
	var iTunesResp iTunesResponse // Define iTunesResponse struct
	if err := json.NewDecoder(resp.Body).Decode(&iTunesResp); err != nil {
		fmt.Println("Error decoding iTunes response:", err)
		return
	}

	// Get Album Art
	if len(iTunesResp.Results) > 0 {
		artworkURL := iTunesResp.Results[0].ArtworkUrl100
		p.artwork = artworkURL // Set arworkURL

		// Set filename
		re := regexp.MustCompile("100x100bb\\.jpg")
		artworkURL = re.ReplaceAllString(artworkURL, "316x316bb.webp")
		_, file := path.Split(artworkURL)
		slugFilename := fmt.Sprintf("%s_%s_%s", slugify(album), slugify(artist), file)

		outputDir := "./static/images/art/albums/"
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			return
		}
		outputPath := path.Join(outputDir, slugFilename)

		// Download Album Art
		response, err := http.Get(artworkURL)
		if err != nil {
			fmt.Printf("Error downloading image: %v\n", err)
			return
		}
		defer response.Body.Close()

		// Create the output file
		outputFile, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			return
		}
		defer outputFile.Close()

		// Copy the image data to the output file
		_, err = io.Copy(outputFile, response.Body)
		if err != nil {
			fmt.Printf("Error saving image: %v\n", err)
			return
		}
		fmt.Printf("🔌 Discoteca: Album Art saved to: %s\n", outputPath)
		// Set Album Art URL
		p.albumart = regexp.MustCompile(`(^|/)static(/|$)`).ReplaceAllString(outputPath, "/")
	} else {
		fmt.Println("🔌 Discoteca: No artwork found for: ", artist, "-", album, "[", year, "]")
	}
}

func slugify(s string) string {
	// Return alphanumeric-only dashed and lowercased
	regExp := regexp.MustCompile("[^a-zA-Z0-9]+")
	s = regexp.MustCompile("\\s+").ReplaceAllString(s, "-")
	s = regexp.MustCompile("-+").ReplaceAllString(regExp.ReplaceAllString(s, "-"), "-")
	return strings.ToLower(s)
}
