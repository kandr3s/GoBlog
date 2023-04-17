package customUIElements

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.goblog.app/app/pkgs/bufferpool"
	"go.goblog.app/app/pkgs/htmlbuilder"
	"go.goblog.app/app/pkgs/plugintypes"
)

func GetPlugin() (plugintypes.SetConfig, plugintypes.UI2, plugintypes.UISummary) {
	p := &plugin{}
	return p, p, p
}

type plugin struct {
	config map[string]any
}

func (p *plugin) SetConfig(config map[string]any) {
	p.config = config
}

// UISummary - Timeline Theme
func (p *plugin) RenderSummaryForPost(rc plugintypes.RenderContext, post plugintypes.Post, doc *goquery.Document) {

	doc.Find(".h-entry").Each(func(i int, s *goquery.Selection) {
		section := doc.Find(".post-section").Text()
		if len(section) == 0 {
			return
		}
		firstChar := string([]rune(section)[0])
		link := doc.Find("a.permalink").AttrOr("href", "")
		hidden := doc.Find("p.hide")
		date := doc.Find(".post-footer time").Text()
		postfooter := doc.Find(".post-footer")
		category := strings.ToTitle(section)
		visibility := doc.Find(".visibility")

		doc.Find("article").PrependHtml(fmt.Sprintf("<span class=\"category\"><a href=\"/%s\" title=\"%v\">%s</a></span>", post.GetSection(), section, firstChar))
		doc.Find("article").AppendHtml(fmt.Sprintf("%s<a href=%s class=\"permalink u-url\">%s</a>", visibility.Text(), link, date))
		postfooter.Remove()
		hidden.Remove()
	})
}

