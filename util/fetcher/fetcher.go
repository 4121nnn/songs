package fetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const geniusAPIBaseURL = "https://api.genius.com"

type GeniusResponse struct {
	Response struct {
		Hits []struct {
			Result struct {
				Title         string `json:"title"`
				PrimaryArtist struct {
					Name string `json:"name"`
				} `json:"primary_artist"`
				URL string `json:"url"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

func GetSongLyrics(accessToken, artistName, songTitle string) (string, error) {
	// Prepare the search query
	query := fmt.Sprintf("%s %s", artistName, songTitle)
	queryURL := fmt.Sprintf("%s/search?q=%s", geniusAPIBaseURL, url.QueryEscape(query))

	// Create the HTTP request
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse the response
	var geniusResp GeniusResponse
	if err := json.Unmarshal(body, &geniusResp); err != nil {
		return "", err
	}

	// Extract the song URL
	if len(geniusResp.Response.Hits) > 0 {
		songURL := geniusResp.Response.Hits[0].Result.URL
		return fetchLyricsFromURL(songURL)
	}

	return "", fmt.Errorf("no lyrics found for %s by %s", songTitle, artistName)
}

func fetchLyricsFromURL(songURL string) (string, error) {
	// Send a request to the song URL to fetch the lyrics
	resp, err := http.Get(songURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read and parse the response body (you may need to extract the lyrics from the HTML)
	// For simplicity, this example assumes that the lyrics are in a <div> with a specific class.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Extract lyrics (you may want to use a proper HTML parser)
	lyrics := extractLyrics(string(body))
	return lyrics, nil
}

func extractLyrics(html string) string {
	fmt.Println(html)
	// Simple string manipulation to extract lyrics from the HTML (implement as needed)
	// This is a placeholder. You should parse HTML properly (e.g., using GoQuery or other libraries).
	start := strings.Index(html, "<div class=\"lyrics\">")
	end := strings.Index(html, "</div>") // This needs to be adjusted based on actual HTML structure
	if start == -1 || end == -1 {
		return "Lyrics not found"
	}
	return html[start:end] // Simplified, for example purpose
}
