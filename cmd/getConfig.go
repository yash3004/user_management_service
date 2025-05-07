package cmd

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

type Config struct {
	Bind BindOptions      `yaml:"bind"`
	DB   DBConfigurations `yaml:"database"`
	Instrument InstrumentConfiguration `yaml:"intrument"`
}

type InstrumentConfiguration struct {
	Enabled          bool          `yaml:"enabled"`
	CollectorAddress string        `yaml:"collector_address"`
	Timeout          time.Duration `yaml:"timeout"`
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

func GetConfigurations() Config {
	klog.InitFlags(nil)
	klog.EnableContextualLogging(true)
	var (
		configFile = flag.String("cfg", "config.yaml", "Configuration File")
	)
	flag.Parse()
	file, err := os.Open(*configFile)
	if err != nil {
		klog.Fatalf("cannot read config file:%v", err)
	}
	defer file.Close()

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		klog.Fatalf("cannot unmarshal the yaml file %v", err)
	}
	return config

}
