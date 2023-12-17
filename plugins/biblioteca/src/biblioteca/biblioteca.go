package biblioteca

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.goblog.app/app/pkgs/plugintypes"
)

type plugin struct {
	app        plugintypes.App
	reading    Read
	readStatus string
	artpath    string
}

type Read struct {
	Path   string
	Shelf  string
	Title  string
	Author string
	Cover  string
}

func GetPlugin() (plugintypes.UIPost, plugintypes.UISummary, plugintypes.UI2, plugintypes.Exec, plugintypes.SetApp, plugintypes.PostCreatedHook, plugintypes.PostUpdatedHook) {
	p := &plugin{}
	return p, p, p, p, p, p, p
}

func (p *plugin) SetApp(app plugintypes.App) {
	p.app = app
}

func (p *plugin) Exec() {
	p.fetchReading()
}

func (p *plugin) RenderWithDocument(rc plugintypes.RenderContext, doc *goquery.Document) {
	path := rc.GetPath()

	if strings.Contains(path, "/reads") && p.reading.Shelf == "reading" {

		doc.Find("main.h-feed hr").AfterHtml(fmt.Sprintf("<article class=\"intro\"><span class=category>📖</span><span class=spotlight-text><b>Currently reading</b></span><div class=\"read-of h-cite book-details\" value=\"reading\"><img class=\"book-cover reading\" src=\"%s\"><div class=\"book-details-text\"><a class=\"p-name external\" href=\"%s\">%s</a> by %v</div></div>", p.reading.Cover, p.reading.Path, p.reading.Title, p.reading.Author))
	}
}

func (p *plugin) RenderPost(rc plugintypes.RenderContext, post plugintypes.Post, doc *goquery.Document) {

	section := post.GetSection()
	// Reads Summaries
	if section == "reads" {
		doc.Find("h1.p-name").Remove() // Remove title

		switch post.GetFirstParameterValue("shelf") {
		case "read":
			p.readStatus = "Finished reading:"
		case "reading":
			p.readStatus = "Currently reading:"
		case "to-read":
			p.readStatus = "Want to read:"
		default:
			p.readStatus = "Added a book to La Biblioteca:"
		}
		authors, ok := post.GetParameters()["author"]
		if !ok || len(authors) == 0 {
			return
		}
		authorsString := strings.Join(authors, ", ")

		var bookcover string
		var bookrating string
		book := post.GetFirstParameterValue("book")
		isbn := post.GetFirstParameterValue("isbn")
		rating := post.GetFirstParameterValue("rating")

		if cover := post.GetFirstParameterValue("bookcover"); cover != "" {
			bookcover = "<p><img class=book-cover src=" + cover + "></p>"
		} else {
			bookcover = "<p><img class=book-cover src=https://covers.openlibrary.org/b/isbn/" + isbn + "-M.jpg></p>"
			p.fetchCover(isbn)
			p.app.SetPostParameter(post.GetPath(), "bookcover", []string{p.artpath})
		}

		if rating != "" {
			bookrating = "<img src=https://kandr3s.co/assets/icons/ratings/" + rating + ".png>"
		}
		doc.Find(".e-content").PrependHtml(fmt.Sprintf("<p>%s</p><div class='read-of h-cite book-details' value=%s>%s<div class=book-details-text><span class=book-title p-name><a target=_blank href=https://openlibrary.org/isbn/%s>%s</a></span><br /><span class=p-author book-author>%s</span><br />%s</div>", p.readStatus, post.GetFirstParameterValue("shelf"), bookcover, isbn, book, authorsString, bookrating))
	}
}

