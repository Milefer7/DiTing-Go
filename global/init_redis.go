package global

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"log"
)

var Rdb *redis.Client

func init() {
	addr := viper.GetString("redis.addr")
	//password := viper.GetString("redis.password")
	Rdb = redis.NewClient(&redis.Options{
		Addr: addr,
		//Password: password, // 密码
		DB:       0,  // 数据库
		PoolSize: 20, // 连接池大小
	})
	// 检查连接是否成功
	_, err := Rdb.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	} else {
		fmt.Println("Connected to Redis successfully.")
	}
}
