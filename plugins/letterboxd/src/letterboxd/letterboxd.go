package letterboxd

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.goblog.app/app/pkgs/bufferpool"
	"go.goblog.app/app/pkgs/plugintypes"
)

type plugin struct {
	app           plugintypes.App
	parameterName string // Syndication parameter
	section       string // Watches section
	username      string // Letterboxd Username
	token         string // Micropub Token
	artpath       string // Film Poster image path
	posters       bool   // Use local Film posters
}

// Letterboxd RSS Feed Struct
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

// Feed Channel Struct
type Channel struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Link        string    `xml:"link"`
	Items       []RSSItem `xml:"item"`
}

// Feed Item data Struct
type RSSItem struct {
	Title            string  `xml:"title"`
	Link             string  `xml:"link"`
	Description      string  `xml:"description"`
	WatchedDate      string  `xml:"https://letterboxd.com watchedDate"`
	Rewatch          string  `xml:"https://letterboxd.com rewatch"`
	LetterboxdTitle  string  `xml:"https://letterboxd.com filmTitle"`
	LetterboxdYear   string  `xml:"https://letterboxd.com filmYear"`
	LetterboxdRating float32 `xml:"https://letterboxd.com memberRating"`
	TmdbID           string  `xml:"https://themoviedb.org movieId"`
}

type Watch struct {
	Path            string
	SyndicationLink string
}

func GetPlugin() (plugintypes.SetConfig, plugintypes.SetApp, plugintypes.UISummary, plugintypes.Exec, plugintypes.UIPost) {
	p := &plugin{}
	return p, p, p, p, p
}

func (p *plugin) SetConfig(config map[string]any) {
	p.parameterName = "syndication" // default

	for key, value := range config {
		switch key {
		case "section":
			p.section = value.(string)
		case "username":
			p.username = value.(string)
		case "token":
			p.token = value.(string)
		case "posters":
			p.posters = value.(bool)
		default:
			fmt.Println("Unknown config key:", key)
		}
	}
}

func (p *plugin) RenderPost(rc plugintypes.RenderContext, post plugintypes.Post, doc *goquery.Document) {

	section := post.GetSection()
	// Watches Posts
	if section == "watches" {
		p.app.PurgeCache()
		post, err := p.app.GetPost(rc.GetPath())
		if err != nil || post == nil {
			return
		}
		letterboxdLink, ok := post.GetParameters()[p.parameterName]
		if !ok || len(letterboxdLink) == 0 {
			return
		}
		if err != nil {
			fmt.Println("🔌 Letterboxd: " + err.Error())
			return
		}
		buf := bufferpool.Get()
		defer bufferpool.Put(buf)
		for _, link := range letterboxdLink {
			boxd := "boxd"
			if strings.Contains(link, boxd) {
				doc.Find("main.h-entry article div.e-content p img").SetAttr("alt", "Film Poster") // Add Microformat and Alt attr to Film Poster
			} else {
				break
			}
		}
		if post.GetFirstParameterValue("filmart") == "" && post.GetFirstParameterValue("syndication") != "" {
			// Fetch and save Film Poster image
			boxdArt, exists := doc.Find("img").Attr("src")
			if !exists {
				log.Fatal("src attribute not found")
			}
			if p.posters == true {
				p.fetchPoster(boxdArt)
			}
			p.app.SetPostParameter(post.GetPath(), "filmart", []string{p.artpath})
		}
		if p.posters == true {
			doc.Find(".e-content p img").SetAttr("src", post.GetFirstParameterValue("filmart"))
		}
	}
}

func (p *plugin) RenderSummaryForPost(rc plugintypes.RenderContext, post plugintypes.Post, doc *goquery.Document) {
	section := post.GetSection()
	if section == "watches" {
		doc.Find(".e-content p img").SetAttr("alt", "Film Poster").SetAttr("height", "380px")

		watchTitle := doc.Find("h2.p-name").Text()
		if watchTitle != "" {
			doc.Find(".e-content").PrependHtml(fmt.Sprintf("<p>%s</p>", watchTitle))
		}
		doc.Find("h2.p-name").Remove() // Remove title
		// Remove non-review paragraph
		lastP := doc.Find(".e-content p").Last()
		if strings.HasPrefix(lastP.Text(), "Watched on ") {
			lastP.Remove()
		}
		if post.GetFirstParameterValue("filmart") == "" && post.GetFirstParameterValue("syndication") != "" && p.posters == true {
			// Fetch and save Film Poster image
			boxdArt, exists := doc.Find("img").Attr("src")
			if !exists {
				log.Fatal("src attribute not found")
			}
			p.fetchPoster(boxdArt)
			p.app.SetPostParameter(post.GetPath(), "filmart", []string{p.artpath})
		}
	}
}

func (p *plugin) SetApp(app plugintypes.App) {
	p.app = app

	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				p.fetchWatches()
			}
		}
	}()
}

