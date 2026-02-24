package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// StorageConfig contains storage backend configuration
type StorageConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Region    string `mapstructure:"region"`
}

// LoadStorageConfig loads storage configuration
func LoadStorageConfig() (*StorageConfig, error) {
	// Default config
	cfg := &StorageConfig{
		Endpoint:  "localhost:9000", // Local MinIO for testing
		AccessKey: "darkstorage",
		SecretKey: "darkstorage123",
		UseSSL:    false, // Local testing without SSL
		Region:    "us-east-1",
	}

	// Override with environment variables
	if endpoint := os.Getenv("DARKSTORAGE_ENDPOINT"); endpoint != "" {
		cfg.Endpoint = endpoint
	}
	if accessKey := os.Getenv("DARKSTORAGE_ACCESS_KEY"); accessKey != "" {
		cfg.AccessKey = accessKey
	}
	if secretKey := os.Getenv("DARKSTORAGE_SECRET_KEY"); secretKey != "" {
		cfg.SecretKey = secretKey
	}
	if sslStr := os.Getenv("DARKSTORAGE_USE_SSL"); sslStr == "true" {
		cfg.UseSSL = true
	}

	// Override with viper config (from ~/.darkstorage/config.yaml)
	if viper.IsSet("storage.endpoint") {
		cfg.Endpoint = viper.GetString("storage.endpoint")
	}
	if viper.IsSet("storage.access_key") {
		cfg.AccessKey = viper.GetString("storage.access_key")
	}
	if viper.IsSet("storage.secret_key") {
		cfg.SecretKey = viper.GetString("storage.secret_key")
	}
	if viper.IsSet("storage.use_ssl") {
		cfg.UseSSL = viper.GetBool("storage.use_ssl")
	}
	if viper.IsSet("storage.region") {
		cfg.Region = viper.GetString("storage.region")
	}

	// Validate
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("storage credentials not configured")
	}

	return cfg, nil
}
