package main

import (
	"github.com/spf13/viper"
	"github.com/vgraveto/snippets/pkg/models"
	"github.com/vgraveto/snippets/pkg/models/dbmysql"
	"log"
	"time"
)

type configType struct {
	HttpPort string

	// api JWT token values
	TD models.TokenData

	// API information
	HttpIdleTimeout      time.Duration // number of seconds
	HttpReadTimeout      time.Duration // number of seconds
	HttpWriteTimeout     time.Duration // number of seconds
	deadlineWaitForClose time.Duration // number of seconds
	sessionLifetime      time.Duration // number of hours

	// Database connection data
	DB dbmysql.DBdata
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
	if !viper.IsSet("token.issuerName") {
		log.Fatalf("Key/Value not set in file %s - token.issuerName", filename)
	}
	if !viper.IsSet("token.validTime") {
		log.Fatalf("Key/Value not set in file %s - token.validTime", filename)
	}
	if !viper.IsSet("token.signingKey") {
		log.Fatalf("Key/Value not set in file %s - token.signingKey", filename)
	}
	// TODO implement all required checks for config file

	globalData.HttpPort = viper.GetString("global.httpPort")

	globalData.TD.TokenIssuerName = viper.GetString("token.issuerName")
	globalData.TD.TokenValidTime = time.Duration(viper.GetInt("token.validTime")) * time.Hour
	globalData.TD.TokenSigningKey = viper.GetString("token.signingKey")

	globalData.HttpIdleTimeout = time.Duration(viper.GetInt("api.httoIdleTimeout")) * time.Second
	globalData.HttpReadTimeout = time.Duration(viper.GetInt("api.httpReadTimeout")) * time.Second
	globalData.HttpWriteTimeout = time.Duration(viper.GetInt("api.httpWriteTimeout")) * time.Second
	globalData.deadlineWaitForClose = time.Duration(viper.GetInt("api.deadlineWaitForClose")) * time.Second
	globalData.sessionLifetime = time.Duration(viper.GetInt("api.sessionLifetime")) * time.Hour

	globalData.DB.Protocol = viper.GetString("dbase.protocol")
	globalData.DB.Server = viper.GetString("dbase.server")
	globalData.DB.Dbase = viper.GetString("dbase.database")
	globalData.DB.Username = viper.GetString("dbase.username")
	globalData.DB.Password = viper.GetString("dbase.password")
	globalData.DB.ServerCA = viper.GetString("dbase.serverCA")
	globalData.DB.ClientCert = viper.GetString("dbase.clientCert")
	globalData.DB.ClientKey = viper.GetString("dbase.clientKey")
	globalData.DB.DbConnMaxLifetime = time.Duration(viper.GetInt("dbase.dbConnMaxLifetime")) * time.Second

	/*	// Push Token values to services.token
		services.IssuerName = GlobalData.tokenIssuerName
		services.TokenValidTime = GlobalData.tokenValidTime
		services.MySigningKey = GlobalData.tokenSigningKey
	*/
	return globalData
}
