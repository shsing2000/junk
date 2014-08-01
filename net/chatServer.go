package main

import (
	"bufio"
	"log"
	"net"
)

type Client struct {
	id       string
	incoming chan string
	outgoing chan string
	quit     chan *Client
	reader   *bufio.Reader
	writer   *bufio.Writer
}

func (c *Client) Read() {
	for {
		line, _ := c.reader.ReadString('\n')
		if line == "quit\n" {
			c.quit <- c
			return
		}

		c.incoming <- line
	}
}

func (c *Client) Write() {
	for data := range c.outgoing {
		c.writer.WriteString(data)
		c.writer.Flush()
	}
}

func (c *Client) Listen() {
	go c.Read()
	go c.Write()
}

func (c *Client) Close() {
	close(c.incoming)
	close(c.outgoing)
}

func NewClient(c net.Conn, quit chan *Client) *Client {
	writer := bufio.NewWriter(c)
	reader := bufio.NewReader(c)

	client := &Client{
		id:       c.RemoteAddr().String(),
		incoming: make(chan string),
		outgoing: make(chan string),
		quit:     quit,
		reader:   reader,
		writer:   writer,
	}

	client.Listen()
	return client
}

type ChatRoom struct {
	clients  []*Client
	joins    chan net.Conn
	quits    chan *Client
	incoming chan string
	outgoing chan string
}

func (c *ChatRoom) Broadcast(data string) {
	log.Printf("received data: %s", data)
	for _, client := range c.clients {
		client.outgoing <- data
	}
}

func (c *ChatRoom) Join(conn net.Conn) {
	log.Printf("Creating client: %s\n", conn.RemoteAddr())
	client := NewClient(conn, c.quits)
	client.outgoing <- "Welcome\n"
	c.clients = append(c.clients, client)
	go func() {
		for {
			c.incoming <- <-client.incoming
		}
	}()
}

func (cr *ChatRoom) remove(client *Client) {
	//cut the client from the slice
	for i, c := range cr.clients {
		if c.id == client.id {
			c.Close()
			log.Printf("Client %s has left the room", c.id)

			copy(cr.clients[i:], cr.clients[i+1:])
			cr.clients = cr.clients[:len(cr.clients)-1]
			return
		}
	}
}

func (c *ChatRoom) Listen() {
	go func() {
		for {
			select {
			case data := <-c.incoming:
				c.Broadcast(data)
			case conn := <-c.joins:
				c.Join(conn)
			case client := <-c.quits:
				c.remove(client)
			}
		}
	}()
}

func NewChatRoom() *ChatRoom {
	chatRoom := &ChatRoom{
		clients:  make([]*Client, 0),
		joins:    make(chan net.Conn),
		quits:    make(chan *Client),
		incoming: make(chan string),
		outgoing: make(chan string),
	}

	chatRoom.Listen()
	return chatRoom
}

func main() {
	chatRoom := NewChatRoom()

	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	log.Printf("Created Server on: %s\n", l.Addr())
	for {
		conn, _ := l.Accept()
		chatRoom.joins <- conn
	}
}
