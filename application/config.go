package application

import (
	"os"
	"strconv"
)

type Config struct {
	RedisAddress string
	ServerPort   uint16
}

func LoadConfig() Config {
	conf := Config{
		RedisAddress: "localhost:6379",
		ServerPort:   3000,
	}

	if redisAddr, exist := os.LookupEnv("REDDIS_ADDR"); exist {
		conf.RedisAddress = redisAddr
	}

	if serverPort, exist := os.LookupEnv("SERVER_PORT"); exist {
		if serverPort, err := strconv.ParseInt(serverPort, 10, 16); err == nil {
			conf.ServerPort = uint16(serverPort)
		}
	}

	return conf
}
