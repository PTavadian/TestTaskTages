package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	IsDebug *bool `yaml:"is_debug" env:"IS_DEBUG"`

	Listen struct {
		GRPC struct {
			Host string `yaml:"host" env:"LISTEN_GRPC_HOST" env-default:"localhost"`
			Port string `yaml:"port" env:"LISTEN_GRPC_PORT" env-default:"50051"`
		} `yaml:"grpc"`
	} `yaml:"listen"`

	Postgres struct {
		Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
		Port     string `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
		Database string `yaml:"database" env:"POSTGRES_DATABASE"`
		Username string `yaml:"username" env:"POSTGRES_USERNAME"`
		Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	} `yaml:"postgres"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Println("No .env file found")
		}

		instance = &Config{}
		err = cleanenv.ReadConfig("config.yaml", instance)
		if err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Println(help)
			log.Fatal(err)
		}

		err = cleanenv.ReadEnv(instance)
		if err != nil {
			log.Fatal(err)
		}
	})
	return instance
}

// для генерации мока
type ConfigProvider interface {
	GetConfig() *Config
}
