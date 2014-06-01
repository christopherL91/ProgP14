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
)

type server struct {
	mutex           *sync.Mutex
	money           map[uint16]uint16    //cardnumber -> money
	connections     map[*net.Conn]bool   //connected users.
	loggedInclients map[*net.Conn]uint16 //connection -> cardnumber.
	FictionalUsers  map[uint16]uint16

	//Channels that can access the datastructures. Handeld by the function banker.
	balanceCh  chan Transaction
	withdrawCh chan Transaction
	depositCh  chan Transaction
}

type Transaction struct {
	Money uint16
	ID    uint16
	Done  func(error, *Protocol.Message)
}

/*---------------------------------------------------*/

//			Server area
/*---------------------------------------------------*/

func (s *server) getCardNumber(conn *net.Conn) uint16 {
	s.mutex.Lock()
	s.mutex.Unlock()
	return s.loggedInclients[conn]
}

func (s *server) addConnection(conn *net.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.connections[conn] = true
}

func (s *server) removeConnection(conn *net.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.connections, conn)
}

func init() {
	//For configurations.
	flag.StringVar(&configPath, "config", "server.gcfg", "Path to config file")
	flag.Parse()                         //Parse the actual string.
	runtime.GOMAXPROCS(runtime.NumCPU()) //Use maximal number of cores.
}

func main() {
	/*---------------------------------------------------*/
	c := make(chan os.Signal)      //A channel to listen on keyboard events.
	signal.Notify(c, os.Interrupt) //If user pressed CTRL - C.
	config := new(Config)          //new config struct.
	server := &server{
		mutex: new(sync.Mutex),

		//A fake vault.
		money: map[uint16]uint16{
			1234: 100,
			1123: 500,
		},
		connections:     make(map[*net.Conn]bool),   //a map of current connections.
		loggedInclients: make(map[*net.Conn]uint16), //A list of connected users.
		balanceCh:       make(chan Transaction, numberOfMessages),
		withdrawCh:      make(chan Transaction, numberOfMessages),
		depositCh:       make(chan Transaction, numberOfMessages),
		//A list of fictional users, for demonstration purposes.
		FictionalUsers: map[uint16]uint16{
			1234: 12,
			1123: 13,
		},
	}
	/* handels os.Interrupt, and closes every connected client,
	and then closes the program.*/
	go server.cleanUp(c)
	go server.banker(server.balanceCh, server.withdrawCh, server.depositCh)

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
		server.addConnection(&conn)
		go server.connectionHandler(&conn) //connection handler for every new connection.
	}
}

func trimToLine(str string) string {
	if len(str) > 80 {
		return str[:80]
	} else {
		return str
	}
}

//See slice tricks on google for a reference.
func insert(arr []string, str string, i int) []string {
	return append(arr[:i], append([]string{str}, arr[i:]...)...)
}

//Handle every new connection here.
func (s *server) connectionHandler(conn *net.Conn) {
	write := make(chan *Protocol.Message, numberOfMessages)
	read := make(chan *Protocol.Message, numberOfMessages)
	color.Printf("@{c}New Client connected with IP %s\n", (*conn).RemoteAddr().String())
	encoder := gob.NewEncoder(*conn)

	menuconfig := Protocol.MenuConfig{}
	data, err := ioutil.ReadFile("menus.json")
	checkError(err)
	err = json.Unmarshal(data, &menuconfig.Menus)
	checkError(err)
	for key, lang := range menuconfig.Menus {
		i := 1
		banner := trimToLine(lang.Banner)
		lang.Login = insert(lang.Login, banner, i)
		lang.Menu = insert(lang.Menu, banner, i)
		menuconfig.Menus[key] = lang
	}

	err = encoder.Encode(menuconfig)
	checkError(err)
	s.readWrite(conn, read, write)
	color.Printf("@{c}Client with IP disconnected %s\n", (*conn).RemoteAddr().String())
	s.removeConnection(conn)
}

/*	This function acts as an banker in the virtual bank. This function is the only one that can
	access the important datastructures.
*/
func (s *server) banker(balanceCh, withdrawCh, depositCh chan Transaction) {
	for {
		select {
		case trans := <-balanceCh:
			trans.Done(nil, &Protocol.Message{
				Code:    Protocol.ResponseOK,
				Payload: s.money[trans.ID],
			})
		case trans := <-withdrawCh:
			cardNumber := trans.ID
			currentBalance := s.money[trans.ID]
			requested := trans.Money
			if currentBalance < requested {
				trans.Done(errors.New("Negative value"), &Protocol.Message{
					Code: Protocol.ResponseMoneyProblem,
				})
			} else {
				newBalance := currentBalance - requested
				s.money[cardNumber] = newBalance
				trans.Done(nil, &Protocol.Message{
					Code:    Protocol.ResponseOK,
					Payload: newBalance,
				})
			}
		case trans := <-depositCh:
			cardNumber := trans.ID
			currentBalance := s.money[cardNumber]
			requested := trans.Money
			newBalance := currentBalance + requested
			s.money[cardNumber] = newBalance
			trans.Done(nil, &Protocol.Message{
				Code:    Protocol.ResponseOK,
				Payload: newBalance,
			})
		}
	}
}

