package config

import (
	"fmt"

	"github.com/mgutz/logxi/v1"
	"github.com/spf13/viper"
)

const (
	AppName        = "bt"
	ConfigFileName = "bt-config"

	defaultMongoAuth     = "mongo:btmongo"
	defaultMongoHost     = "localhost:27017"
	defaultMongoDatabase = "bt"
	defaultWWWHost       = "https://localhost:8889"

	envWWWHost       = "WWW_HOST"
	envMongoAuth     = "MONGO_AUTH"
	envMongoHost     = "MONGO_HOST"
	envMongoDatabase = "MONGO_DATABASE"
	envSecret        = "SECRET"
	envTesting       = "TESTING"
)

var (
	logger = log.New("config")
)

func GetMongoDatabase() string {
	return viper.GetString(envMongoDatabase)
}

func GetMongoURL() string {
	mongoAuth := viper.GetString(envMongoAuth)
	mongoHost := viper.GetString(envMongoHost)
	databaseName := viper.GetString(envMongoDatabase)
	return fmt.Sprintf("mongodb://%v@%v/%v", mongoAuth, mongoHost, databaseName)
}

func GetSecret() []byte {
	return []byte(viper.GetString(envSecret))
}

func GetWWWHost() string {
	return viper.GetString(envWWWHost)
}

func IsTesting() bool {
	return viper.GetBool(envTesting)
}

func init() {
	// Set default values
	viper.SetEnvPrefix(AppName)
	viper.SetDefault(envMongoAuth, defaultMongoAuth)
	viper.SetDefault(envMongoHost, defaultMongoHost)
	viper.SetDefault(envMongoDatabase, defaultMongoDatabase)
	viper.AutomaticEnv()

	// Set config files
	viper.SetConfigName(ConfigFileName)
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("$HOME/.bt_config")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	// Ensure that secret is set
	if secret := viper.GetString(envSecret); len(secret) == 0 {
		panic("env.SECRET must be defined")
	}
}
