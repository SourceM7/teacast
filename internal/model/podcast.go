package model

import "time"

type Podcast struct {
	ID          string
	Title       string
	Author      string
	Description string
	FeedURL     string
	ArtworkURL   string
	Episodes     []Episode
	EpisodeCount int
}

type Episode struct {
	ID          string
	PodcastID   string
	Title       string
	Description string
	Duration    time.Duration
	PublishedAt time.Time
	AudioURL    string
	FileSize    int64
	Played      bool
	Progress    time.Duration
}

type DownloadStatus int

const (
	NotDownloaded DownloadStatus = iota
	Downloading
	Downloaded
)

type Download struct {
	Episode  Episode
	Podcast  Podcast
	Status   DownloadStatus
	Progress float64
}

type Bookmark struct {
	Episode  Episode
	Podcast  Podcast
	AddedAt  time.Time
}

type HistoryEntry struct {
	Episode    Episode
	Podcast    Podcast
	PlayedAt   time.Time
	Progress   time.Duration
}
