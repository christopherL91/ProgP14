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
	"io"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//Configuration stuff.
var (
	configPath     string
	prompt         = "Unicorn@ATM> "
	version        = 1.5
	author         = "Christopher Lillthors. Unicorn INC"
	address        = "Unicorn road 1337"
	mh             codec.MsgpackHandle               //MessagePack
	cardnumberTest = regexp.MustCompile(`^(\d){4}$`) //4 digits.
	passnumberTest = regexp.MustCompile(`^(\d){2}$`) //2 digits.
	moneyTest      = regexp.MustCompile(`^(\d+)$`)   //at least one digit.
)

const (
	loginStatusCode    = 1
	LogoutStatusCode   = 2
	DepositStatusCode  = 3
	WithdrawStatusCode = 4
	BalanceStatusCode  = 5

	responseOK              = 0
	responseAlreadyLoggedIn = 1
	responseNotAccepted     = 2
	responseMoneyProblem    = 3
)

//Struct to hold all the configurations.
type Config struct {
	Client struct {
		Address string
		Port    string
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
	//Server has 30 seconds to respond.
	conn, err := net.DialTimeout("tcp", address, 30*time.Second)
	checkError(err)

	//For UNIX signal handling.
	c := make(chan os.Signal)      //A channel to listen on keyboard events.
	signal.Notify(c, os.Interrupt) //If user pressed CTRL - C.
	go cleanUp(c, &conn)

	inputCh := make(chan string)
	go listen(&conn, inputCh) //listen on keyboard events.

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
func listen(conn *net.Conn, input chan string) {
	var counter int     //to increment the menu options.
	var language string //string to hold the language that the user choosed.
	decoder := gob.NewDecoder(*conn)
	menuconfig := new(Protocol.MenuConfig)

	color.Println("@{g}Downloading config files...")
	err := decoder.Decode(menuconfig) //
	checkError(err)
	color.Println("@{g}Config files downloaded\n")
	color.Println("\t\t\t\t@{b}Choose language")

	writeCh := make(chan *Protocol.Message) //send messages.
	readCh := make(chan *Protocol.Message)  //read messages.
	go listenMessage(readCh, conn)          // function to listen to server.
	go writeMessage(writeCh, conn)          //function to write to server.

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
				//Login failed.
				color.Printf("@{r}%s\n", err.Error())
				color.Println("@{r}Please try again")
				continue
			}
			fmt.Println(strings.Join(menuconfig.Menus[language].Login, "\n")) //print out the rest of the menu.
			break K                                                           //Break outer for loop.
		case "2":
			color.Printf("@{b}Version:%f\nAuthor:%s\nAddress:%s\n", version, author, address)
		default:
			color.Printf("@{r}%s\n", menuconfig.Menus[language].Invalid)
		}
	}

	// 	"1) Withdraw",
	// 	"2) Deposit",
	// 	"3) Balance"
L:
	for {
		fmt.Print(prompt)
		switch <-input {
		case "1":
			err := bankHandler(WithdrawStatusCode, input, writeCh, readCh)
			if err != nil {
				color.Printf("@{r}%s\n", err.Error())
				color.Println("@{r}Please try again")
				continue
			}
			color.Println("@{gB}Success")
			break L
		case "2":
			err := bankHandler(DepositStatusCode, input, writeCh, readCh)
			if err != nil {
				color.Printf("@{r}%s\n", err.Error())
				color.Println("@{r}Please try again")
				continue
			}
			color.Println("@{gB}Success")
			break L
		case "3":
			err := bankHandler(BalanceStatusCode, input, writeCh, readCh)
			if err != nil {
				color.Printf("@{r}%s\n", err.Error())
				color.Println("@{r}Please try again")
				continue
			}
			color.Println("@{gB}Success")
			break L
		default:
			color.Printf("@{r}%s\n", menuconfig.Menus[language].Invalid)
		}
	}
}

func listenMessage(readCh chan *Protocol.Message, conn *net.Conn) {
	decoder := codec.NewDecoder(*conn, &mh)
	for {
		message := new(Protocol.Message)
		err := decoder.Decode(message)
		if err == io.EOF {
			color.Println("@{r}Server closed connection")
			os.Exit(1)
		}
		checkError(err)
		readCh <- message
	}
}

//write messages to the server.
func writeMessage(write chan *Protocol.Message, conn *net.Conn) {
	encoder := codec.NewEncoder(*conn, &mh)
	for {
		select {
		case message := <-write:
			err := encoder.Encode(message)
			if err != nil {
				color.Printf("@{r}%s", err.Error())
				break
			}
		}
	}
}

func bankHandler(code uint8, input chan string, writeCh, readCh chan *Protocol.Message) error {
	switch code {
	case BalanceStatusCode:
		color.Println("@{b}Contacting bank... Please wait")
		message := &Protocol.Message{
			Code: BalanceStatusCode,
		}
		writeCh <- message
		response := <-readCh
		if response.Code == responseOK {
			color.Println("@{gB}\nYour balance:")
		}
	case WithdrawStatusCode:
		fmt.Println("Getting money from bank")
	case DepositStatusCode:
		fmt.Println("Input money to bank")
	default:
		return errors.New("Unkown code")
	}
	return nil
}

//input chan is for keyboard input.
func login(input chan string, writeCh, readCh chan *Protocol.Message) error {
	var cardNum, passNum string
	for {
		color.Println("@{gB}Input cardnumber.")
		fmt.Print(prompt)
		cardNum = <-input
		color.Println("@{gB}Input password.")
		fmt.Print(prompt)
		passNum = <-input
		if cardnumberTest.MatchString(cardNum) && passnumberTest.MatchString(passNum) {
			break
		} else {
			color.Println("@{r}\nInvalid credentials. Please try again.")
		}
	}

	card, _ := strconv.Atoi(cardNum)
	pass, _ := strconv.Atoi(passNum)

	message := &Protocol.Message{
		Code:    loginStatusCode,
		Number:  uint16(card),
		Payload: uint16(pass),
	}
	writeCh <- message //send message from server.
	answer := <-readCh //read answer from server.
	if answer.Code == responseNotAccepted {
		return errors.New("You were not accepted")
	} else if answer.Code == responseAlreadyLoggedIn {
		return errors.New("Already logged in")
	}
	color.Println("@{gB}\nYou were granted access.")
	return nil
}

func cleanUp(c chan os.Signal, conn *net.Conn) {
	<-c
	(*conn).Close() //close connection.
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
