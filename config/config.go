package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func ReadConfig(appID string) {
	// Set default values
	viper.SetDefault("HTTP.ListenAddress", ":8080")

	// Set directories to read config from
	if d := os.Getenv("XDG_CONFIG_HOME"); len(d) > 0 {
		viper.AddConfigPath(filepath.Join(d, appID))
	}
	if d, err := homedir.Dir(); err == nil {
		viper.AddConfigPath(filepath.Join(d, "."+appID))
	}
	if runtime.GOOS != "windows" {
		viper.AddConfigPath(filepath.Join("/", "etc", appID))
	}
	viper.AddConfigPath(".")
	viper.SetConfigName(appID)
	viper.SetEnvPrefix(appID)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// AutomaticEnv is not enough for Unmarshal: a key has to be bound to its
	// environment variable before it can override the config file.
	for _, key := range []string{"HTTP.ListenAddress", "HTTP.AuthenticationRealm"} {
		err := viper.BindEnv(key)
		if err != nil {
			log.Fatalf("Failed to bind environment variable for %s: %s", key, err)
		}
	}

	// Read the configuration
	var notFound viper.ConfigFileNotFoundError

	err := viper.ReadInConfig()

	switch {
	case err == nil:
		log.Printf("Read config from %s", viper.ConfigFileUsed())
	case errors.As(err, &notFound):
		log.Print("No config file found, using defaults and environment variables")
	default:
		log.Fatalf("Failed to read config: %s", err)
	}

	go viper.WatchConfig()
}

func GetConfig() (c *Config) {
	c = new(Config)
	err := viper.Unmarshal(c)
	if err != nil {
		panic(err)
	}
	return
}

func GetBackendConfig(backendID string) *viper.Viper {
	b := viper.Sub("Backends")
	if b == nil {
		return nil
	}
	return b.Sub(backendID)
}

type HTTPConfig struct {
	ListenAddress       string
	AuthenticationRealm string
	// TODO - tls config
}

type Config struct {
	Backends              map[string]map[string]interface{}
	AuthenticationBackend string
	StorageBackend        string
	HTTP                  *HTTPConfig
}