func (p *plugin) RenderWithDocument(rc plugintypes.RenderContext, doc *goquery.Document) {
	blog := rc.GetBlog()
	if blog == "" {
		fmt.Println("Custom UI Elements plugin: blog is empty!")
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
					if name == "microsub" {
						buf.Reset()
						link := element["link"]
						hb.WriteElementOpen("link", "rel", "microsub", "href", link)
						doc.Find("head").AppendHtml(buf.String())
					}
					if name == "manifest" {
						buf.Reset()
						hb.WriteElementOpen("link", "rel", "manifest", "href", "/manifest.json")
						doc.Find("head").AppendHtml(buf.String())
					}
					if name == "socialicons" {
						buf.Reset()
						hb.WriteElementOpen("div", "class", "social-icons")
						hb.WriteElementOpen("a", "class", "social-link", "href", "/feeds", "title", "Subscribe")
						hb.WriteElementOpen("svg", "width", "16px", "height", "16px", "xmlns", "http://www.w3.org/2000/svg", "viewBox", "0 0 448 512", "style", "filter: invert(53%) sepia(20%) saturate(2929%) hue-rotate(352deg) brightness(94%) contrast(93%);")
						hb.WriteElementOpen("path", "d", "M25.57 176.1C12.41 175.4 .9117 185.2 .0523 198.4s9.173 24.65 22.39 25.5c120.1 7.875 225.7 112.7 233.6 233.6C256.9 470.3 267.4 480 279.1 480c.5313 0 1.062-.0313 1.594-.0625c13.22-.8438 23.25-12.28 22.39-25.5C294.6 310.3 169.7 185.4 25.57 176.1zM32 32C14.33 32 0 46.31 0 64s14.33 32 32 32c194.1 0 352 157.9 352 352c0 17.69 14.33 32 32 32s32-14.31 32-32C448 218.6 261.4 32 32 32zM63.1 351.9C28.63 351.9 0 380.6 0 416s28.63 64 63.1 64s64.08-28.62 64.08-64S99.37 351.9 63.1 351.9z")
						hb.WriteElementClose("svg")
						hb.WriteElementClose("a")
						// hb.WriteElementOpen("a", "class", "social-link", "href", "/@kandr3s", "title", "Mastodon")
						// hb.WriteElementOpen("svg", "width", "16px", "height", "16px", "xmlns", "http://www.w3.org/2000/svg", "viewBox", "0 0 448 512", "style", "filter: invert(44%) sepia(99%) saturate(488%) hue-rotate(168deg) brightness(89%) contrast(85%);")
						// hb.WriteElementOpen("path", "d", "M433 179.11c0-97.2-63.71-125.7-63.71-125.7-62.52-28.7-228.56-28.4-290.48 0 0 0-63.72 28.5-63.72 125.7 0 115.7-6.6 259.4 105.63 289.1 40.51 10.7 75.32 13 103.33 11.4 50.81-2.8 79.32-18.1 79.32-18.1l-1.7-36.9s-36.31 11.4-77.12 10.1c-40.41-1.4-83-4.4-89.63-54a102.54 102.54 0 0 1-.9-13.9c85.63 20.9 158.65 9.1 178.75 6.7 56.12-6.7 105-41.3 111.23-72.9 9.8-49.8 9-121.5 9-121.5zm-75.12 125.2h-46.63v-114.2c0-49.7-64-51.6-64 6.9v62.5h-46.33V197c0-58.5-64-56.6-64-6.9v114.2H90.19c0-122.1-5.2-147.9 18.41-175 25.9-28.9 79.82-30.8 103.83 6.1l11.6 19.5 11.6-19.5c24.11-37.1 78.12-34.8 103.83-6.1 23.71 27.3 18.4 53 18.4 175z")
						// hb.WriteElementClose("svg")
						// hb.WriteElementClose("a")
						hb.WriteElementOpen("a", "class", "social-link", "rel", "nofollow", "target", "_blank", "href", "https://micro.blog/kandr3s?remote_follow=1", "title", "Micro.blog")
						hb.WriteElementOpen("svg", "width", "16px", "height", "16px", "xmlns", "http://www.w3.org/2000/svg", "viewBox", "0 0 448 512", "style", "filter: invert(81%) sepia(50%) saturate(5948%) hue-rotate(0deg) brightness(105%) contrast(103%);")
						hb.WriteElementOpen("path", "d", "M399.36,362.23c29.49-34.69,47.1-78.34,47.1-125.79C446.46,123.49,346.86,32,224,32S1.54,123.49,1.54,236.44,101.14,440.87,224,440.87a239.28,239.28,0,0,0,79.44-13.44,7.18,7.18,0,0,1,8.12,2.56c18.58,25.09,47.61,42.74,79.89,49.92a4.42,4.42,0,0,0,5.22-3.43,4.37,4.37,0,0,0-.85-3.62,87,87,0,0,1,3.69-110.69ZM329.52,212.4l-57.3,43.49L293,324.75a6.5,6.5,0,0,1-9.94,7.22L224,290.92,164.94,332a6.51,6.51,0,0,1-9.95-7.22l20.79-68.86-57.3-43.49a6.5,6.5,0,0,1,3.8-11.68l71.88-1.51,23.66-67.92a6.5,6.5,0,0,1,12.28,0l23.66,67.92,71.88,1.51a6.5,6.5,0,0,1,3.88,11.68Z")
						hb.WriteElementClose("svg")
						hb.WriteElementClose("a")
						hb.WriteElementOpen("a", "class", "social-link", "rel", "nofollow", "target", "_blank", "href", "/sonemic")
						hb.WriteElementOpen("img", "src", "/assets/icons/sonemic.svg", "title", "RYM/Sonemic", "width", "16px", "height", "16px")
						hb.WriteElementClose("a")
						hb.WriteElementOpen("a", "class", "social-link", "rel", "nofollow", "target", "_blank", "href", "/letterboxd")
						hb.WriteElementOpen("img", "src", "/assets/icons/letterboxd.svg", "title", "Letterboxd")
						hb.WriteElementClose("a")
						hb.WriteElementOpen("a", "class", "social-link", "rel", "nofollow", "target", "_blank", "href", "/github", "title", "GitHub")
						hb.WriteElementOpen("svg", "width", "16px", "height", "16px", "xmlns", "http://www.w3.org/2000/svg", "viewBox", "0 0 448 512", "style", "filter: invert(12%) sepia(7%) saturate(1293%) hue-rotate(173deg) brightness(100%) contrast(90%); background-color: white;border-radius:50%")
						hb.WriteElementOpen("path", "d", "M165.9 397.4c0 2-2.3 3.6-5.2 3.6-3.3.3-5.6-1.3-5.6-3.6 0-2 2.3-3.6 5.2-3.6 3-.3 5.6 1.3 5.6 3.6zm-31.1-4.5c-.7 2 1.3 4.3 4.3 4.9 2.6 1 5.6 0 6.2-2s-1.3-4.3-4.3-5.2c-2.6-.7-5.5.3-6.2 2.3zm44.2-1.7c-2.9.7-4.9 2.6-4.6 4.9.3 2 2.9 3.3 5.9 2.6 2.9-.7 4.9-2.6 4.6-4.6-.3-1.9-3-3.2-5.9-2.9zM244.8 8C106.1 8 0 113.3 0 252c0 110.9 69.8 205.8 169.5 239.2 12.8 2.3 17.3-5.6 17.3-12.1 0-6.2-.3-40.4-.3-61.4 0 0-70 15-84.7-29.8 0 0-11.4-29.1-27.8-36.6 0 0-22.9-15.7 1.6-15.4 0 0 24.9 2 38.6 25.8 21.9 38.6 58.6 27.5 72.9 20.9 2.3-16 8.8-27.1 16-33.7-55.9-6.2-112.3-14.3-112.3-110.5 0-27.5 7.6-41.3 23.6-58.9-2.6-6.5-11.1-33.3 2.6-67.9 20.9-6.5 69 27 69 27 20-5.6 41.5-8.5 62.8-8.5s42.8 2.9 62.8 8.5c0 0 48.1-33.6 69-27 13.7 34.7 5.2 61.4 2.6 67.9 16 17.7 25.8 31.5 25.8 58.9 0 96.5-58.9 104.2-114.8 110.5 9.2 7.9 17 22.9 17 46.4 0 33.7-.3 75.4-.3 83.6 0 6.5 4.6 14.4 17.3 12.1C428.2 457.8 496 362.9 496 252 496 113.3 383.5 8 244.8 8zM97.2 352.9c-1.3 1-1 3.3.7 5.2 1.6 1.6 3.9 2.3 5.2 1 1.3-1 1-3.3-.7-5.2-1.6-1.6-3.9-2.3-5.2-1zm-10.8-8.1c-.7 1.3.3 2.9 2.3 3.9 1.6 1 3.6.7 4.3-.7.7-1.3-.3-2.9-2.3-3.9-2-.6-3.6-.3-4.3.7zm32.4 35.6c-1.6 1.3-1 4.3 1.3 6.2 2.3 2.3 5.2 2.6 6.5 1 1.3-1.3.7-4.3-1.3-6.2-2.2-2.3-5.2-2.6-6.5-1zm-11.4-14.7c-1.6 1-1.6 3.6 0 5.9 1.6 2.3 4.3 3.3 5.6 2.3 1.6-1.3 1.6-3.9 0-6.2-1.4-2.3-4-3.3-5.6-2z")
						hb.WriteElementClose("svg")
						hb.WriteElementClose("a")
						// hb.WriteElementOpen("a", "class", "social-link", "rel", "nofollow", "target", "_blank", "href", "/spotify", "title", "Spotify")
						// hb.WriteElementOpen("svg", "width", "16px", "height", "16px", "xmlns", "http://www.w3.org/2000/svg", "viewBox", "0 0 448 512", "style", "filter: invert(87%) sepia(14%) saturate(5946%) hue-rotate(75deg) brightness(89%) contrast(89%);")
						// hb.WriteElementOpen("path", "d", "M248 8C111.1 8 0 119.1 0 256s111.1 248 248 248 248-111.1 248-248S384.9 8 248 8zm100.7 364.9c-4.2 0-6.8-1.3-10.7-3.6-62.4-37.6-135-39.2-206.7-24.5-3.9 1-9 2.6-11.9 2.6-9.7 0-15.8-7.7-15.8-15.8 0-10.3 6.1-15.2 13.6-16.8 81.9-18.1 165.6-16.5 237 26.2 6.1 3.9 9.7 7.4 9.7 16.5s-7.1 15.4-15.2 15.4zm26.9-65.6c-5.2 0-8.7-2.3-12.3-4.2-62.5-37-155.7-51.9-238.6-29.4-4.8 1.3-7.4 2.6-11.9 2.6-10.7 0-19.4-8.7-19.4-19.4s5.2-17.8 15.5-20.7c27.8-7.8 56.2-13.6 97.8-13.6 64.9 0 127.6 16.1 177 45.5 8.1 4.8 11.3 11 11.3 19.7-.1 10.8-8.5 19.5-19.4 19.5zm31-76.2c-5.2 0-8.4-1.3-12.9-3.9-71.2-42.5-198.5-52.7-280.9-29.7-3.6 1-8.1 2.6-12.9 2.6-13.2 0-23.3-10.3-23.3-23.6 0-13.6 8.4-21.3 17.4-23.9 35.2-10.3 74.6-15.2 117.5-15.2 73 0 149.5 15.2 205.4 47.8 7.8 4.5 12.9 10.7 12.9 22.6 0 13.6-11 23.3-23.2 23.3z")
						// hb.WriteElementClose("svg")
						// hb.WriteElementClose("a")
						hb.WriteElementClose("div")
						doc.Find("header p:last-of-type").AfterHtml(buf.String())
					}
					if name == "indiewebicons" {
						buf.Reset()
						hb.WriteElementOpen("p")
						hb.WriteElementClose("a")
						hb.WriteElementOpen("a", "target", "_blank", "href", "https://goblog.app/")
						hb.WriteElementOpen("img", "src", "/images/goblog.png", "title", "Powered by GoBlog")
						hb.WriteElementClose("a")
						hb.WriteElementOpen("a", "target", "_blank", "href", "https://indieweb.org/Webmention")
						hb.WriteElementOpen("img", "src", "/images/webmention_button.png", "title", "Webmentions accepted ✅")
						hb.WriteElementClose("a")
						hb.WriteElementOpen("a", "target", "_blank", "href", "https://microformats.org")
						hb.WriteElementOpen("img", "src", "/images/microformats_button.png", "title", "Microformats support")
						hb.WriteElementClose("a")
						hb.WriteElementOpen("a", "target", "_blank", "href", "https://indieweb.org/")
						hb.WriteElementOpen("img", "src", "/images/indieweb_button.png", "title", "Part of the IndieWeb")
						hb.WriteElementClose("p")
						doc.Find("footer").AppendHtml(buf.String())
					}
				}
			}

		}
	}
}
