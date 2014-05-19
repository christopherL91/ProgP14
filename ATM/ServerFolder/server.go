package main

//ATM server.

import (
	// "fmt"
	"code.google.com/p/gcfg"
	"net"
	"sync"
	"time"
	"runtime"
)

//A convenience type.
type menu []string

type client struct {
	conn    net.Conn
	id int
	
}

type server struct {
	clients []client
	mutex   *sync.Mutex
	menus   map[string]menu
}

/*---------------------------------------------------*/
/*Everything here will be printed out to the client.*/
var (
	banner = "UNICORN INC\n" //Corporate banner.

	//A menu in swedish
	swedish = menu{
		".................................................", //Starting of menu
		banner,
		"Time " + time.Now().String() + "\n",
		"1) Logga in",
		"2) Kontakta oss",
		".................................................", //End of menu
	}

	//A menu in english.
	english = menu{
		".................................................", //Starting of menu
		banner,
		"Time " + time.Now().String() + "\n",
		"1) Log in",
		"2) Contact us",
		".................................................", //End of menu
	}
)

/*---------------------------------------------------*/

const (
	//Use this to check how long the banner is.
	width = 80
)

func newServer() *server {
	return &server{
		clients: make([]client,1)
		mutex: new(sync.Mutex),
		menus: make(map[string]menu, 2),
	}
}

func (s *server) addMenu(name string, m menu) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.menus[name] = m
}

func (s *server) numOfMenus() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return len(s.menus)
}

func (s *server) addClient() {
	
}
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	/*---------------------------------------------------*/
	/*Setup area. create a new server*/
	server := newServer()
	server.addMenu("Swedish", swedish)
	server.addMenu("English", english)
	/*---------------------------------------------------*/
}
