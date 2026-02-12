package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type DaemonConfig struct {
	Daemon        DaemonSettings       `yaml:"daemon"`
	SyncFolders   []SyncFolderConfig   `yaml:"sync_folders"`
	Notifications NotificationSettings `yaml:"notifications"`
	API           APIConfig            `yaml:"api"`
}

type DaemonSettings struct {
	Enabled            bool          `yaml:"enabled"`
	LogLevel           string        `yaml:"log_level"`
	LogFile            string        `yaml:"log_file"`
	PIDFile            string        `yaml:"pid_file"`
	IPCSocket          string        `yaml:"ipc_socket"`
	WorkerThreads      int           `yaml:"worker_threads"`
	MaxQueueSize       int           `yaml:"max_queue_size"`
	DebounceDelay      time.Duration `yaml:"debounce_delay"`
	Timeout            time.Duration `yaml:"timeout"`
	RetryAttempts      int           `yaml:"retry_attempts"`
	RetryDelay         time.Duration `yaml:"retry_delay"`
	BandwidthLimitUp   int           `yaml:"bandwidth_limit_up"`
	BandwidthLimitDown int           `yaml:"bandwidth_limit_down"`
}

type SyncFolderConfig struct {
	ID                 int      `yaml:"id"`
	Name               string   `yaml:"name"`
	LocalPath          string   `yaml:"local_path"`
	RemotePath         string   `yaml:"remote_path"`
	Direction          string   `yaml:"direction"`
	Enabled            bool     `yaml:"enabled"`
	Excludes           []string `yaml:"excludes"`
	ConflictResolution string   `yaml:"conflict_resolution"`
	SyncMode           string   `yaml:"sync_mode"`
	SyncInterval       int      `yaml:"sync_interval"`
	SyncSchedule       string   `yaml:"sync_schedule"`
	BandwidthLimit     int      `yaml:"bandwidth_limit"`
	MaxFileSize        int64    `yaml:"max_file_size"`
}

type NotificationSettings struct {
	Enabled       bool `yaml:"enabled"`
	ShowSuccess   bool `yaml:"show_success"`
	ShowErrors    bool `yaml:"show_errors"`
	ShowConflicts bool `yaml:"show_conflicts"`
}

type APIConfig struct {
	Endpoint string `yaml:"endpoint"`
}

func LoadDaemonConfig() (*DaemonConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".darkstorage")
	v := viper.New()
	v.SetConfigName("daemon")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	v.SetDefault("daemon.enabled", true)
	v.SetDefault("daemon.log_level", "info")
	v.SetDefault("daemon.log_file", filepath.Join(configPath, "daemon.log"))
	v.SetDefault("daemon.pid_file", filepath.Join(configPath, "daemon.pid"))
	v.SetDefault("daemon.ipc_socket", filepath.Join(configPath, "daemon.sock"))
	v.SetDefault("daemon.worker_threads", 4)
	v.SetDefault("daemon.max_queue_size", 1000)
	v.SetDefault("daemon.debounce_delay", "3s")
	v.SetDefault("daemon.timeout", "30s")
	v.SetDefault("daemon.retry_attempts", 3)
	v.SetDefault("daemon.retry_delay", "5s")
	v.SetDefault("notifications.enabled", true)
	v.SetDefault("notifications.show_success", false)
	v.SetDefault("notifications.show_errors", true)
	v.SetDefault("notifications.show_conflicts", true)
	v.SetDefault("api.endpoint", "https://api.darkstorage.io")

	config := &DaemonConfig{}
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}

func SaveDaemonConfig(config *DaemonConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".darkstorage")
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigName("daemon")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	v.Set("daemon", config.Daemon)
	v.Set("sync_folders", config.SyncFolders)
	v.Set("notifications", config.Notifications)
	v.Set("api", config.API)

	return v.WriteConfig()
}

func GetDefaultDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".darkstorage"), nil
}
