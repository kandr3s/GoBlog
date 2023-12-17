package nowplaying

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/carlmjohnson/requests"
	"go.goblog.app/app/pkgs/bufferpool"
	"go.goblog.app/app/pkgs/htmlbuilder"
	"go.goblog.app/app/pkgs/plugintypes"
)

type plugin struct {
	app plugintypes.App

	apiKey         string
	user           string
	nowPlaying     *Track
	lastLovedTrack *LovedTrack
	topAlbumsChart string
	token          string // Micropub Token
	favorites      bool
	topalbums      bool
}

func GetPlugin() (plugintypes.SetConfig, plugintypes.SetApp, plugintypes.UI, plugintypes.UI2, plugintypes.Exec) {
	p := &plugin{}
	return p, p, p, p, p
}

type Lfm struct {
	Recenttracks *Recenttracks `xml:"recenttracks"`
}

type Recenttracks struct {
	Track []*Track `xml:"track"`
}

type Track struct {
	Nowplaying string `xml:"nowplaying,attr"`
	Artist     *struct {
		Text string `xml:",chardata"`
	} `xml:"artist"`
	Name  string `xml:"name"`
	Album *struct {
		Text string `xml:",chardata"`
	} `xml:"album"`
	URL   string `xml:"url"`
	Image []*struct {
		Text string `xml:",chardata"`
		Size string `xml:"size,attr"`
	} `xml:"image"`
	Date *struct {
		Uts string `xml:"uts,attr"`
	} `xml:"date"`
}

type LfmLoved struct {
	Lovedtracks *Lovedtracks `xml:"lovedtracks"` // Updated name: Lovedtracks -> lovedtracks
}

type Lovedtracks struct {
	Track []*LovedTrack `xml:"track"`
}
type LovedTrack struct {
	Artist *struct {
		Text string `xml:",chardata"`
		Name string `xml:"name"`
	} `xml:"artist"`
	Name string `xml:"name"`
	Url  string `xml:"url"`
}

type Album struct {
	Artist struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	} `json:"artist"`
	Image []struct {
		Size string `json:"size"`
		URL  string `json:"#text"`
	} `json:"image"`
	Name      string `json:"name"`
	MBID      string `json:"mbid"`
	URL       string `json:"url"`
	Playcount string `json:"playcount"`
	Attr      struct {
		Rank string `json:"rank"`
	} `json:"@attr"`
}

type LastFMTopAlbums struct {
	TopAlbums struct {
		Album []Album `json:"album"`
	} `json:"topalbums"`
}

type Listen struct {
	Path       string
	ListenLink string
}

func (p *plugin) SetConfig(config map[string]any) {

	for key, value := range config {
		switch key {
		case "key":
			p.apiKey = value.(string) // Last.FM API Key
		case "user":
			p.user = value.(string) //Last.FM User
		case "token":
			p.token = value.(string) // Micropub Token to OwnLovedTracks
		case "favorites":
			p.favorites = value.(bool)
		case "topalbums":
			p.topalbums = value.(bool)
		default:
			fmt.Println("Unknown config key:", key)
		}
	}

}

func (p *plugin) SetApp(app plugintypes.App) {
	p.app = app

	// Start ticker to refresh now playing every 2 minutes
	ticker := time.NewTicker(2 * time.Minute)
	done := make(chan bool)

	// Start ticker to fetch last loved track every hour
	tickerLastLoved := time.NewTicker(1 * time.Hour)
	doneLastLoved := make(chan bool)

	// Fetch Now Playing periodically
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				// fmt.Println("🔌 NowPlaying: Fetch now playing at", t)
				p.fetchNowPlaying()
			}
		}
	}()

	// Fetch Last Loved Track periodically
	go func() {
		for {
			select {
			case <-doneLastLoved:
				return
			case t := <-tickerLastLoved.C:
				// fmt.Println("🔌 NowPlaying: Fetched ❤ Track at", t)
				p.fetchLastLovedTrack()
			}
		}
	}()
}

func (p *plugin) Exec() {
	p.fetchNowPlaying()
	p.fetchLastLovedTrack()

	// Fetch Top Albums Chart on Run
	if p.topalbums == true {
		url := fmt.Sprintf("http://ws.audioscrobbler.com/2.0/?method=user.gettopalbums&user=%s&api_key=%s&period=1month&format=json", p.user, p.apiKey)

		response, err := http.Get(url)
		if err != nil {
			fmt.Println("Error fetching top albums:", err)
			return
		}
		defer response.Body.Close()

		var data LastFMTopAlbums
		err = json.NewDecoder(response.Body).Decode(&data)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}

		topAlbums := data.TopAlbums.Album
		if len(topAlbums) < 4 {
			fmt.Println("Not enough top albums found.")
			return
		}

		albumTemplate := `<div class="np-item">
				<img src="%s" loading=lazy alt="Album art">
				<div class="np-info">
					<span class="np-title">%s</span>
					<span class="np-artist">%s</span>
				</div>
			</div>
			`

		var albumsChart string
		for i := 0; i < 4; i++ {
			album := topAlbums[i]
			albumsChart += fmt.Sprintf(albumTemplate, album.Image[3].URL, album.Name, album.Artist.Name)
		}
		p.topAlbumsChart = albumsChart
	} else {
		// Top Albums Disabled
		fmt.Println("🔌 NowPlaying: Top Albums Chart disabled.")
	}
}

