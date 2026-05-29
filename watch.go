package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/phamphu232/go-watch-git/config"
)

type State struct {
	changeFileCount int
}

var currentState = map[string]State{}

func checkChanges(folder string) []string {
	changes := []string{}
	cmd := exec.Command("git", "-C", folder, "status", "--porcelain")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return []string{fmt.Sprintf("Error: %v | Git Details: %s", err, strings.TrimSpace(string(output)))}
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return changes
	}

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if line != "" {
			changes = append(changes, line)
		}
	}

	return changes
}

func checkSourceCode() {
	for _, folder := range config.GetConfig().WatchFolders {
		if folder == "" {
			continue
		}

		state := currentState[folder]

		changes := checkChanges(folder)
		changeCount := len(changes)
		shouldPushNotify := false

		if state.changeFileCount != changeCount || (config.GetConfig().AlwaysNotifyOnChange && changeCount > 0) {
			shouldPushNotify = true
			state.changeFileCount = changeCount
		}

		if shouldPushNotify {
			if changeCount > 0 {
				changeStr := strings.Join(changes, "\n")
				log.Printf("Changes detected in: %s\n%s", folder, changeStr)
				runes := []rune(changeStr)
				limitChar := 3500
				if len(runes) > limitChar {
					changeStr = "```" + string(runes[:limitChar]) + "\n...```"
				}
				for _, webhook := range config.GetConfig().GoogleWebhooks {
					if webhook == "" {
						log.Printf("Webhook URL is empty, skip")
						continue
					}
					log.Printf("Post to Google Chat")
					message := fmt.Sprintf("%s\n⚠️There is/are %d modified file(s) at: %s:%s\n```%s```", config.GetConfig().MesageTemplate, changeCount, config.GetConfig().ServerAddress, folder, changeStr)
					PostToGoogleChat(message, webhook)
				}
			} else {
				for _, webhook := range config.GetConfig().GoogleWebhooks {
					if webhook == "" {
						log.Printf("Webhook URL is empty, skip")
						continue
					}
					log.Printf("Post to Google Chat")
					message := fmt.Sprintf("%s\n✅There is/are %d modified file(s) at: %s:%s", config.GetConfig().MesageTemplate, 0, config.GetConfig().ServerAddress, folder)
					PostToGoogleChat(message, webhook)
				}
			}

			currentState[folder] = state
		}
	}
}
