package controller

import (
	"log"
	"os/exec"
	"strings"

	"sysagent/internal/models"
)

func StartCommandListener(commandCh <-chan models.ControlCommand) {
	go func() {
		for cmd := range commandCh {
			log.Printf("Executing command %s for App %s", cmd.Command, cmd.AppID)

			// Extract original type and name from AppID (e.g. systemd_nginx -> type=systemd, name=nginx)
			parts := strings.SplitN(cmd.AppID, "_", 2)
			if len(parts) != 2 {
				log.Println("Invalid app_id format")
				continue
			}

			appType := parts[0]
			appName := parts[1]

			if appType == "systemd" {
				handleSystemdCommand(cmd.Command, appName)
			} else if appType == "docker" {
				handleDockerCommand(cmd.Command, appName)
			}
		}
	}()
}

func handleSystemdCommand(command string, unitName string) {
	action := ""
	switch command {
	case "START_SERVICE":
		action = "start"
	case "STOP_SERVICE":
		action = "stop"
	case "RESTART_SERVICE":
		action = "restart"
	default:
		return
	}

	cmd := exec.Command("systemctl", action, unitName)
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to %s systemd service %s: %v", action, unitName, err)
	}
}

func handleDockerCommand(command string, containerName string) {
	action := ""
	switch command {
	case "START_SERVICE":
		action = "start"
	case "STOP_SERVICE":
		action = "stop"
	case "RESTART_SERVICE":
		action = "restart"
	default:
		return
	}

	cmd := exec.Command("docker", action, containerName)
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to %s docker container %s: %v", action, containerName, err)
	}
}
