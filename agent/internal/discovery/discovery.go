package discovery

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"sysagent/internal/models"
)

type Discoverer struct{}

func NewDiscoverer() *Discoverer {
	return &Discoverer{}
}

func (d *Discoverer) Discover() []models.AppConfig {
	var apps []models.AppConfig

	// 1. Discover Systemd services (filtering some common ones? For now, we take a static prefix or all user ones)
	// Example: we can look for specific services. In a real scenario, this might be filtered by a config file.
	// For demonstration, we'll try to find a few common ones or parse basic running services.
	systemdApps := discoverSystemd()
	apps = append(apps, systemdApps...)

	// 2. Discover Docker containers
	dockerApps := discoverDocker()
	apps = append(apps, dockerApps...)

	return apps
}

func discoverSystemd() []models.AppConfig {
	var apps []models.AppConfig
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--state=running", "--no-pager", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		log.Println("Error discovering systemd:", err)
		return apps
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			unitName := fields[0]
			// Find PID for this unit
			pidCmd := exec.Command("systemctl", "show", "--property=MainPID", "--value", unitName)
			pidOut, pidErr := pidCmd.Output()
			if pidErr == nil {
				pidStr := strings.TrimSpace(string(pidOut))
				if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
					apps = append(apps, models.AppConfig{
						ID:   "systemd_" + unitName,
						Type: "systemd",
						Name: unitName,
						PID:  pid,
					})
				}
			}
		}
	}
	// Limit for demonstration to avoid overwhelmingly large numbers
	if len(apps) > 10 {
		apps = apps[:10]
	}
	return apps
}

type dockerContainer struct {
	Id    string   `json:"Id"`
	Names []string `json:"Names"`
}

func discoverDocker() []models.AppConfig {
	var apps []models.AppConfig

	// Fast way to hit docker socket without big dependencies
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	resp, err := client.Get("http://localhost/containers/json")
	if err != nil {
		// Docker might not be running or accessible
		return apps
	}
	defer resp.Body.Close()

	var containers []dockerContainer
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return apps
	}

	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		// To get PID we need to inspect the container
		resp2, err := client.Get("http://localhost/containers/" + c.Id + "/json")
		if err == nil {
			var details struct {
				State struct {
					Pid int `json:"Pid"`
				} `json:"State"`
			}
			json.NewDecoder(resp2.Body).Decode(&details)
			resp2.Body.Close()

			if details.State.Pid > 0 {
				apps = append(apps, models.AppConfig{
					ID:   "docker_" + name,
					Type: "docker",
					Name: name,
					PID:  details.State.Pid,
				})
			}
		}
	}

	return apps
}
