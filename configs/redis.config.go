package configs

import "os"

type RedisConfigStruct struct {
	Addr     string
	Password string
	UserName string
	// DB int
}

var RedisConfig RedisConfigStruct

func SetRedisConfig() {
	RedisConfig.Addr = os.Getenv("REDIS_ADDR")
	RedisConfig.Password = os.Getenv("REDIS_PASSWORD")
	RedisConfig.UserName = os.Getenv("REDIS_USER")
	// redisConfig.DB
}