//Set the login state of client.
func (s *server) setLogin(conn *net.Conn, number uint16) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loggedInclients[conn] = number
}

//check if user is logged in.
func (s *server) isLoggedIn(conn *net.Conn) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.loggedInclients[conn] != 0 {
		return true
	} else {
		return false
	}
	return false
}

func (s *server) isAccepted(card, pass uint16) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.FictionalUsers[card]
	if !ok {
		return false //could not find user in map.
	}
	if s.FictionalUsers[card] == pass { //match.
		return true
	}
	return false //user exists, but pass does not match.
}

func readMessages(decoder *codec.Decoder, read chan *Protocol.Message, errChan chan error, conn *net.Conn) {
	for {
		message := new(Protocol.Message)
		//If you haven't heard anything from user in 15 min, close the connection.
		(*conn).SetReadDeadline(time.Now().Add(15 * time.Minute))
		err := decoder.Decode(message)
		if err != nil {
			opErr, ok := err.(*net.OpError)
			if ok && opErr.Timeout() {
				errChan <- errors.New("Client connection timeout.")
				break
			} else {
				errChan <- err
				break
			}
		}
		read <- message //send back the message to readWrite that you just got in.
	}
}

func (s *server) readWrite(conn *net.Conn, write, read chan *Protocol.Message) {
	decoder := codec.NewDecoder(*conn, &mh)       //encode messages.
	encoder := codec.NewEncoder(*conn, &mh)       //decode messages.
	errChan := make(chan error)                   //error channel. If some error occurs, a function will handle it.
	var err error                                 //a convenience variable.
	go readMessages(decoder, read, errChan, conn) //launch goroutine to listen on messages.
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

			switch message.Code {
			case Protocol.LoginStatusCode: //login
				color.Printf("@{g}User with IP %s are trying to login\n", ip)
				//is the user already logged in?
				if s.isLoggedIn(conn) {
					write <- &Protocol.Message{
						Code: Protocol.ResponseAlreadyLoggedIn,
					}
				} else {
					//is user an accepted user?
					if s.isAccepted(message.Number, message.Payload) {
						color.Println("@{g}Sending granted message...")
						write <- &Protocol.Message{
							Code: Protocol.ResponseOK,
						}
						s.setLogin(conn, message.Number)
						color.Printf("@{g}Successfully sent granted message to user with IP %s\n", ip)
					} else {
						write <- &Protocol.Message{
							Code: Protocol.ResponseNotAccepted,
						}
						color.Printf("@{g}Client with IP %s tried to log in with wrong credentials\n", ip)
					}
				}
			case 2: //logout
				color.Printf("@{g}User with IP %s are trying to logout\n", ip)
				//is the user already logged out?
				if !s.isLoggedIn(conn) {
					write <- &Protocol.Message{
						Code: Protocol.ResponseAlreadyLoggedIn,
					}
				} else {
					s.setLogin(conn, message.Number)
				}
			case 3: //deposit
				s.depositCh <- Transaction{
					Money: message.Payload,
					ID:    s.getCardNumber(conn),
					Done: func(err error, message *Protocol.Message) {
						if err != nil {
							println(err.Error())
						}
						write <- message
					},
				}
			case 4: //withdraw
				s.withdrawCh <- Transaction{
					Money: message.Payload,
					ID:    s.getCardNumber(conn),
					Done: func(err error, message *Protocol.Message) {
						if err != nil {
							println(err.Error())
						}
						write <- message
					},
				}
			case 5: //balance
				s.balanceCh <- Transaction{
					ID: s.getCardNumber(conn),
					Done: func(err error, message *Protocol.Message) {
						if err != nil {
							println(err.Error())
						}
						write <- message
					},
				}
			default:
				color.Println("@{r}Server received unknown code")
			}
		case err = <-errChan: //error chan.
			break Outer
		}
	}
}

func (s *server) cleanUp(c chan os.Signal) {
	<-c
	color.Println("@{c}\nClosing every client connection...")
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for conn, _ := range s.connections {
		if conn == nil {
			continue
		}
		err := (*conn).Close()
		if err != nil {
			continue
		}
	}
	color.Println("@{r}\nServer is now closing...")
	os.Exit(1)
}

func checkError(err error) {
	if err != nil {
		color.Printf("@{r}Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
