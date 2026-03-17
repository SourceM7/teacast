package api

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gocast/internal/model"
)

const (
	itunesSearchURL = "https://itunes.apple.com/search"
	itunesTopURL    = "https://itunes.apple.com/us/rss/toppodcasts/limit=20/json"
	httpTimeout     = 15 * time.Second
)

var client = &http.Client{Timeout: httpTimeout}

// --- iTunes Search API ---

type itunesSearchResponse struct {
	ResultCount int             `json:"resultCount"`
	Results     []itunesResult  `json:"results"`
}

type itunesResult struct {
	CollectionID   int    `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	ArtistName     string `json:"artistName"`
	FeedURL        string `json:"feedUrl"`
	ArtworkURL     string `json:"artworkUrl100"`
	TrackCount     int    `json:"trackCount"`
	GenreNames     []string `json:"genres"`
}

func SearchPodcasts(query string) ([]model.Podcast, error) {
	u, _ := url.Parse(itunesSearchURL)
	q := u.Query()
	q.Set("term", query)
	q.Set("media", "podcast")
	q.Set("limit", "20")
	u.RawQuery = q.Encode()

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("itunes search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result itunesSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	var podcasts []model.Podcast
	for _, r := range result.Results {
		podcasts = append(podcasts, model.Podcast{
			ID:          fmt.Sprintf("%d", r.CollectionID),
			Title:       r.CollectionName,
			Author:      r.ArtistName,
			FeedURL:     r.FeedURL,
			ArtworkURL:  r.ArtworkURL,
			EpisodeCount: r.TrackCount,
		})
	}
	return podcasts, nil
}

// --- Top Podcasts (iTunes RSS Feed) ---

type topFeedResponse struct {
	Feed struct {
		Entry []topFeedEntry `json:"entry"`
	} `json:"feed"`
}

type topFeedEntry struct {
	Name   jsonLabel `json:"im:name"`
	Artist jsonLabel `json:"im:artist"`
	ID     struct {
		Attrs struct {
			ID string `json:"im:id"`
		} `json:"attributes"`
	} `json:"id"`
	Summary jsonLabel `json:"summary"`
}

type jsonLabel struct {
	Label string `json:"label"`
}

func FetchTopPodcasts() ([]model.Podcast, error) {
	resp, err := client.Get(itunesTopURL)
	if err != nil {
		return nil, fmt.Errorf("top podcasts: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result topFeedResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	var podcasts []model.Podcast
	for _, r := range result.Feed.Entry {
		podcasts = append(podcasts, model.Podcast{
			ID:          r.ID.Attrs.ID,
			Title:       r.Name.Label,
			Author:      r.Artist.Label,
			Description: r.Summary.Label,
		})
	}

	// Lookup feed URLs via iTunes lookup API
	if len(podcasts) > 0 {
		ids := make([]string, 0, len(podcasts))
		for _, p := range podcasts {
			if p.ID != "" {
				ids = append(ids, p.ID)
			}
		}
		feeds := lookupFeedURLs(ids)
		for i := range podcasts {
			if feedURL, ok := feeds[podcasts[i].ID]; ok {
				podcasts[i].FeedURL = feedURL
			}
		}
	}

	return podcasts, nil
}

func lookupFeedURLs(ids []string) map[string]string {
	u, _ := url.Parse("https://itunes.apple.com/lookup")
	q := u.Query()
	q.Set("id", strings.Join(ids, ","))
	q.Set("entity", "podcast")
	u.RawQuery = q.Encode()

	resp, err := client.Get(u.String())
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result itunesSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	feeds := make(map[string]string)
	for _, r := range result.Results {
		if r.FeedURL != "" {
			feeds[fmt.Sprintf("%d", r.CollectionID)] = r.FeedURL
		}
	}
	return feeds
}

// --- RSS Feed Parsing ---

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	PubDate     string       `xml:"pubDate"`
	Enclosure   rssEnclosure `xml:"enclosure"`
	Duration    string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd duration"`
	Summary     string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd summary"`
	GUID        string       `xml:"guid"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func FetchEpisodes(feedURL string) ([]model.Episode, error) {
	if feedURL == "" {
		return nil, fmt.Errorf("no feed URL")
	}

	resp, err := client.Get(feedURL)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read feed: %w", err)
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}

	limit := 25
	if len(feed.Channel.Items) < limit {
		limit = len(feed.Channel.Items)
	}

	var episodes []model.Episode
	for i, item := range feed.Channel.Items[:limit] {
		ep := model.Episode{
			ID:          item.GUID,
			Title:       cleanString(item.Title),
			Description: cleanString(firstN(coalesce(item.Summary, item.Description), 200)),
			Duration:    parseDuration(item.Duration),
			PublishedAt: parseDate(item.PubDate),
			AudioURL:    item.Enclosure.URL,
			FileSize:    item.Enclosure.Length,
		}
		if ep.ID == "" {
			ep.ID = fmt.Sprintf("ep-%d", i)
		}
		episodes = append(episodes, ep)
	}

	return episodes, nil
}

func parseDuration(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	parts := strings.Split(s, ":")
	switch len(parts) {
	case 3: // HH:MM:SS
		h := parseInt(parts[0])
		m := parseInt(parts[1])
		sec := parseInt(parts[2])
		return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second
	case 2: // MM:SS
		m := parseInt(parts[0])
		sec := parseInt(parts[1])
		return time.Duration(m)*time.Minute + time.Duration(sec)*time.Second
	case 1: // seconds only
		sec := parseInt(parts[0])
		return time.Duration(sec) * time.Second
	}
	return 0
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

func parseDate(s string) time.Time {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, strings.TrimSpace(s)); err == nil {
			return t
		}
	}
	return time.Time{}
}

func cleanString(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.TrimSpace(s)
	// Strip basic HTML tags
	for {
		start := strings.Index(s, "<")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], ">")
		if end == -1 {
			break
		}
		s = s[:start] + s[start+end+1:]
	}
	return strings.TrimSpace(s)
}

func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
