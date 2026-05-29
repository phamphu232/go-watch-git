package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	ServerAddress    string   `json:"server_address"`     // Server address
	WatchFolders     []string `json:"watch_folders"`      // Folders to watch (comma separated)
	GoogleWebhooks   []string `json:"google_webhooks"`    // Google Chat webhook URLs(comma separated)
	MesageTemplate   string   `json:"message_template"`   // Message template
	Interval         int      `json:"interval"`           // Interval (seconds)
	LogRetentionDays int      `json:"log_retention_days"` // Log retention days
}

var (
	AppConfig   Config
	configLock  sync.RWMutex
	lastModTime time.Time
)

func configFilePath() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	baseDir := filepath.Dir(exePath)
	return filepath.Join(baseDir, "config.json")
}

func GetConfig() Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return AppConfig
}

func Load() {
	path := configFilePath()

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		configLock.Lock()
		AppConfig = Config{
			ServerAddress:    "",
			WatchFolders:     []string{""},
			GoogleWebhooks:   []string{""},
			MesageTemplate:   "Warning",
			Interval:         30,
			LogRetentionDays: 14,
		}
		configLock.Unlock()

		data, _ := json.MarshalIndent(AppConfig, "", "    ")
		_ = os.WriteFile(path, data, 0666)
		os.Chmod(path, 0666)

		if newInfo, err := os.Stat(path); err == nil {
			lastModTime = newInfo.ModTime()
		}
		return
	}

	file, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading config: %v", err)
		return
	}

	var tempConfig Config
	if err := json.Unmarshal(file, &tempConfig); err != nil {
		log.Printf("Invalid JSON in config file: %v. Keeping old config.", err)
		return
	}

	configLock.Lock()
	AppConfig = tempConfig
	lastModTime = info.ModTime()
	configLock.Unlock()

	log.Println("Config loaded successfully")
}

func WatchConfig(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			path := configFilePath()
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			if info.ModTime().After(lastModTime) {
				log.Println("Detected config change, reloading...")
				Load()
			}
		}
	}()
}
