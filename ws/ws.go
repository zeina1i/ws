package ws

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	clients = make(map[*websocket.Conn]bool)
	mu      sync.Mutex

	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx = context.Background()
)

func echo(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	mu.Lock()
	clients[ws] = true
	mu.Unlock()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		log.Printf("Received: %s", message)

		err = rdb.Publish(ctx, "channel", message).Err()
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func clientCount(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	count := len(clients)
	mu.Unlock()

	fmt.Fprintf(w, "Number of active clients: %d", count)
}

func status(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("full"))
}

func handleMessages() {
	pubsub := rdb.Subscribe(ctx, "channel")
	_, err := pubsub.Receive(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	ch := pubsub.Channel()

	for msg := range ch {
		mu.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

func Serve() {
	http.HandleFunc("/ws", echo)
	http.HandleFunc("/clientCount", clientCount)
	http.HandleFunc("/status", status)

	go handleMessages()

	log.Fatal(http.ListenAndServe(":8081", nil))
}
