package config

import (
	"github.com/kelseyhightower/envconfig"
	"log"
)

type Config struct {
	DBConfig
	DRPConfig
}

type DBConfig struct {
	Name string `required:"true"`
	Host string `required:"true"`
	User string `required:"true"`
	Pass string `required:"true"`
}

type DRPConfig struct {
	Instance string `required:"true" split_words:"true"`
}

func GetConf() Config {
	var dbConf DBConfig
	if err := envconfig.Process("db", &dbConf); err != nil {
		log.Fatal(err.Error())
	}

	var drpConf DRPConfig
	if err := envconfig.Process("drp", &drpConf); err != nil {
		log.Fatal(err.Error())
	}

	return Config{dbConf, drpConf}
}
