package servertcp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/nxczje/sreq/pkg/sreq"
)

var method = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD", "TRACE", "CONNECT"}

// Server ...
type server struct {
	host string
	port string
}

// Client ...
type client struct {
	conn net.Conn
}

// Config ...
type Config struct {
	Host string //default 127.0.0.1
	Port string //default 8443
}

// Create a new server
func New(config *Config) *server {
	if config.Host == "" {
		config.Host = "127.0.0.1"
	}
	if config.Port == "" {
		config.Port = "8443"
	}
	return &server{
		host: config.Host,
		port: config.Port,
	}
}

// Run server with time up (seconds) and u can push func to send request to server
// use time = 0 -> no shutdown
func (server *server) Run(timerunning int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", server.host, server.port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Println("Server running on", fmt.Sprintf("%s:%s", server.host, server.port))
	start := time.Now()
	end := start.Add(time.Duration(timerunning) * time.Second)
	for {
		if timerunning != 0 {
			if time.Now().After(end) {
				break
			}
			if timerunning > 0 {
				go func() {
					time.Sleep(time.Duration(timerunning) * time.Second)
					url := fmt.Sprintf("http://%s:%s/req2exit", server.host, server.port)
					req, _ := http.NewRequest("GET", url, nil)
					sreq.Send(&sreq.Request{Req: req}, "")
				}()
			}
		}
		if err != nil {
			log.Fatal(err)
		}
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		client := &client{
			conn: conn,
		}
		go client.handleRequest()
	}
}

func (client *client) handleRequest() {
	reader := bufio.NewReader(client.conn)
	peek, _ := reader.Peek(7)
	check := false // true is http request
	for _, v := range method {
		if strings.Contains(string(peek), v) {
			check = true
		}
	}
	pretty.Printf("::::::::::::%20s   :::::::::::::\n", time.Now().Format("02/01/2006-15:04"))
	if check {
		req, err := http.ReadRequest(reader)
		if err != nil {
			client.conn.Close()
			return
		}
		var buffer bytes.Buffer
		req.Write(&buffer)
		pretty.Printf("%s\n", buffer.String())
		SaveLog(buffer.String())
	} else {
		for {
			message, err := io.ReadAll(reader)
			if err != nil {
				client.conn.Close()
				return
			}
			pretty.Printf("%s", string(message))
			SaveLog(string(message))
		}
	}
	pretty.Println("::::::::::::::::::::::::::::::::::::::::::::::::")
	client.conn.Close()
}

func SaveLog(data string) {
	//check exist file
	_, err := os.Stat("log.txt")
	if os.IsNotExist(err) {
		// create file
		file, err := os.Create("log.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}
	//append file
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	if _, err := file.WriteString(data + "\n"); err != nil {
		log.Fatal(err)
	}
}
