// The MIT License (MIT)

// Copyright (c) 2014 Christopher Lillthors

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"code.google.com/p/gcfg"
	"flag"
	"github.com/wsxiaoys/terminal/color"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

//			Config
/*---------------------------------------------------*/
//Configuration stuff.
var (
	configPath string
	prompt     = "Unicorn@ATM>"
	version    = 1.0
	author     = "Christopher Lillthors. Unicorn INC"
	width      = 80
)

//Struct to hold all the configurations.
type Config struct {
	Server struct {
		Address string
		Port    string
	}
}

/*---------------------------------------------------*/

//			Server/Client
/*---------------------------------------------------*/
//A convenience type.
type menu []string

type client struct {
	conn net.Conn
	id   int
}

type server struct {
	clients map[*client]bool //For fast checkup if the client is connected.
	mutex   *sync.Mutex
	menus   map[string]menu
}

/*---------------------------------------------------*/

//			Configure for your own good.
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

//			Server area
/*---------------------------------------------------*/
func newServer() *server {
	return &server{
		clients: make(map[*client]bool),
		mutex:   new(sync.Mutex),
		menus:   make(map[string]menu),
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

func (s *server) addClient(c *client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clients[c] = true
}

/*---------------------------------------------------*/

//Convinience function.
func checkError(err error) {
	if err != nil {
		color.Printf("@{r}Fatal error %s", err.Error())
		os.Exit(1)
	}
}

func init() {
	//For configurations.
	flag.StringVar(&configPath, "config", "server.gcfg", "Path to config file")
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
}

func main() {
	/*---------------------------------------------------*/
	config := new(Config)
	color.Println("\t\t\t\t@{b}ATM started")
	color.Println("@{g}Reading config file...")
	err := gcfg.ReadFileInto(config, configPath)
	checkError(err)
	color.Println("@{g}Config read OK")
	/*---------------------------------------------------*/

	/*---------------------------------------------------*/
	/*Setup area. create a new server*/
	server := newServer()
	server.addMenu("Swedish", swedish)
	server.addMenu("English", english)
	/*---------------------------------------------------*/
}
