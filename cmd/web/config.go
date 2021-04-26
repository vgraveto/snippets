package main

import (
	"github.com/spf13/viper"
	"github.com/vgraveto/snippets/pkg/models/dbapi"
	"log"
	"time"
)

type configType struct {
	HttpPort string

	// API information
	HttpIdleTimeout      time.Duration // number of seconds
	HttpReadTimeout      time.Duration // number of seconds
	HttpWriteTimeout     time.Duration // number of seconds
	deadlineWaitForClose time.Duration // number of seconds
	sessionLifetime      time.Duration // number of hours

	// Database connection data
	DB dbapi.DBapi
}

func readConfig(errorLog *log.Logger, path, filename string) (globalData configType) {
	viper.SetConfigName(filename)
	if len(path) == 0 {
		path = "."
	}
	viper.AddConfigPath(path)
	err := viper.ReadInConfig()
	if err != nil {
		errorLog.Fatalf("config: %s\\%s.toml - %v\n", path, filename, err)
	}

	// impose mandatory key/values in config file
	if !viper.IsSet("dbase.url") {
		log.Fatalf("Key/Value not set in file %s - dbase.url", filename)
	}
	// TODO implement all required checks for config file

	globalData.HttpPort = viper.GetString("global.httpPort")

	globalData.HttpIdleTimeout = time.Duration(viper.GetInt("api.httoIdleTimeout")) * time.Second
	globalData.HttpReadTimeout = time.Duration(viper.GetInt("api.httpReadTimeout")) * time.Second
	globalData.HttpWriteTimeout = time.Duration(viper.GetInt("api.httpWriteTimeout")) * time.Second
	globalData.deadlineWaitForClose = time.Duration(viper.GetInt("api.deadlineWaitForClose")) * time.Second
	globalData.sessionLifetime = time.Duration(viper.GetInt("api.sessionLifetime")) * time.Hour

	globalData.DB.URL = viper.GetString("dbase.url")

	return globalData
}
