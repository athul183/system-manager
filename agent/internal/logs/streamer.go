package logs

import (
	"bufio"
	"log"
	"os/exec"
	"strings"
	"sync"
)

type LogMessage struct {
	AppID   string `json:"app_id"`
	Line    string `json:"line"`
	IsError bool   `json:"is_error"`
}

type Streamer struct {
	activeStreams map[string]*exec.Cmd
	streamMu      sync.Mutex
	logChan       chan<- LogMessage
}

func NewStreamer(logChan chan<- LogMessage) *Streamer {
	return &Streamer{
		activeStreams: make(map[string]*exec.Cmd),
		logChan:       logChan,
	}
}

func (s *Streamer) StartStream(appID string) {
	s.streamMu.Lock()
	defer s.streamMu.Unlock()

	if _, exists := s.activeStreams[appID]; exists {
		// Already streaming
		return
	}

	parts := strings.SplitN(appID, "_", 2)
	if len(parts) != 2 {
		return
	}
	appType, appName := parts[0], parts[1]

	var cmd *exec.Cmd
	if appType == "systemd" {
		cmd = exec.Command("journalctl", "-u", appName, "-f", "-n", "100")
	} else if appType == "docker" {
		cmd = exec.Command("docker", "logs", "-f", "--tail", "100", appName)
	} else {
		return
	}

	s.activeStreams[appID] = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start logs for %s: %v", appID, err)
		delete(s.activeStreams, appID)
		return
	}

	// Read stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			s.logChan <- LogMessage{AppID: appID, Line: scanner.Text(), IsError: false}
		}
	}()

	// Read stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			s.logChan <- LogMessage{AppID: appID, Line: scanner.Text(), IsError: true}
		}

		// Clean up on exit
		s.streamMu.Lock()
		delete(s.activeStreams, appID)
		s.streamMu.Unlock()
	}()
}

func (s *Streamer) StopStream(appID string) {
	s.streamMu.Lock()
	defer s.streamMu.Unlock()

	if cmd, exists := s.activeStreams[appID]; exists {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		delete(s.activeStreams, appID)
	}
}
