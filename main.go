package main

import (
	"fmt"
	"os"
	netcat "netcat/server"
)

func main() {
	port := ":8989"
	if len(os.Args) == 2 {
		port = ":" + os.Args[1]
	} else if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}
	server := netcat.NewServer(port)
	server.Start()
}
