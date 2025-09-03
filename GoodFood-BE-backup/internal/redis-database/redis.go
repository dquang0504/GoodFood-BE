package redisdatabase

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var(
	Client *redis.Client
	Ctx = context.Background()
)

func InitRedis(){
	Client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})

	//Checking redis connection
	pong,err := Client.Ping(Ctx).Result()
	if err != nil{
		log.Fatalf("Couldn't connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis:",pong)

	//Setting key-value
	err = Client.Set(Ctx,"foo","bar",0).Err()
	if err != nil{
		log.Fatalf("Lỗi khi set key: %v", err)
	}

	//Get value from redis
	val,err := Client.Get(Ctx,"foo").Result()
	if err != nil{
		log.Fatalf("Lỗi khi get key: %v", err)
	}
	fmt.Println("Test value: ",val)
}

func GetClient() *redis.Client{
	return Client
}