func (p *plugin) fetchNowPlaying() {
	// Check config
	if p == nil || p.apiKey == "" || p.user == "" {
		fmt.Println("🔌 NowPlaying: Not configured")
		return
	}
	// Remember previous playing
	hadPrevious := p.nowPlaying != nil
	previousUrl := ""
	if hadPrevious {
		previousUrl = p.nowPlaying.URL
	}
	// Create exit function that clears now playing and cache on errors
	exit := func() {
		p.nowPlaying = nil
		if hadPrevious {
			p.app.PurgeCache()
		}
	}
	// Fetch current now playing
	result := &Lfm{}
	pr, pw := io.Pipe()
	go func() {
		_ = pw.CloseWithError(
			requests.URL("http://ws.audioscrobbler.com/2.0/").
				Param("method", "user.getrecenttracks").
				Param("limit", "3").
				Param("user", p.user).
				Param("api_key", p.apiKey).
				Client(p.app.GetHTTPClient()).
				ToWriter(pw).
				Fetch(context.Background()),
		)
	}()
	err := xml.NewDecoder(pr).Decode(result)
	_ = pr.CloseWithError(err)
	if err != nil {
		exit()
		return
	}
	// Check result
	recents := result.Recenttracks
	if recents == nil {
		exit()
		return
	}
	tracks := recents.Track
	if tracks == nil {
		exit()
		return
	}
	p.nowPlaying = nil
	for _, track := range tracks {
		if track.Nowplaying != "true" {
			unixTimestamp, _ := strconv.ParseInt(track.Date.Uts, 10, 64)
			timestamp := time.Unix(int64(unixTimestamp), 0)
			if time.Since(timestamp) < 10*time.Minute {
				p.nowPlaying = track
				p.app.PurgeCache()
			} else {
				continue
			}
		}
		if track.URL != previousUrl {
			p.nowPlaying = track
			p.app.PurgeCache()
		}
		return
	}
	// Clear GoBlog cache
	if hadPrevious {
		p.app.PurgeCache()
	}
}

func (p *plugin) fetchLastLovedTrack() {
	// Fetch last loved track
	result := &LfmLoved{}
	pr, pw := io.Pipe()
	go func() {
		_ = pw.CloseWithError(
			requests.URL("https://ws.audioscrobbler.com/2.0/").
				Param("method", "user.getlovedtracks").
				Param("limit", "1").
				Param("user", p.user).
				Param("api_key", p.apiKey).
				Client(p.app.GetHTTPClient()).
				ToWriter(pw).
				Fetch(context.Background()),
		)
	}()
	err := xml.NewDecoder(pr).Decode(result)
	_ = pr.CloseWithError(err)
	if err != nil {
		fmt.Println("🔌 NowPlaying: Error fetching last loved track:", err)
		return
	}

	// Check result
	lovedTracks := result.Lovedtracks
	if lovedTracks == nil || len(lovedTracks.Track) == 0 {
		fmt.Println("🔌 NowPlaying: No loved tracks found.")
		return
	}

	// Save the last loved track to p.lastLovedTrack
	p.lastLovedTrack = lovedTracks.Track[0]

	if p.favorites == true {

		// Query DB for Last Owned Loved Track (posts with 'listenlink' parameter)
		query := `
		SELECT p.path, pp.value AS listenlink
		FROM posts AS p
		JOIN post_parameters AS pp ON p.path = pp.path
		WHERE pp.parameter = 'listenlink'
		ORDER BY p.published DESC
		LIMIT 1
		`

		row := p.app.GetDatabase().QueryRow(query)

		var ownListen Listen
		if err := row.Scan(&ownListen.Path, &ownListen.ListenLink); err != nil {
			fmt.Println(fmt.Errorf("failed to fetch last owned track: %w", err))
			return
		}

		syndicatedTrack := ownListen.ListenLink
		link := p.lastLovedTrack.Url

		// Post new Loved Track if it's not already syndicated
		if syndicatedTrack != link {
			// TODO: Create via DB query
			// Send data via HTTP POST / Micropub
			formData := url.Values{
				"tags":        {"NowPlaying", p.lastLovedTrack.Artist.Name},
				"visibility":  {"unlisted"},
				"listenlink":  {p.lastLovedTrack.Url},
				"section":     {"listens"},
				"content":     {"<span class=emoji-tag>🎵 </span> Now playing: <a class=\"external\" href=" + p.lastLovedTrack.Url + " rel=\"nofollow noreferrer\" target=\"_blank\">" + p.lastLovedTrack.Name + "</a> by " + "<b>" + p.lastLovedTrack.Artist.Name + "</b>"},
				"syndication": {"https://fed.brid.gy/web/kandr3s.co"},
			}

			// Create the Micropub request
			req, err := http.NewRequest("POST", p.app.GetBlogURL()+"/micropub", strings.NewReader(formData.Encode())) // GoBlog's Micropub Endpoint
			if err != nil {
				panic(fmt.Errorf("error creating request: %v", err))
			}

			if p.token != "" {
				// Set the authorization header
				req.Header.Set("Authorization", "Bearer "+p.token) // Micropub Token

				// Set the content type header
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				// Send the HTTP POST request
				client := &http.Client{}
				res, err := client.Do(req)
				if err != nil {
					panic(fmt.Errorf("error creating request: %v", err))
				}
				defer res.Body.Close()

				// Print the response
				fmt.Println("🔌 NowPlaying: New ❤ Track:", p.lastLovedTrack.Name, "by", p.lastLovedTrack.Artist.Name)
				fmt.Println("↳", res.Status)
			} else {
				fmt.Println("🔌 NowPlaying: No Micropub Token configured.")
			}
		} else {
			fmt.Println("🔌 NowPlaying: Loved Tracks up to date.")
			fmt.Println("↳ ", p.lastLovedTrack.Name+" by "+p.lastLovedTrack.Artist.Name)
		}
	} else {
		fmt.Println("🔌 NowPlaying: Loved Tracks disabled.")
	}
}

