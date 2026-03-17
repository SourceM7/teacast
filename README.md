# teacast

A terminal podcast player built with Go and Bubbletea.

## Features

- Browse iTunes top podcasts
- Search podcasts via iTunes API
- Stream episodes directly through mpv
- Library — subscribe and manage your podcasts
- Bookmarks — save episodes for later
- Listening history
- Keyboard-driven navigation

## Requirements

- Go 1.21+
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
| `b` | Bookmark episode |
| `/` | Search |
| `Space` | Play / pause |
| `← / →` | Seek ±10s |
| `Esc` | Back |
| `q` | Quit |
