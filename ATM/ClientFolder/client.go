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
	Read configurations from file.
	Make contact with server.
*/

package main

import (
	"code.google.com/p/gcfg"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/wsxiaoys/terminal/color"
	"net"
	"os"
	"os/signal"
)

//Configuration stuff.
var (
	configPath string
	prompt     = "Unicorn@ATM>"
	version    = 1.0
	author     = "Christopher Lillthors. Unicorn INC"
)

//Struct to hold all the configurations.
type Config struct {
	Client struct {
		Address string
		Port    string
	}
}

type client struct {
	*Config
	conn *net.TCPConn
}

//Struct to hold an actual message beetween client and server.
type Message struct {
	Banner string
	Body   string
	Type   string
}

//Convinience function.
func checkError(err error) {
	if err != nil {
		color.Printf("@{r}Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func init() {
	//For configurations.
	flag.StringVar(&configPath, "config", "client.gcfg", "Path to config file")

	//For UNIX signal handling.
	c := make(chan os.Signal, 1)   //A channel to listen on keyboard events.
	signal.Notify(c, os.Interrupt) //If user pressed CTRL - C.

	//A goroutine to check for keyboard events.
	go func() {
		<-c //blocking.
		//inform server that I will quit.
		fmt.Fprintln(os.Stderr, "\nThank you for using a ATM from Unicorn INC")
		os.Exit(1) //will just quit client if user pressed CTRL - C
	}() //Execute goroutine
}

func main() {
	//			Config area.
	/*---------------------------------------------------*/
	client := &client{Config: new(Config)}
	var address string //holds the address to the server.
	var port string    //holds the port to the server.

	err := gcfg.ReadFileInto(client.Config, configPath)
	checkError(err)
	address = client.Config.Client.Address
	port = client.Config.Client.Port
	address += ":" + port
	/*---------------------------------------------------*/

	//			Connection area
	/*---------------------------------------------------*/
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	checkError(err)
	client.conn, err = net.DialTCP("tcp4", nil, tcpAddr)
	checkError(err)
	// encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(client.conn)
	message := new(Message)
	decoder.Decode(message) //blocking while wating for message.
	color.Printf("@{b}Message: %s\n", message)
}
