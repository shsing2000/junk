package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"regexp"
)

var quitCmd = regexp.MustCompile("^quit\r?\n$")

type Client struct {
	id   string
	send chan string
	recv chan string
	quit chan struct{}
	conn net.Conn
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	c := &Client{
		id:   conn.LocalAddr().String(),
		send: make(chan string),
		recv: make(chan string),
		quit: make(chan struct{}),
		conn: conn,
	}

	//read user messages
	go readMessages(os.Stdin, c.send, c.quit)
	//read server messages
	go readMessages(conn, c.recv, c.quit)

	return c, nil
}

func (c *Client) write(msg string) {
	c.conn.Write([]byte(msg))
}

func (c *Client) Close() {
	close(c.send)
	close(c.recv)
	close(c.quit)

	c.conn.Close()
}

func (c *Client) Listen() {
	log.Printf("connected to server: %s\n", c.id)

	for {
		select {
		case msg := <-c.recv:
			//msg from the server
			log.Print(msg)
		case msg := <-c.send:
			//msg to the server
			c.write(msg)
		case <-c.quit:
			log.Println("closing")
			c.write("quit\n")
			return
		}
	}
}

func readMessages(rd io.Reader, dst chan string, quit chan struct{}) {
	r := bufio.NewReader(rd)
	for {
		s, err := r.ReadString('\n')
		if err != nil {
			return
		}

		if quitCmd.MatchString(s) {
			quit <- struct{}{}
			return
		}

		log.Printf("sending: %s", s)
		dst <- s
	}
}

func main() {

	client, err := NewClient("127.0.0.1:8888")
	if err != nil {
		log.Fatal("could not create client ", err)
	}
	defer client.Close()

	client.Listen()

	//ch := make(chan os.Signal)
	//signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}
