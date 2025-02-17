package application

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	RedisAddress string
	ServerPort   uint16
}

func LoadConfig() Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conf := Config{
		RedisAddress: "localhost:6379",
		ServerPort:   3000,
	}

	if redisAddr, exist := os.LookupEnv("GOSERVER_REDDIS_ADDR"); exist {
		conf.RedisAddress = redisAddr
	}

	if serverPort, exist := os.LookupEnv("GOSERVER_SERVER_PORT"); exist {
		if serverPort, err := strconv.ParseInt(serverPort, 10, 16); err == nil {
			conf.ServerPort = uint16(serverPort)
		}
	}

	return conf
}
