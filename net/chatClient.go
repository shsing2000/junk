package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"regexp"
)

type client struct {
	id   string
	send chan string
	recv chan string
	quit chan struct{}
	conn net.Conn
}

func newClient(address string) (*client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	c := &client{
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

func (c *client) write(msg string) {
	c.conn.Write([]byte(msg))
}

func (c *client) close() {
	close(c.send)
	close(c.recv)
	close(c.quit)

	c.conn.Close()
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

		dst <- s
	}
}

var quitCmd = regexp.MustCompile("^quit\r?\n$")

func main() {

	c, err := newClient("127.0.0.1:8888")
	if err != nil {
		log.Fatal("could not create client")
	}
	defer c.close()

	log.Printf("connected to server: %s\n", c.conn.LocalAddr().String())

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
			return
		}
	}

	/*l, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	log.Println("logged in to chatroom")

	go func(r io.Reader) {
		reader := bufio.NewReader(r)

		for {
			data, _ := reader.ReadString('\n')
			log.Println(data)
		}
	}(l)

	r := bufio.NewReader(os.Stdin)
	for {
		msg, _ := r.ReadString('\n')

		if quitCmd.MatchString(msg) {
			l.Write([]byte("quit\n"))
			log.Println("exiting chatroom")
			return
		}
		l.Write([]byte(msg))
	}*/

}

/*
// An uninteresting service.
type Service struct {
	ch        chan bool
	waitGroup *sync.WaitGroup
}

// Make a new Service.
func NewService() *Service {
	s := &Service{
		ch:        make(chan bool),
		waitGroup: &sync.WaitGroup{},
	}
	s.waitGroup.Add(1)
	return s
}

// Accept connections and spawn a goroutine to serve each one.  Stop listening
// if anything is received on the service's channel.
func (s *Service) Serve(listener *net.TCPListener) {
	defer s.waitGroup.Done()
	for {
		select {
		case <-s.ch:
			log.Println("stopping listening on", listener.Addr())
			listener.Close()
			return
		default:
		}
		listener.SetDeadline(time.Now().Add(1e9))
		conn, err := listener.AcceptTCP()
		if nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println(err)
		}
		log.Println(conn.RemoteAddr(), "connected")
		s.waitGroup.Add(1)
		go s.serve(conn)
	}
}

// Stop the service by closing the service's channel.  Block until the service
// is really stopped.
func (s *Service) Stop() {
	close(s.ch)
	s.waitGroup.Wait()
}

// Serve a connection by reading and writing what was read.  That's right, this
// is an echo service.  Stop reading and writing if anything is received on the
// service's channel but only after writing what was read.
func (s *Service) serve(conn *net.TCPConn) {
	defer conn.Close()
	defer s.waitGroup.Done()
	for {
		select {
		case <-s.ch:
			log.Println("disconnecting", conn.RemoteAddr())
			return
		default:
		}
		conn.SetDeadline(time.Now().Add(1e9))
		buf := make([]byte, 4096)
		if _, err := conn.Read(buf); nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println(err)
			return
		}
		if _, err := conn.Write(buf); nil != err {
			log.Println(err)
			return
		}
	}
}

func main() {

	// Listen on 127.0.0.1:48879.  That's my favorite port number because in
	// hex 48879 is 0xBEEF.
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:48879")
	if nil != err {
		log.Fatalln(err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if nil != err {
		log.Fatalln(err)
	}
	log.Println("listening on", listener.Addr())

	// Make a new service and send it into the background.
	service := NewService()
	go service.Serve(listener)

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	// Stop the service gracefully.
	service.Stop()

}

*/
