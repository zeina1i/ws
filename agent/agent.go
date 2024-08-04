package agent

import (
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	hosts = []string{
		"localhost:8080",
		"localhost:8081",
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx = context.Background()
)

func Observe() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, host := range hosts {
				go checkStatus(host)
			}
		}
	}
}

func checkStatus(host string) {
	resp, err := http.Get("http://" + host + "/status")
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	err = rdb.Set(ctx, host, string(body), 0).Err()
	if err != nil {
		log.Println(err)
	}
}
