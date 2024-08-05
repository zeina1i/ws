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
	fiveSecondsAgo := float64(time.Now().Unix() - 5)

	nodes, err := rdb.ZRangeByScore(ctx, "hosts", &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", fiveSecondsAgo),
		Max: "+inf",
	}).Result()

	if err != nil {
		http.Error(w, "Error getting nodes from Redis", http.StatusInternalServerError)
		return
	}

	if len(nodes) == 0 {
		http.Error(w, "No free nodes available", http.StatusNotFound)
		return
	}

	rand.Seed(time.Now().UnixNano())
	node := nodes[rand.Intn(len(nodes))]

	fmt.Fprint(w, node)
}

func Serve() {
	http.HandleFunc("/getFreeNode", getFreeNode)

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
