package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"sysagent/internal/models"
)

type Sampler struct {
	apps      []models.AppConfig
	outChan   chan<- models.AppMetrics
	stopChan  chan struct{}
	wg        sync.WaitGroup
	lastUsage map[int]cpuUsageData
}

type cpuUsageData struct {
	utime  uint64
	stime  uint64
	uptime float64 // system uptime
}

func NewSampler(apps []models.AppConfig, outChan chan<- models.AppMetrics) *Sampler {
	return &Sampler{
		apps:      apps,
		outChan:   outChan,
		stopChan:  make(chan struct{}),
		lastUsage: make(map[int]cpuUsageData),
	}
}

func (s *Sampler) Start() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.collectAll()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Sampler) Stop() {
	close(s.stopChan)
}

func (s *Sampler) collectAll() {
	for _, app := range s.apps {
		s.wg.Add(1)
		go func(a models.AppConfig) {
			defer s.wg.Done()

			cpu, memMB, ok := s.gatherProcStats(a.PID)
			status := "running"
			if !ok {
				status = "stopped"
			}

			s.outChan <- models.AppMetrics{
				AppID:     a.ID,
				CPU:       cpu,
				Memory:    memMB,
				Status:    status,
				Timestamp: time.Now().Unix(),
			}
		}(app)
	}
	s.wg.Wait()
}

// gatherProcStats reads from /proc/[pid]/stat and /proc/[pid]/status
// This avoids heavy external commands like 'top' and keeps CPU usage minimal.
func (s *Sampler) gatherProcStats(pid int) (float64, int, bool) {
	// 1. Read Uptime
	uptimeData, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, 0, false
	}
	uptimeFields := strings.Fields(string(uptimeData))
	if len(uptimeFields) < 1 {
		return 0, 0, false
	}
	uptime, _ := strconv.ParseFloat(uptimeFields[0], 64)

	// 2. Read App Stat
	statData, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, 0, false
	}
	statFields := strings.Fields(string(statData))
	if len(statFields) < 15 {
		return 0, 0, false
	}

	utime, _ := strconv.ParseUint(statFields[13], 10, 64)
	stime, _ := strconv.ParseUint(statFields[14], 10, 64)

	// Calculate CPU %
	// HZ in Linux is usually 100
	const HZ = 100.0
	var cpuPercent float64 = 0.0

	last, hasLast := s.lastUsage[pid]
	if hasLast {
		totalTicks := float64(utime+stime) - float64(last.utime+last.stime)
		uptimeDiff := uptime - last.uptime
		if uptimeDiff > 0 {
			cpuPercent = (totalTicks / HZ) / uptimeDiff * 100.0
		}
	}
	s.lastUsage[pid] = cpuUsageData{utime: utime, stime: stime, uptime: uptime}

	// 3. Read App Status for Memory (VmRSS)
	memBytes := 0
	statusFile, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err == nil {
		scanner := bufio.NewScanner(statusFile)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "VmRSS:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					kb, _ := strconv.Atoi(fields[1])
					memBytes = kb * 1024
				}
				break
			}
		}
		statusFile.Close()
	}

	return cpuPercent, memBytes / (1024 * 1024), true
}
