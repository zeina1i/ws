package manager

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"math/rand"
	"net/http"
	"time"
)

var (
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx = context.Background()
)

func getFreeNode(w http.ResponseWriter, r *http.Request) {
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		http.Error(w, "Error getting keys from Redis", http.StatusInternalServerError)
		return
	}

	var freeNodes []string

	for _, key := range keys {
		status, err := rdb.Get(ctx, key).Result()
		if err != nil {
			http.Error(w, "Error getting status from Redis", http.StatusInternalServerError)
			return
		}

		if status == "free" {
			freeNodes = append(freeNodes, key)
		}
	}

	if len(freeNodes) == 0 {
		http.Error(w, "No free nodes available", http.StatusNotFound)
		return
	}

	rand.Seed(time.Now().UnixNano())

	freeNode := freeNodes[rand.Intn(len(freeNodes))]

	fmt.Fprintf(w, "Free node: %s", freeNode)
}

func Serve() {
	http.HandleFunc("/getFreeNode", getFreeNode)

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
