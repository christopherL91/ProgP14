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

/*
	TODO:
	Remove clients when they disconnect.
*/

package main

import (
	"code.google.com/p/gcfg"
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"github.com/christopherL91/Protocol"
	"github.com/ugorji/go/codec"
	"github.com/wsxiaoys/terminal/color"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

//			Config
/*---------------------------------------------------*/

//Struct to hold all the configurations.
type Config struct {
	Server struct {
		Address string
		Port    string
	}
}

var (
	configPath       string
	mh               codec.MsgpackHandle //MessagePack
	numberOfMessages = 10
	users            = map[uint16]uint16{
		1234: 12,
		1123: 13,
	}
)

type server struct {
	mutex           *sync.Mutex
	money           map[uint16]uint16    //cardnumber -> money
	connections     map[*net.Conn]bool   //connected users.
	loggedInclients map[*net.Conn]uint16 //connection -> cardnumber.
	balanceCh       chan *Protocol.Message
	withdrawCh      chan *Protocol.Message
	depositCh       chan *Protocol.Message
}

/*---------------------------------------------------*/

//			Server area
/*---------------------------------------------------*/

//Handle every new connection here.
func (s *server) connectionHandler(conn *net.Conn) {
	defer (*conn).Close()
	write := make(chan Protocol.Message, numberOfMessages)
	read := make(chan Protocol.Message, numberOfMessages)
	color.Printf("@{c}New Client connected with IP %s\n", (*conn).RemoteAddr().String())
	encoder := gob.NewEncoder(*conn)

	menuconfig := Protocol.MenuConfig{}
	data, err := ioutil.ReadFile("menus.json")
	checkError(err)
	err = json.Unmarshal(data, &menuconfig.Menus)
	checkError(err)

	err = encoder.Encode(menuconfig)
	checkError(err)
	s.readWrite(conn, read, write)
	color.Printf("@{c}Client with IP disconnected %s\n", (*conn).RemoteAddr().String())
}

func readMessages(decoder *codec.Decoder, read chan Protocol.Message, errChan chan error, conn *net.Conn) {
	for {
		message := new(Protocol.Message)
		(*conn).SetReadDeadline(time.Now().Add(15 * time.Minute))
		err := decoder.Decode(message)
		opErr, ok := err.(*net.OpError)
		if ok && opErr.Timeout() {
			errChan <- errors.New("Client connection timeout.")
			break
		}
		if err != nil {
			errChan <- err
			break
		}
		read <- *message
	}

}

func (s *server) banker(balanceCh, withdrawCh, depositCh chan *Protocol.Message) {
	for {
		select {
		case client := <-balanceCh:
			balanceCh <- &Protocol.Message{
				Payload: s.money[client.Number],
			}
		case client := <-withdrawCh:
			cardNumber := client.Number
			currentBalance := s.money[cardNumber]
			requested := client.Payload
			if currentBalance < requested || requested < 0 {
				balanceCh <- &Protocol.Message{
					Payload: 0, //operation failed.
				}
			} else {
				s.money[cardNumber] = currentBalance - requested
				balanceCh <- &Protocol.Message{
					Payload: 1, //operation succed.
				}
			}
		case client := <-depositCh:
			cardNumber := client.Number
			currentBalance := s.money[cardNumber]
			requested := client.Payload
			s.money[cardNumber] = currentBalance + requested
			balanceCh <- &Protocol.Message{
				Payload: 1,
			}
		}
	}
}

func (s *server) setLogin(state bool, conn *net.Conn, number uint16) {
	s.mutex.Lock()
	s.loggedInclients[conn] = number
	s.mutex.Unlock()
}

func (s *server) isAccepted(card, pass uint16) bool {
	_, ok := users[card]
	if !ok {
		return false //could not find user in map.
	}
	if users[card] == pass {
		return true
	}
	return false
}

func (s *server) readWrite(conn *net.Conn, write, read chan Protocol.Message) {
	decoder := codec.NewDecoder(*conn, &mh)
	encoder := codec.NewEncoder(*conn, &mh)
	errChan := make(chan error)
	var err error
	go readMessages(decoder, read, errChan, conn)
Outer:
	for {
		select {
		case message := <-write: //write messages.
			err = encoder.Encode(message)
			if err != nil {
				color.Printf("@{r}%s", err.Error())
				break Outer
			}
		case message := <-read:
			ip := (*conn).RemoteAddr().String()
			if !message.LoggedIn {
				color.Printf("@{g}User with IP %s are trying to login\n", ip)
				if s.isAccepted(message.Number, message.Payload) {
					color.Println("@{g}Sending granted message...")
					write <- Protocol.Message{
						LoggedIn: true,
					}
					color.Printf("@{g}Successfully sent granted message to user with IP %s\n", ip)
				} else {
					write <- Protocol.Message{
						LoggedIn: false,
					}
					color.Printf("@{g}Client with IP %s tried to log in with wrong credentials\n", ip)
				}
			}
		case err = <-errChan: //error chan.
			break Outer
		}
	}
}

func init() {
	//For configurations.
	flag.StringVar(&configPath, "config", "server.gcfg", "Path to config file")
	flag.Parse()                         //Parse the actual string.
	runtime.GOMAXPROCS(runtime.NumCPU()) //Use maximal number of cores.
}

func (s *server) addConnection(conn *net.Conn) {
	s.mutex.Lock()
	s.connections[conn] = true
	s.mutex.Unlock()
}

func (s *server) cleanUp(c chan os.Signal) {
	<-c
	color.Println("@{c}\nClosing every client connection...")
	s.mutex.Lock()
	for conn, _ := range s.connections {
		if conn == nil {
			continue
		}
		err := (*conn).Close()
		if err != nil {
			continue
		}
	}
	s.mutex.Unlock()
	color.Println("@{r}\nServer is now closing...")
	os.Exit(1)
}

func main() {
	/*---------------------------------------------------*/
	c := make(chan os.Signal)      //A channel to listen on keyboard events.
	signal.Notify(c, os.Interrupt) //If user pressed CTRL - C.
	config := new(Config)          //new config struct.
	server := &server{
		mutex:           new(sync.Mutex),
		money:           make(map[uint16]uint16),
		connections:     make(map[*net.Conn]bool),
		loggedInclients: make(map[*net.Conn]uint16),
		balanceCh:       make(chan *Protocol.Message, numberOfMessages),
		withdrawCh:      make(chan *Protocol.Message, numberOfMessages),
		depositCh:       make(chan *Protocol.Message),
	}
	go server.cleanUp(c)

	var address string                           //holds the address to the server.
	var port string                              //holds the port to the server.
	color.Println("\t\t\t\t@{b}ATM started")     //Print out with colors.
	color.Println("@{g}Reading config file...")  //Print out with colors.
	err := gcfg.ReadFileInto(config, configPath) //Read config file.
	checkError(err)
	color.Println("@{g}Config read OK")
	address = config.Server.Address
	port = config.Server.Port
	address += ":" + port
	/*---------------------------------------------------*/

	listener, err := net.Listen("tcp", address)
	checkError(err)
	color.Printf("@{g}Listening on %s\n\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			color.Printf("@{r}%s", err.Error())
			continue
		}
		server.addConnection(&conn)        //add connections.
		go server.connectionHandler(&conn) //connection handler for every new connection.
	}
}

func checkError(err error) {
	if err != nil {
		color.Printf("@{r}Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
