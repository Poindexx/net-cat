package netcat

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
	}
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func (s *Server) addClient(client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients = append(s.clients, client)
}

func (s *Server) removeClient(client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for i, c := range s.clients {
		if c == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
}

func (s *Server) broadcast(message string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for _, client := range s.clients {
		_, err := client.writer.WriteString(message + "\n")
		if err != nil {
			fmt.Printf("Error writing to client %s: %s\n", client.name, err)
		}
		err = client.writer.Flush()
		if err != nil {
			fmt.Printf("Error flushing client %s: %s\n", client.name, err)
		}
	}
}

func (s *Server) logMessage(message string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.messageLog = append(s.messageLog, message)
}

func (s *Server) sendPreviousMessages(client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for _, msg := range s.messageLog {
		_, err := client.writer.WriteString(msg + "\n")
		if err != nil {
			fmt.Printf("Error sending previous messages to client %s: %s\n", client.name, err)
			return
		}
	}
	err := client.writer.Flush() // сбрасываем данные из буфера в файл
	if err != nil {
		fmt.Printf("Error flushing previous messages to client %s: %s\n", client.name, err)
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	fmt.Printf("listening on the port: %s\n", s.listenAddr)
	defer ln.Close()
	s.ln = ln
	go s.acceptLoop()

	<-s.quitch
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("accept err: ", err)
			continue
		}
		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	client := NewClient(conn)
	linuxLogo := `
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\\dS"qML
 |    '.       |  \\' Zq
_)      \\.___.,|     .'
\\____   )MMMMMP|   .'
     '-'       '--'`

	conn.Write([]byte("Welcome to TCP-Chat!\n"))
	conn.Write([]byte(linuxLogo))
	conn.Write([]byte("\n[ENTER YOUR NAME]: "))
	name, _ := client.reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		conn.Write([]byte("\n[ENTER CORRECT YOUR NAME]: "))
		name1, _ := client.reader.ReadString('\n')
		name1 = strings.TrimSpace(name1)
		if name1 != "" {
			name = name1
		} else {
			return
		}
	}
	client.name = name

	s.addClient(client)
	defer s.removeClient(client)

	fmt.Printf("%s has joined our chat...\n", client.name)
	s.sendPreviousMessages(client)

	s.broadcast(fmt.Sprintf("%s has joined out chat...", client.name))
	s.logMessage(fmt.Sprintf("%s has joined out chat...", client.name))

	for {
		message, err := client.reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%s has left our chat...\n", client.name)
			s.broadcast(fmt.Sprintf("%s has left our chat...", client.name))
			s.logMessage(fmt.Sprintf("%s has left our chat...", client.name))
			return
		}
		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("[%s][%s]: %s\n", timestamp, client.name, message)
		s.broadcast(fmt.Sprintf("[%s][%s]: %s", timestamp, client.name, message))
		s.logMessage(fmt.Sprintf("[%s][%s]: %s", timestamp, client.name, message))
	}
}
