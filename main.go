package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly/v2"
)

// fetchLyrics scrapes the lyrics from AZLyrics
func fetchLyrics(artist string, song string) (string, error) {
	// Create the URL based on the artist and song name
	url := fmt.Sprintf("https://www.azlyrics.com/lyrics/%s/%s.html", strings.ToLower(artist), strings.ToLower(song))

	// Create a new collector
	c := colly.NewCollector()

	// Variable to hold the scraped lyrics
	var lyrics string

	// This callback will be triggered when the HTML page is loaded
	c.OnHTML("div", func(e *colly.HTMLElement) {
		html, _ := e.DOM.Html()

		// Find the part of the HTML between the specified comment markers
		start := strings.Index(html, "<!-- Usage of azlyrics.com content by any third-party lyrics provider is prohibited by our licensing agreement. Sorry about that. -->")
		end := strings.Index(html, "<!-- MxM banner -->")

		if start != -1 && end != -1 && end > start {
			// Extract the section of the lyrics
			lyricsSection := html[start+len("<!-- Usage of azlyrics.com content by any third-party lyrics provider is prohibited by our licensing agreement. Sorry about that. -->") : end]
			lyrics = cleanUpLyrics(lyricsSection)
		}
	})

	// Error handling for when the scraping fails
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request failed:", err)
	})

	// Visit the URL
	err := c.Visit(url)
	if err != nil {
		return "", err
	}

	// Return the scraped lyrics
	return lyrics, nil
}

// cleanUpLyrics removes unwanted HTML tags and trims the string
func cleanUpLyrics(lyrics string) string {
	// Remove any HTML tags
	lyrics = strings.ReplaceAll(lyrics, "<br/>", "")
	lyrics = strings.ReplaceAll(lyrics, "</div>", "")
	lyrics = strings.ReplaceAll(lyrics, "<i>", "")
	lyrics = strings.ReplaceAll(lyrics, "</i>", "")
	lyrics = strings.ReplaceAll(lyrics, "&#39;", "'")
	lyrics = strings.TrimSpace(lyrics)

	return lyrics
}

func main() {
	artist := "Muse"
	song := "Supermassiveblackhole"

	lyrics, err := fetchLyrics(artist, song)
	if err != nil {
		log.Fatalf("Failed to fetch lyrics: %v", err)
	}

	lines := strings.Split(lyrics, "\n")

	if lyrics == "" {
		fmt.Println("Lyrics not found.")
	} else {
		for _, line := range lines {
			fmt.Println(line)
			fmt.Println("=====")
		}

		//fmt.Println("Lyrics:\n", lyrics)
	}
}