func (p *plugin) Render(rc plugintypes.RenderContext, rendered io.Reader, modified io.Writer) {
	if p.nowPlaying == nil {
		_, _ = io.Copy(modified, rendered)
		return
	}

	doc, err := goquery.NewDocumentFromReader(rendered)
	if err != nil {
		return
	}

	buf := bufferpool.Get()
	defer bufferpool.Put(buf)
	hb := htmlbuilder.NewHtmlBuilder(buf)

	hb.WriteElementOpen("a", "title", "🎧 Now playing: "+p.nowPlaying.Name+" by "+p.nowPlaying.Artist.Text, "rel", "nofollow ugc", "alt", "Now playing", "href", "/listens#nowplaying")
	hb.WriteElementOpen("i", "class", "np-now")
	hb.WriteElementsClose("i")
	hb.WriteElementClose("a")
	doc.Find("header h1 span").AppendHtml(buf.String())
	_ = goquery.Render(modified, doc.Selection)
}

func (p *plugin) RenderWithDocument(rc plugintypes.RenderContext, doc *goquery.Document) {

	path := rc.GetPath()
	if strings.Contains(path, "/listens") {

		track := p.nowPlaying

		buf := bufferpool.Get()
		defer bufferpool.Put(buf)
		hb := htmlbuilder.NewHtmlBuilder(buf)

		if p.nowPlaying == nil {
			// Nothing playing - Display last Loved Track
			if p.lastLovedTrack.Name != "" && p.lastLovedTrack.Artist.Text != "" {
				// TODO: Fetch Album art MBID to Album MBID to Cover Art Archive Image
				// hb.WriteElementOpen("img", "src", "https://lastfm.freetls.fastly.net/i/u/300x300/2a96cbd8b46e442fc41c2b86b821562f.png", "title", p.lastLovedTrack.Name+" by "+p.lastLovedTrack.Artist.Text, "alt", "Now playing image")
				// hb.WriteElementOpen("div", "class", "np-info")
				// hb.WriteElementOpen("span", "class", "np-title")
				// hb.WriteElementOpen("a", "class", "external", "rel", "nofollow", "href", p.lastLovedTrack.Url)
				// hb.WriteEscaped(p.lastLovedTrack.Name)
				// hb.WriteElementClose("a")
				// hb.WriteElementClose("span")
				// hb.WriteElementOpen("span", "class", "np-artist")
				// hb.WriteEscaped(p.lastLovedTrack.Artist.Name)
				// hb.WriteElementClose("span")
				// hb.WriteElementClose("div")
				doc.Find(".nowplaying details").SetAttr("open", "")
			} else {
				return
			}
		} else {
			trackImage := track.Image[len(track.Image)-1].Text
			if trackImage == "" {
				trackImage = "https://lastfm.freetls.fastly.net/i/u/64s/4128a6eb29f94943c9d206c08e625904.jpg"
			}
			hb.WriteElementOpen("img", "src", trackImage, "title", track.Name+" by "+track.Artist.Text, "alt", track.Album.Text+" album cover")
			hb.WriteElementOpen("div", "class", "np-info")
			hb.WriteElementOpen("span")
			hb.WriteUnescaped("🎵 ")
			hb.WriteElementOpen("a", "class", "external", "rel", "nofollow", "href", track.URL)
			hb.WriteEscaped(track.Name)
			hb.WriteElementClose("a")
			hb.WriteElementClose("span")
			hb.WriteElementOpen("span", "class", "np-artist")
			hb.WriteUnescaped("🎤 ")
			hb.WriteEscaped(track.Artist.Text)
			hb.WriteElementClose("span")
			hb.WriteElementClose("div")
		}
		doc.Find(".np-track").AppendHtml(buf.String())
		// Top Albums
		if p.topAlbumsChart != "" {
			doc.Find(".np-topalbums").AppendHtml(fmt.Sprintf(p.topAlbumsChart))
		}
	}
}
