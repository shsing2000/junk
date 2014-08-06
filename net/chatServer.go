package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"regexp"
)

type Client struct {
	id     string
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	recv   chan string
	send   chan string
	quit   chan struct{}
}

func (c *Client) Listen() {
	for {
		select {
		case s := <-c.send:
			c.write(s)
		case <-c.quit:
			c.Close()
			return
		}
	}
	//read server messages and relay to client
	//read client messages and relay to server (server relays to other clients in the room)
}

func (c *Client) write(s string) {
	_, err := c.writer.WriteString(s)
	if err != nil {
		log.Printf("error writing message for %s: %s\n", c.id, s)
		return
	}

	c.writer.Flush()
}

func (c *Client) Close() {
	c.write("server is closing the connection\n")
	c.conn.Close()
}

func NewClient(conn net.Conn) *Client {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	client := &Client{
		id:     conn.RemoteAddr().String(),
		conn:   conn,
		reader: reader,
		writer: writer,
		recv:   make(chan string),
		send:   make(chan string),
		quit:   make(chan struct{}),
	}

	go func(reader io.Reader) {
		r := bufio.NewReader(reader)
		for {
			s, err := r.ReadString('\n')
			if err != nil {
				log.Println("error reading message from client: ", err)
				//if err.WSARecv, return
				continue
			}

			client.recv <- s
		}
	}(conn)

	go client.Listen()

	return client
}

type message struct {
	clientId string
	text     string
}

type ChatRoom struct {
	id          string
	clients     []*Client
	conn        net.Listener
	joins       chan net.Conn
	disconnects chan *Client
	messages    chan message
	quit        chan struct{}
}

func (c *ChatRoom) broadcast(msg message) {
	log.Printf("received data: %s", msg.text)
	for _, client := range c.clients {
		if client.id != msg.clientId {
			client.send <- msg.text
		}
	}
}

func (chat *ChatRoom) join(conn net.Conn) {
	log.Printf("Creating client: %s\n", conn.RemoteAddr())
	client := NewClient(conn)
	client.send <- "Welcome\n"
	chat.clients = append(chat.clients, client)

	go func() {
		for {
			s := <-client.recv
			chat.messages <- message{clientId: client.id, text: s}
		}
	}()
}

func (chat *ChatRoom) remove(client *Client) {
	//cut the client from the slice
	for i, c := range chat.clients {
		if c.id == client.id {
			c.Close()
			log.Printf("Client %s has left the room", c.id)

			copy(chat.clients[i:], chat.clients[i+1:])
			chat.clients = chat.clients[:len(chat.clients)-1]
			return
		}
	}
}

func (c *ChatRoom) Listen() {
	log.Printf("Listening in chatroom: %s\n", c.conn.Addr().String())

	for {
		select {
		case msg := <-c.messages:
			c.broadcast(msg)
		case conn := <-c.joins:
			c.join(conn)
		case client := <-c.disconnects:
			c.remove(client)
		case <-c.quit:
			log.Println("closed chatroom")
			return
		}
	}
}

func NewChatRoom(addr string) (*ChatRoom, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	chatRoom := &ChatRoom{
		id:          listener.Addr().String(),
		conn:        listener,
		clients:     make([]*Client, 0),
		joins:       make(chan net.Conn),
		disconnects: make(chan *Client),
		messages:    make(chan message),
		quit:        make(chan struct{}),
	}

	go func() {
		for {
			client, err := listener.Accept()
			if err != nil {
				log.Println("error connecting client ", err)
			}

			chatRoom.joins <- client
		}
	}()

	go func() {
		r := bufio.NewReader(os.Stdin)
		for {
			s, err := r.ReadString('\n')
			if err != nil {
				log.Println(err)
				return
			}

			if quitCmd.MatchString(s) {
				chatRoom.quit <- struct{}{}
				return
			}

			log.Println("sending message ", s)
			chatRoom.messages <- message{clientId: chatRoom.id, text: s}
		}
	}()

	return chatRoom, nil
}

func (chat *ChatRoom) Close() {
	//issue disconnect to all clients
	for _, client := range chat.clients {
		client.conn.Write([]byte("server is closing the chatroom\nquit\n"))
		client.conn.Close()
	}

	//cleanup the chatroom
	chat.conn.Close()
}

var quitCmd = regexp.MustCompile("^quit\r?\n$")

func main() {
	chatRoom, err := NewChatRoom(":8888")
	if err != nil {
		log.Fatal("could not create chatroom ", err)
	}
	defer chatRoom.Close()

	chatRoom.Listen()
}