func (p *plugin) RenderSummaryForPost(rc plugintypes.RenderContext, post plugintypes.Post, doc *goquery.Document) {
	section := post.GetSection()
	// Reads Summaries
	if section == "reads" {
		doc.Find("h2.p-name").Remove() // Remove title
		switch post.GetFirstParameterValue("shelf") {
		case "read":
			p.readStatus = "Finished reading:"
		case "reading":
			p.readStatus = "Currently reading:"
			doc.Find("article").AddClass("hide")
		case "to-read":
			p.readStatus = "Want to read:"
		default:
			p.readStatus = "Added a book to La Biblioteca:"
		}
		authors, ok := post.GetParameters()["author"]
		if !ok || len(authors) == 0 {
			return
		}
		authorsString := strings.Join(authors, ", ")

		book := post.GetFirstParameterValue("book")
		cover := post.GetFirstParameterValue("bookcover")
		isbn := post.GetFirstParameterValue("isbn")
		rating := post.GetFirstParameterValue("rating")
		var bookcover string
		var bookrating string

		if cover != "" {
			bookcover = "<p><img class=book-cover src=" + cover + "></p>"
		}
		if isbn != "" {
			bookcover = "<p><img class=book-cover src=https://covers.openlibrary.org/b/isbn/" + isbn + "-M.jpg></p>"
			// p.fetchCover(isbn)
			// p.app.SetPostParameter(post.GetPath(), "bookcover", []string{p.artpath})
		}

		if rating != "" {
			bookrating = "<img src=https://kandr3s.co/assets/icons/ratings/" + rating + ".png>"
		}
		doc.Find(".e-content").PrependHtml(fmt.Sprintf("<p class=p-summary>%s</p><div class='read-of h-cite book-details' value=%s>%s<div class=book-details-text><span class=book-title p-name>%s</span><br /><span class=p-author book-author>%s</span><br />%s</div>", post.GetFirstParameterValue("title"), post.GetFirstParameterValue("shelf"), bookcover, book, authorsString, bookrating))
	}
}

func (p *plugin) PostCreated(post plugintypes.Post) {
	if post.GetSection() == "reads" {
		p.fetchReading()
	}
}

// Syndicate on Post Update
func (p *plugin) PostUpdated(post plugintypes.Post) {
	if post.GetSection() == "reads" {
		p.fetchReading()
	}
}

func (p *plugin) fetchReading() {
	query := `
        SELECT p.path,
        pp_shelf.value AS shelf,
        pp_book.value AS book,
        pp_bookcover.value AS bookcover,
        pp_author.value AS author
        FROM posts AS p
        LEFT JOIN post_parameters AS pp_shelf ON p.path = pp_shelf.path AND pp_shelf.parameter = 'shelf'
        LEFT JOIN post_parameters AS pp_book ON p.path = pp_book.path AND pp_book.parameter = 'book'
        LEFT JOIN post_parameters AS pp_bookcover ON p.path = pp_bookcover.path AND pp_bookcover.parameter = 'bookcover'
        LEFT JOIN post_parameters AS pp_author ON p.path = pp_author.path AND pp_author.parameter = 'author'
        WHERE p.section = 'reads' AND pp_shelf.value = 'reading'
        ORDER BY p.published DESC
        LIMIT 1;
    `

	row := p.app.GetDatabase().QueryRow(query)
	var reading Read
	err := row.Scan(&reading.Path, &reading.Shelf, &reading.Title, &reading.Cover, &reading.Author)
	if err != nil || reading.Path == "" {
		fmt.Println(fmt.Errorf("🔌 Biblioteca: Failed to fetch Currently Reading: %w", err))
		return
	}
	p.reading = reading
}

func (p *plugin) fetchCover(isbn string) {
	// Download Book Cover
	coverURL := fmt.Sprintf("https://covers.openlibrary.org/b/isbn/%s-M.jpg", isbn)

	resp, err := http.Get(coverURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Errorf("failed to fetch book art for ISBN %s", isbn)
		return
	}

	_, file := path.Split(coverURL)
	outputDir := "./static/images/art/books/"
	outputPath := path.Join(outputDir, file)

	response, err := http.Get(coverURL)
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
	fmt.Printf("🔌 Biblioteca: Book cover saved to: %s\n", outputPath)
	p.artpath = regexp.MustCompile(`(^|/)static(/|$)`).ReplaceAllString(outputPath, "/")
}
