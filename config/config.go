package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func ReadConfig(appID string) {
	viper.Debug()

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
	viper.AutomaticEnv()

	// Read the configuration
	log.Println("Read in config:", viper.ReadInConfig())
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
