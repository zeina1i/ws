package main

import (
	"flag"
	"github.com/zeina1i/ws/ws"
)

func main() {
	port := flag.String("port", "8081", "port to serve on")
	flag.Parse()

	ws.Serve(*port)
}
