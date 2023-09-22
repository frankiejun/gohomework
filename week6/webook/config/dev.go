//go:build !k8s

package config

import (
	"fmt"
	"time"
)

var Config = WebookConfig{
	DB: DBConfig{
		DSN: "root:root@tcp(localhost:13316)/webook",
	},
	Redis: RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	},
}

func main() {
	timer := time.NewTimer(2 * time.Second)

	go func() {
		<-timer.C
		fmt.Println("定时器触发")
	}()

	time.Sleep(1 * time.Second)

	reset := timer.Reset(3 * time.Second)
	if reset {
		fmt.Println("定时器已重置")
	}

	time.Sleep(4 * time.Second)
}
