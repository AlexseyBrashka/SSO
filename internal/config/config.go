package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env            string     `yaml:"env" env-default:"local"`
	StoragePath    string     `yaml:"storage_path" env-required:"true"`
	GRPC           GRPCConfig `yaml:"grpc"`
	MigrationsPath string
	TokenTTL       time.Duration `yaml:"token_ttl" env-default:"1h"`
}

type AuthService struct {
	regLimiter   RegLimiter   `yaml:"regLimiter" env-required:"true"`
	loginLimiter LoginLimiter `yaml:"loginLimiter" env-required:"true"`
}

type RegLimiter struct {
	time     int `yaml:"time" env-required:"true"`
	activity int `yaml:"activity" env-required:"true"`
}

type LoginLimiter struct {
	time     int `yaml:"time" env-required:"true"`
	activity int `yaml:"activity" env-required:"true"`
}

type GRPCConfig struct {
	Port        int           `yaml:"port"`
	Timeout     time.Duration `yaml:"timeout"`
	AuthService AuthService   `yaml:"AuthService"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config path is empty: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {

	res, err := os.LookupEnv("CONFIG_PATH")

	if !err {
		return ""
	}

	return res
}
