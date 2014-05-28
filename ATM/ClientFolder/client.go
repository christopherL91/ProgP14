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
	"bufio"
	"code.google.com/p/gcfg"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/christopherL91/Protocol"
	"github.com/wsxiaoys/terminal/color"
	"net"
	"os"
	"os/signal"
	"strings"
)

//Configuration stuff.
var (
	configPath string
	prompt             = "Unicorn@ATM> "
	version    float32 = 1.5
	author             = "Christopher Lillthors. Unicorn INC"
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

func (c *client) listen(conn net.Conn, input chan string) {
	color.Println("@{g}Downloading config files...")
	var counter int //to increment the menu options.
	menuconfig := new(Protocol.MenuConfig)
	err := gob.NewDecoder(conn).Decode(menuconfig)
	checkError(err)
	color.Println("@{g}Config files downloaded\n")

	color.Println("\t\t\t\t@{b}Choose language")

	//print out the different languages you can choose on the screen.
	for language, _ := range menuconfig.Menus {
		counter += 1
		color.Printf("@{g} %d) %s\n", counter, language)
	}

	//User chooses languages.
	for {
		fmt.Print(prompt)
		menu, ok := menuconfig.Menus[<-input]
		if !ok {
			color.Println("@{r}Invalid input. Please try again")
		} else {
			fmt.Println(strings.Join(menu.Menu, "\n"))
			break
		}
	}

	//User makes more options.
Outer:
	for {
		fmt.Print(prompt)
		switch <-input {
		case "1":
			break Outer //Break outer for loop.
		case "2":
			color.Printf("@{b}Version:%f\nAuthor:%s\n", version, author)
		default:
			color.Println("@{r}Invalid input. Please try again")
		}
	}
}

func init() {
	//For configurations.
	flag.StringVar(&configPath, "config", "client.gcfg", "Path to config file")
	flag.Parse()
}

func main() {
	//For UNIX signal handling.
	c := make(chan os.Signal, 1)   //A channel to listen on keyboard events.
	signal.Notify(c, os.Interrupt) //If user pressed CTRL - C.

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

	//A goroutine to check for keyboard events.
	go func() {
		<-c              //blocking.
		defer os.Exit(1) //will just quit client if user pressed CTRL - C
		client.conn.Close()
		fmt.Fprintln(os.Stderr, "\nThank you for using a ATM from Unicorn INC")
	}() //Execute goroutine
	inputCh := make(chan string)
	reader := bufio.NewReader(os.Stdin)

	go client.listen(client.conn, inputCh)

	for {
		// fmt.Print(prompt)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		inputCh <- line
	}
}

//Convinience function.
func checkError(err error) {
	if err != nil {
		color.Printf("@{r}Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
