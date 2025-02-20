package application

import (
	"encoding/json"
	"fmt"
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

	setPostgresAddressFromEnvVariables(&conf)

	if serverPort, exist := os.LookupEnv("GOSERVER_SERVER_PORT"); exist {
		if serverPort, err := strconv.ParseInt(serverPort, 10, 16); err == nil {
			conf.ServerPort = uint16(serverPort)
		}
	}

	return conf
}

func setPostgresAddressFromEnvVariables(conf *Config) {
	var postgresCredentials struct {
		DatabaseName string `json:"databaseName"`
		Username     string `json:"username"`
		Password     string `json:"password"`
	}

	postgresAddr, exist := os.LookupEnv("GOSERVER_POSTGRES_ADDR")
	if !exist {
		fmt.Println("No postgres address specified in environment")
		return
	}

	postgresCredentialsFileName, exist := os.LookupEnv("GOSERVER_POSTGRES_CREDENTIALS")
	if !exist {
		fmt.Println("No postgres credentials file specified in environment")
		return
	}

	file, err := os.ReadFile(postgresCredentialsFileName)
	if err != nil {
		_ = fmt.Errorf("error reading postgres credentials file: %v", err)
		return
	}

	err = json.Unmarshal(file, &postgresCredentials)
	if err != nil {
		_ = fmt.Errorf("error unmarshalling postgres credentials: %v", err)
		return
	}

	conf.PostgresAddress = "postgresql://" + postgresAddr +
		"/" + postgresCredentials.DatabaseName +
		"?user=" + postgresCredentials.Username + "&password=" + postgresCredentials.Password
}
