package database

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"org.donghyuns.com/hls/converter/configs"
)

type RedisStruct struct {
	*redis.Client
}

func RedisInstance() (*RedisStruct, error) {
	redisConfig := configs.RedisConfig

	redisInstance := redis.NewClient(&redis.Options{
		Addr: redisConfig.Addr,
		// Username: redisConfig.UserName,
		Password: redisConfig.Password,
		DB:       0,
	})

	var ctx = context.Background()

	_, pingError := redisInstance.Ping(ctx).Result()

	if pingError != nil {
		log.Printf("[REDIS] Ping Redis Error: %v", pingError)
		return nil, pingError
	}
	connect := &RedisStruct{redisInstance}
	return connect, nil
}

func (rdb *RedisStruct) Set(key string, token string) error {
	var ctx = context.Background()
	// var item = map[string]string {
	// 	objKey: token}

	// tokenItem := types.RedisToken{
	// 	TokenItem: item,
	// }

	expireDuration := 3 * time.Hour
	setErr := rdb.Client.Set(ctx, key, token, expireDuration).Err()

	if setErr != nil {
		log.Printf("[REDIS] Key Set Error: %v", setErr)

		return setErr
	}

	return nil
}

func (rdb *RedisStruct) Get(key string) (string, error) {
	var ctx = context.Background()

	getItem, getErr := rdb.Client.Get(ctx, key).Result()

	switch {
	case getErr == redis.Nil:
		log.Printf("[REDIS] No Value Found")
		return "", nil

	case getErr != nil:
		log.Printf("[REDIS] Get Key Error: %v", getErr)
		return "", getErr

	default:
		return getItem, nil
	}
}

func (rdb *RedisStruct) GetAll(key string) (string, error) {
	var ctx = context.Background()

	getItemList, getErr := rdb.Client.Get(ctx, key).Result()

	if getErr != nil {
		log.Printf("[REDIS] Get Key Error: %v", getErr)
		return "", getErr
	}

	return getItemList, nil
}

func (rdb *RedisStruct) Delete(key string) error {
	var ctx = context.Background()

	deleteErr := rdb.Client.Del(ctx, key).Err()

	if deleteErr != nil {
		log.Printf("[REDIS] Delete Key Error: %v", deleteErr)
		return deleteErr
	}

	return nil
}

func (rdb *RedisStruct) RedisLoginSet(refreshToken string, email string, userId string, oauthName string, isAdmin string) error {
	sessionInfo := map[string]string{
		"userEmail": email,
		"userId":    userId,
		"oauthName": oauthName,
		"isAdmin":   isAdmin,
	}

	var ctx = context.Background()

	var data []interface{}

	for key, value := range sessionInfo {
		data = append(data, key, value)
	}

	err := rdb.Client.HSet(ctx, refreshToken, data...).Err()

	if err != nil {
		log.Printf("[REDIS] Set Value Error: %v", err)
		return err
	}

	return nil
}

func (rdb *RedisStruct) RedisLoginGet(sessionId string) (map[string]string, error) {
	var ctx = context.Background()

	getItem, getErr := rdb.Client.HGetAll(ctx, sessionId).Result()

	switch {
	case getErr == redis.Nil:
		log.Printf("[REDIS] No Value Found")
		return nil, nil

	case getErr != nil:
		log.Printf("[REDIS] Get Key Error: %v", getErr)
		return nil, getErr

	default:
		return getItem, nil
	}
}
