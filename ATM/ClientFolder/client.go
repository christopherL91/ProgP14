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
	"errors"
	"flag"
	"fmt"
	"github.com/christopherL91/Protocol"
	"github.com/ugorji/go/codec"
	"github.com/wsxiaoys/terminal/color"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
)

//Configuration stuff.
var (
	configPath     string
	prompt         = "Unicorn@ATM> "
	version        = 1.5
	author         = "Christopher Lillthors. Unicorn INC"
	mh             codec.MsgpackHandle                  //MessagePack
	cardnumberTest = regexp.MustCompile(`^([0-9]){4}$`) //4 digits.
	passnumberTest = regexp.MustCompile(`^([0-9]){2}$`) //2 digits.
)

//Struct to hold all the configurations.
type Config struct {
	Client struct {
		Address string
		Port    string
	}
}

func listen(conn net.Conn, input chan string) {
	var counter int     //to increment the menu options.
	var language string //string to hold the language that the user choosed.
	decoder := gob.NewDecoder(conn)
	menuconfig := new(Protocol.MenuConfig)

	color.Println("@{g}Downloading config files...")
	err := decoder.Decode(menuconfig) //
	checkError(err)
	color.Println("@{g}Config files downloaded\n")
	decoder = nil
	color.Println("\t\t\t\t@{b}Choose language")

	writeCh := make(chan Protocol.Message) //send messages.
	readCh := make(chan Protocol.Message)  //read messages.
	go listenMessage(readCh, conn)         // function to listen to server.
	go writeMessage(writeCh, conn)         //function to write to server.

	//print out the different languages you can choose on the screen.
	for language, _ := range menuconfig.Menus {
		counter += 1
		color.Printf("@{g} %d) %s\n", counter, language)
	}

	//  1) Swedish
	//  2) English

	//User chooses languages.
	for {
		fmt.Print(prompt)
		language = <-input
		menu, ok := menuconfig.Menus[language]
		if !ok {
			color.Printf("@{r}%s\n", "Invalid input")
		} else {
			fmt.Println(strings.Join(menu.Menu, "\n"))
			break
		}
	}

	// ".................................................",
	// 		"UNICORN INC",
	// 		"1) Log in",
	// 		"2) Contact us",
	// 		"................................................."
K:
	for {
		fmt.Print(prompt)
		switch <-input {
		case "1":
			//user choosed "log in" Do something about it.
			err := login(input, writeCh, readCh) //handle login from user.
			if err != nil {
				conn.Close()
				color.Printf("@{r}%s", err.Error())
			}
			fmt.Println(strings.Join(menuconfig.Menus[language].Login, "\n")) //print out the rest of the menu.
			break K                                                           //Break outer for loop.
		case "2":
			color.Printf("@{b}Version:%f\nAuthor:%s\n", version, author)
		default:
			color.Printf("@{r}%s\n", menuconfig.Menus[language].Invalid)
		}
	}

	// 	"1) Withdraw",
	// 	"2) Input",
	// 	"3) Balance"
L:
	for {
		fmt.Print(prompt)
		switch <-input {
		case "1":
			color.Println("@{b}Input $")
			break L
		case "2":
			color.Println("@{b}Input $")
			break L
		case "3":
			color.Println("@{b}Checking balance...")
			break L
		default:
			color.Printf("@{r}%s\n", menuconfig.Menus[language].Invalid)
		}
	}
}

func listenMessage(readCh chan Protocol.Message, conn net.Conn) {
	message := new(Protocol.Message)
	decoder := codec.NewDecoder(conn, &mh)
	for {
		err := decoder.Decode(message)
		if err != nil {
			break
		}
		//something came in.
		readCh <- *message
	}
}

//write messages to the server.
func writeMessage(write chan Protocol.Message, conn net.Conn) {
	encoder := codec.NewEncoder(conn, &mh)
	for {
		select {
		case message := <-write:
			err := encoder.Encode(message)
			if err != nil {
				break
			}
		}
	}
}

//input chan is for keyboard input.
func login(input chan string, writeCh, readCh chan Protocol.Message) error {
	var cardNum, passNum string
	for {
		color.Println("@{b}Input cardnumber.")
		cardNum = <-input
		color.Println("@{b}Input password.")
		passNum = <-input
		if cardnumberTest.MatchString(cardNum) && passnumberTest.MatchString(passNum) {
			break
		} else {
			color.Println("@{r}Invalid credentials. Please try again.")
		}
	}

	card, _ := strconv.Atoi(cardNum)
	pass, _ := strconv.Atoi(passNum)

	login := Protocol.Message{
		Number:   uint16(card),
		Pass:     uint16(pass),
		LoggedIn: false,
	}
	writeCh <- login   //send message from server.
	answer := <-readCh //read answer from server.
	if answer.LoggedIn {
		color.Println("@{gB}\nYou were granted access")
		return nil
	} else {
		return errors.New("Something happened. Please restart session")
	}
}

func init() {
	//For configurations.
	flag.StringVar(&configPath, "config", "client.gcfg", "Path to config file")
	flag.Parse()
}

func main() {
	//			Config area.
	/*---------------------------------------------------*/
	config := new(Config)
	var address string //holds the address to the server.
	var port string    //holds the port to the server.
	err := gcfg.ReadFileInto(config, configPath)
	checkError(err)
	address = config.Client.Address
	port = config.Client.Port
	address += ":" + port
	/*---------------------------------------------------*/

	//			Connection area
	/*---------------------------------------------------*/

	conn, err := net.Dial("tcp", address) //connect to server.
	checkError(err)

	//For UNIX signal handling.
	c := make(chan os.Signal)      //A channel to listen on keyboard events.
	signal.Notify(c, os.Interrupt) //If user pressed CTRL - C.
	go cleanUp(c, conn)

	inputCh := make(chan string)
	go listen(conn, inputCh) //listen on keyboard events.

	reader := bufio.NewReader(os.Stdin)
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		inputCh <- line //listen on keyboard.
	}
}

func cleanUp(c chan os.Signal, conn net.Conn) {
	<-c
	conn.Close() //close connection.
	fmt.Fprintln(os.Stderr, "\nThank you for using a ATM from Unicorn INC")
	os.Exit(1)
}

//Convinience function.
func checkError(err error) {
	if err != nil {
		color.Printf("@{r}Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
