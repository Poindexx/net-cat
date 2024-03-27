package netcat

import (
	"bufio"
	"net"
	"sync"
)

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	clients    []*Client
	clientsMu  sync.Mutex
	messageLog []string
}

type Client struct {
	name   string
	reader *bufio.Reader
	writer *bufio.Writer
}
