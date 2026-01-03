package utils

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Configuration struct {
	AppName     string
	Port        string
	Debug       bool
	Limit       int
	PathLogging string
	DB          DatabaseConfig
}

type DatabaseConfig struct {
	Name     string
	Username string
	Password string
	Host     string
	Port     string
	MaxConn  int32
}

func ReadConfiguration() (Configuration, error) {
	// get config from env file
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	err := viper.ReadInConfig()
	if err != nil {
		return Configuration{}, err
	}

	// get config from os variable
	viper.AutomaticEnv()

	// get config from flag
	pflag.Int("port-app", 0, "port for app golang")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	return Configuration{
		AppName:     viper.GetString("APP_NAME"),
		Port:        viper.GetString("PORT"),
		Debug:       viper.GetBool("DEBUG"),
		Limit:       viper.GetInt("LIMIT"),
		PathLogging: viper.GetString("PATH_LOGGING"),
		DB: DatabaseConfig{
			Name:     viper.GetString("DATABASE_NAME"),
			Username: viper.GetString("DATABASE_USERNAME"),
			Password: viper.GetString("DATABASE_PASSWORD"),
			Host:     viper.GetString("DATABASE_HOST"),
			Port:     viper.GetString("DATABASE_PORT"),
			MaxConn:  viper.GetInt32("DATABASE_MAX_CONN"),
		},
	}, nil

}
