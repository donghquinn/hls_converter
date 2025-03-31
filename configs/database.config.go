package configs

import "os"

type DatabaseConfig struct {
	// PostgreSQL
	// DbUrl    string
	DbHost   string
	DbPort   string
	DbName   string
	DbUser   string
	DbPasswd string
	// Neo4j
	// GraphUri    string
	// GraphUser   string
	// GraphPasswd string
}

var DatabaseConfiguration DatabaseConfig

func SetDatabaseConfiguration() {
	// PostgreSQL
	DatabaseConfiguration.DbHost = os.Getenv("POSTGRES_HOST")
	DatabaseConfiguration.DbPort = os.Getenv("POSTGRES_PORT")
	DatabaseConfiguration.DbName = os.Getenv("POSTGRES_NAME")
	DatabaseConfiguration.DbUser = os.Getenv("POSTGRES_USER")
	DatabaseConfiguration.DbPasswd = os.Getenv("POSTGRES_PASSWD")
	// Neo4j
	// DatabaseConfiguration.GraphUri = os.Getenv("GRAPH_URL")
	// DatabaseConfiguration.GraphUser = os.Getenv("GRAPH_USER")
	// DatabaseConfiguration.GraphPasswd = os.Getenv("GRAPH_PASSWORD")
}
