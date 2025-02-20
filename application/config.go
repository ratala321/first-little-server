package application

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type EnvDatabase string

const (
	PostgresEnv EnvDatabase = "postgres"
	ReddisEnv   EnvDatabase = "reddis"
)

type Config struct {
	Database        EnvDatabase
	RedisAddress    string
	PostgresAddress string
	ServerPort      uint16
}

func LoadConfig() Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conf := Config{
		Database:        ReddisEnv,
		RedisAddress:    "localhost:6379",
		PostgresAddress: "localhost:5432",
		ServerPort:      3000,
	}

	if databaseEnv, exist := os.LookupEnv("GOSERVER_DATABASE"); exist {
		conf.Database = EnvDatabase(databaseEnv)
	}

	if redisAddr, exist := os.LookupEnv("GOSERVER_REDDIS_ADDR"); exist {
		conf.RedisAddress = redisAddr
	}

	if postgresAddr, exist := os.LookupEnv("GOSERVER_POSTGRES_ADDR"); exist {
		conf.PostgresAddress = postgresAddr
	}

	if serverPort, exist := os.LookupEnv("GOSERVER_SERVER_PORT"); exist {
		if serverPort, err := strconv.ParseInt(serverPort, 10, 16); err == nil {
			conf.ServerPort = uint16(serverPort)
		}
	}

	return conf
}
