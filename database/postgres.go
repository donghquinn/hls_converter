package database

import (
	"log"
	"strconv"

	"github.com/donghquinn/gdct"
	"github.com/donghquinn/hls_converter/configs"
)

type DbInstance struct {
	*gdct.DataBaseConnector
}

func DatabaseInstance() *gdct.DataBaseConnector {
	databaseConfig := configs.DatabaseConfiguration
	portNumber, convErr := strconv.Atoi(databaseConfig.DbPort)

	if convErr != nil {
		log.Printf("Error converting port number: %v", convErr)
		return nil
	}

	dbInstance, conErr := gdct.InitConnection(gdct.PostgreSQL, gdct.DBConfig{
		Host:         databaseConfig.DbHost,
		Port:         portNumber,
		UserName:     databaseConfig.DbUser,
		Password:     databaseConfig.DbPasswd,
		Database:     databaseConfig.DbName,
		SslMode:      "disable",
		MaxLifeTime:  600,
		MaxIdleConns: 50,
		MaxOpenConns: 10,
	})

	if conErr != nil {
		log.Printf("Error connecting to database: %v", conErr)
		return nil
	}

	return dbInstance
}
