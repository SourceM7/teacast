# teacast

A terminal podcast player built with Go and Bubbletea.

## Features

- Browse iTunes top podcasts with featured card
- Search podcasts via iTunes API
- Stream episodes directly through mpv
- Library — subscribe and manage your podcasts
- Bookmarks — save and remove episodes for later
- Listening history
- Volume control
- Dual-pane podcast detail view with episode list
- Keyboard-driven navigation with context-sensitive help

## Requirements

- Go 1.24+
- [mpv](https://mpv.io/) — for audio playback

```bash
# Fedora
sudo dnf install mpv

# Ubuntu/Debian
sudo apt install mpv

# Arch
sudo pacman -S mpv
```

## Install & Run

```bash
git clone https://github.com/SourceM7/teacast.git
cd teacast
go build -o teacast ./cmd/main.go
./teacast
```

## Keybindings

| Key | Action |
|-----|--------|
| `1-6` / `Tab` | Switch tabs |
| `↑/↓` or `j/k` | Navigate |
| `g / G` | First / last item |
| `Enter` | Open podcast / play episode |
| `s` | Subscribe |
| `u` | Unsubscribe (Library tab) |
| `b` | Bookmark episode |
| `d` | Remove bookmark (Bookmarks tab) |
| `/` | Search |
| `Space` | Play / pause |
| `+ / -` | Volume up / down |
| `← / →` | Seek ±10s |
| `Esc` | Back |
| `q` | Quit |
