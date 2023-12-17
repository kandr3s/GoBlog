package trakt

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.goblog.app/app/pkgs/plugintypes"
)

type plugin struct {
	app         plugintypes.App
	api         string
	token       string
	lastWatched *WatchHistoryItem
}

type WatchHistoryItem struct {
	ID        int    `json:"id"`
	WatchedAt string `json:"watched_at"`
	Action    string `json:"action"`
	Type      string `json:"type"`
	Movie     struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
	} `json:"movie"`
	Episode struct {
		Season int    `json:"season"`
		Number int    `json:"number"`
		Title  string `json:"title"`
	} `json:"episode"`
	Show struct {
		Title string `json:"title"`
	} `json:"show"`
}

func GetPlugin() (plugintypes.SetConfig, plugintypes.SetApp, plugintypes.UI2, plugintypes.Exec) {
	p := &plugin{}
	return p, p, p, p
}

func (p *plugin) SetApp(app plugintypes.App) {
	p.app = app

	// Start ticker to fetch last loved track every hour
	tickerLastWatched := time.NewTicker(1 * time.Hour)
	doneLastWatched := make(chan bool)

	// Fetch Last Loved Track periodically
	go func() {
		for {
			select {
			case <-doneLastWatched:
				return
			case t := <-tickerLastWatched.C:
				// fmt.Println("🔌 NowPlaying: Fetched ❤ Track at", t)
				p.fetchLastWatchedEpisode()
			}
		}
	}()
}

func (p *plugin) SetConfig(config map[string]any) {
	for key, value := range config {
		switch key {
		case "token":
			p.token = value.(string)
		case "api":
			p.api = value.(string)
		default:
			fmt.Println("Unknown config key:", key)
		}
	}
}

func (p *plugin) Exec() {
	p.fetchLastWatchedEpisode()
}

func (p *plugin) RenderWithDocument(rc plugintypes.RenderContext, doc *goquery.Document) {
	path := rc.GetPath()
	if strings.Contains(path, "/watches") && p.lastWatched != nil {

		// Display Trakt.TV info
		doc.Find("main.h-feed ul").AppendHtml(fmt.Sprintf("<li id=now>📺 <a href=/watches/tv>TV Log</a>: <b>%s</b> (%dx%d: %s)</li>", p.lastWatched.Show.Title, p.lastWatched.Episode.Season, p.lastWatched.Episode.Number, p.lastWatched.Episode.Title))
	} else {
		return
	}
}

func (p *plugin) fetchLastWatchedEpisode() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.trakt.tv/users/me/history", nil)
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("trakt-api-key", ""+p.api)
	req.Header.Add("trakt-api-version", "2")
	req.Header.Add("Authorization", "Bearer "+p.token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Request failed with status: %s\n", resp.Status)

		// Print the response body for debugging
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Response body: %s\n", string(body))

		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
		return
	}

	var watchHistory []WatchHistoryItem
	err = json.Unmarshal(body, &watchHistory)
	if err != nil {
		fmt.Printf("Error decoding JSON response: %s\n", err)
		return
	}

	// Retrieve the most recent episode (first item in the response)
	if len(watchHistory) == 0 {
		fmt.Println("No watched episodes found")
		return
	}

	p.lastWatched = &watchHistory[0]
}