func (p *plugin) Exec() {
	p.fetchWatches()
}

func (p *plugin) fetchWatches() {
	// Fetch & Parse Letterboxd RSS feed
	resp, err := http.Get("https://letterboxd.com/" + p.username + "/rss")
	if err != nil {
		fmt.Println("Error fetching Letterboxd feed:", err)
		return
	}
	defer resp.Body.Close()
	rss := RSS{}
	err = xml.NewDecoder(resp.Body).Decode(&rss)
	if err != nil {
		fmt.Println("Error parsing Letterboxd feed:", err)
		return
	}
	// Find the last item & extract fields from Letterboxd feed
	lastItem := rss.Channel.Items[:1]
	for _, item := range lastItem {
		title := item.Title
		film := item.LetterboxdTitle
		year := item.LetterboxdYear
		link := item.Link
		watchedDate := item.WatchedDate
		rewatch := item.Rewatch
		tmdbID := item.TmdbID
		watchStatus := "Watched "

		if rewatch == "Yes" {
			watchStatus = "Rewatched "
		}

		// Convert Rating
		converRating := item.LetterboxdRating * 2
		rating := fmt.Sprint(converRating)
		// Extract Film poster from the Description field
		re := regexp.MustCompile(`<img.*?src="(.*?)".*?>`)
		matches := re.FindStringSubmatch(item.Description)
		filmArt := ""
		if len(matches) > 1 {
			filmArt = matches[1]
		}
		// Set Description and Slug values
		desc := item.Description
		watchName := strings.Replace(strings.TrimPrefix(link, "https://letterboxd.com/"+p.username+"/film/"), "/", "", -1)
		slug := fmt.Sprintf("%s-%s", watchedDate, watchName)

		// Set Publishing Date
		publishedDate := ""
		today := time.Now().Format("2006-01-02")

		if watchedDate != today {
			publishedDate = watchedDate + "T21:00:00+02:00"
		} else {
			publishedDate = today
		}
		// Fetch Last Owned Watch
		query := `
		SELECT p.path, pp.value AS syndication
		FROM posts AS p
		JOIN post_parameters AS pp ON p.path = pp.path
		WHERE p.section = 'watches' AND pp.parameter = 'syndication'
		ORDER BY p.published DESC
		LIMIT 1
		`
		row := p.app.GetDatabase().QueryRow(query)

		var ownWatch Watch
		if err := row.Scan(&ownWatch.Path, &ownWatch.SyndicationLink); err != nil {
			fmt.Println(fmt.Errorf("🔌 Letterboxd: Failed to fetch last owned watch: %w", err))
			return
		}
		// TODO: Create via DB query
		// Update Watches with new entry
		if ownWatch.SyndicationLink != link {
			p.fetchPoster(filmArt)
			// Send data via HTTP POST / Micropub
			formData := url.Values{
				"section":     {p.section},
				"slug":        {slug},
				"published":   {publishedDate},
				"syndication": {link},
				"title":       {"🍿 " + watchStatus + title},
				"film":        {film},
				"year":        {year},
				"rating":      {rating},
				"content":     {desc},
				"filmart":     {p.artpath},
				"tmbd":        {tmdbID}, // TODO: Use ID to fetch Backdrops
			}
			// Create and send request
			req, err := http.NewRequest("POST", p.app.GetBlogURL()+"/micropub", strings.NewReader(formData.Encode())) // GoBlog's Micropub Endpoint
			if err != nil {
				panic(fmt.Errorf("error creating request: %v", err))
			}
			// Set headers
			req.Header.Set("Authorization", "Bearer "+p.token) // Micropub Token
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Send the HTTP POST request
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				panic(fmt.Errorf("error creating request: %v", err))
			}
			defer res.Body.Close()
			fmt.Println("🔌 Letterboxd: New Watch fetched:", title)
		} else {
			fmt.Println("🔌 Letterboxd: Watches up to date.")
		}
	}
}

func (p *plugin) fetchPoster(boxdArt string) {
	_, file := path.Split(boxdArt)
	artIDMatch := regexp.MustCompile(`^(.+?)-0-600-0-900`).FindStringSubmatch(file)
	var artID string

	if len(artIDMatch) > 1 {
		artID = artIDMatch[1]
	} else {
		artID := regexp.MustCompile(`^([^\-]+)`).FindStringSubmatch(file)[1]
	}
	slugFilename := fmt.Sprintf("%s.jpg", artID)

	outputDir := "./static/images/art/films/"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}
	outputPath := path.Join(outputDir, slugFilename)

	// Download Film Art
	response, err := http.Get(boxdArt)
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
	fmt.Printf("🔌 Letterboxd: Film Art saved to: %s\n", outputPath)
	// Set Film Art URL
	p.artpath = regexp.MustCompile(`(^|/)static(/|$)`).ReplaceAllString(outputPath, "/")
}
