package cmd

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

// Config holds all application configuration
type Config struct {
	Bind       BindOptions             `yaml:"bind"`
	DB         DBConfigurations        `yaml:"database"`
	Instrument InstrumentConfiguration `yaml:"intrument"`
	Auth       AuthConfig              `yaml:"auth"`
	OAuth      OAuthConfig             `yaml:"oauth"`
}

type InstrumentConfiguration struct {
	Enabled          bool          `yaml:"enabled"`
	CollectorAddress string        `yaml:"collector_address"`
	Timeout          time.Duration `yaml:"timeout"`
}

type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type OAuthConfig struct {
	Google    OAuthProviderConfig `yaml:"google"`
	Facebook  OAuthProviderConfig `yaml:"facebook"`
	GitHub    OAuthProviderConfig `yaml:"github"`
	Microsoft OAuthProviderConfig `yaml:"microsoft"`
}

type OAuthProviderConfig struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
}

type BindOptions struct {
	HTTP int `yaml:"http"`
	GRPC int `yaml:"grpc"`
}

type DBConfigurations struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func (cfg DBConfigurations) CreateDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
}

// Define package-level variables to store configuration
var (
	configOnce sync.Once
	config     Config
	klogOnce   sync.Once
)

// GetConfigurations loads the configuration from a yaml file
func GetConfigurations() Config {
	// Initialize configuration only once using sync.Once
	configOnce.Do(func() {
		// Initialize klog only once
		klogOnce.Do(func() {
			klog.InitFlags(nil)
			klog.EnableContextualLogging(true)
		})

		// Define a configPath variable to store the path to the config file
		var configPath string

		// Create a dedicated FlagSet for this function
		flagSet := flag.NewFlagSet("config", flag.ContinueOnError)
		flagSet.StringVar(&configPath, "cfg", "config.yaml", "Configuration File")

		// If the main flags have been parsed, extract the config path from there
		if flag.Parsed() {
			if cfgFlag := flag.Lookup("cfg"); cfgFlag != nil {
				configPath = cfgFlag.Value.String()
			}
		} else {
			// Otherwise, parse the arguments
			flag.StringVar(&configPath, "cfg", "config.yaml", "Configuration File")
			flag.Parse()
		}

		file, err := os.Open(configPath)
		if err != nil {
			klog.Fatalf("cannot read config file:%v", err)
		}
		defer file.Close()

		if err := yaml.NewDecoder(file).Decode(&config); err != nil {
			klog.Fatalf("cannot unmarshal the yaml file %v", err)
		}
	})

	return config
}
