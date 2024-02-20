package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `yaml:"env" env-default:"local"` // it's better not to use in prod
	Storage  `yaml:"storage"`
	TokenTTL time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPC     `yaml:"grpc"`
}

type Storage struct {
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	Host     string `yaml:"host" env-default:"localhost"`
	DbName   string `yaml:"db_name" env-required:"true"`
}

func (s Storage) DSN() string {
	return fmt.Sprintf("%s:%s@%s/%s", s.User, s.Password, s.Host, s.DbName)
}

type GRPC struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config { // this func will not return error, but panic instead
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist" + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config" + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config value")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	return res
}
