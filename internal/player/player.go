package player

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type State int

const (
	Stopped State = iota
	Playing
	Paused
)

type Status struct {
	State    State
	Position time.Duration
	Duration time.Duration
	Title    string
	URL      string
}

type Player struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	sockPath string
	conn     net.Conn
	status   Status
	reqID    int
}

func New() *Player {
	return &Player{
		sockPath: filepath.Join(os.TempDir(), fmt.Sprintf("gocast-mpv-%d.sock", os.Getpid())),
	}
}

func (p *Player) Play(url, title string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop any existing playback
	p.stopLocked()

	// Clean up stale socket
	os.Remove(p.sockPath)

	p.cmd = exec.Command("mpv",
		"--no-video",
		"--really-quiet",
		"--no-terminal",
		"--input-ipc-server="+p.sockPath,
		url,
	)
	p.cmd.Stdout = nil
	p.cmd.Stderr = nil

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("start mpv: %w", err)
	}

	p.status = Status{
		State: Playing,
		Title: title,
		URL:   url,
	}

	// Connect to IPC socket with retries
	go p.connectIPC()

	return nil
}

func (p *Player) connectIPC() {
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		conn, err := net.Dial("unix", p.sockPath)
		if err == nil {
			p.mu.Lock()
			p.conn = conn
			p.mu.Unlock()
			return
		}
	}
}

func (p *Player) TogglePause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status.State == Stopped {
		return
	}

	p.sendCommand("cycle", "pause")

	if p.status.State == Playing {
		p.status.State = Paused
	} else {
		p.status.State = Playing
	}
}

func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()
}

func (p *Player) stopLocked() {
	if p.conn != nil {
		p.sendCommandLocked("quit")
		p.conn.Close()
		p.conn = nil
	}
	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- p.cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			p.cmd.Process.Kill()
		}
		p.cmd = nil
	}
	os.Remove(p.sockPath)
	p.status = Status{State: Stopped}
}

func (p *Player) Seek(seconds float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sendCommand("seek", seconds, "relative")
}

func (p *Player) SetVolume(delta float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sendCommand("add", "volume", delta)
}

func (p *Player) GetStatus() Status {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status.State == Stopped {
		return p.status
	}

	// Check if process is still running
	if p.cmd != nil && p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited() {
		p.status.State = Stopped
		return p.status
	}

	// Poll position and duration from mpv
	if pos, err := p.getPropertyFloat("time-pos"); err == nil {
		p.status.Position = time.Duration(pos * float64(time.Second))
	}
	if dur, err := p.getPropertyFloat("duration"); err == nil && dur > 0 {
		p.status.Duration = time.Duration(dur * float64(time.Second))
	}

	// Check pause state
	if paused, err := p.getPropertyBool("pause"); err == nil {
		if paused {
			p.status.State = Paused
		} else {
			p.status.State = Playing
		}
	}

	return p.status
}

func (p *Player) Cleanup() {
	p.Stop()
}

// --- IPC helpers ---

type mpvCommand struct {
	Command   []interface{} `json:"command"`
	RequestID int           `json:"request_id"`
}

type mpvResponse struct {
	Data      interface{} `json:"data"`
	RequestID int         `json:"request_id"`
	Error     string      `json:"error"`
}

func (p *Player) sendCommand(args ...interface{}) {
	p.sendCommandLocked(args...)
}

func (p *Player) sendCommandLocked(args ...interface{}) {
	if p.conn == nil {
		return
	}
	p.reqID++
	cmd := mpvCommand{
		Command:   args,
		RequestID: p.reqID,
	}
	data, _ := json.Marshal(cmd)
	data = append(data, '\n')
	p.conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	p.conn.Write(data)
}

func (p *Player) getPropertyFloat(name string) (float64, error) {
	if p.conn == nil {
		return 0, fmt.Errorf("no connection")
	}

	p.reqID++
	id := p.reqID
	cmd := mpvCommand{
		Command:   []interface{}{"get_property", name},
		RequestID: id,
	}
	data, _ := json.Marshal(cmd)
	data = append(data, '\n')

	p.conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	if _, err := p.conn.Write(data); err != nil {
		return 0, err
	}

	p.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 4096)
	n, err := p.conn.Read(buf)
	if err != nil {
		return 0, err
	}

	// Parse potentially multiple JSON lines, find our response
	lines := splitLines(buf[:n])
	for _, line := range lines {
		var resp mpvResponse
		if json.Unmarshal(line, &resp) == nil && resp.RequestID == id {
			if resp.Error != "" && resp.Error != "success" {
				return 0, fmt.Errorf("mpv: %s", resp.Error)
			}
			switch v := resp.Data.(type) {
			case float64:
				return v, nil
			}
			return 0, fmt.Errorf("unexpected type")
		}
	}
	return 0, fmt.Errorf("no response")
}

func (p *Player) getPropertyBool(name string) (bool, error) {
	if p.conn == nil {
		return false, fmt.Errorf("no connection")
	}

	p.reqID++
	id := p.reqID
	cmd := mpvCommand{
		Command:   []interface{}{"get_property", name},
		RequestID: id,
	}
	data, _ := json.Marshal(cmd)
	data = append(data, '\n')

	p.conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	if _, err := p.conn.Write(data); err != nil {
		return false, err
	}

	p.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 4096)
	n, err := p.conn.Read(buf)
	if err != nil {
		return false, err
	}

	lines := splitLines(buf[:n])
	for _, line := range lines {
		var resp mpvResponse
		if json.Unmarshal(line, &resp) == nil && resp.RequestID == id {
			if resp.Error != "" && resp.Error != "success" {
				return false, fmt.Errorf("mpv: %s", resp.Error)
			}
			switch v := resp.Data.(type) {
			case bool:
				return v, nil
			}
			return false, fmt.Errorf("unexpected type")
		}
	}
	return false, fmt.Errorf("no response")
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			if i > start {
				lines = append(lines, data[start:i])
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
