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
	"strconv"
	"strings"
	"time"
)

//Configuration stuff.
var (
	configPath string
	prompt     = "Unicorn@ATM> "
	version    = 1.5
	author     = "Christopher Lillthors. Unicorn INC"
	address    = "Unicorn road 1337"
	mh         codec.MsgpackHandle //MessagePack
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
	go readMessage(readCh, conn)            // function to listen to server.
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
			fmt.Println()
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
	for {
		fmt.Print(prompt)
		switch <-input {
		case "1":
			err := bankHandler(Protocol.WithdrawStatusCode, input, writeCh, readCh)
			if err != nil {
				color.Printf("@{r}%s\n", err.Error())
				color.Println("@{r}Please try again")
				continue
			}
		case "2":
			err := bankHandler(Protocol.DepositStatusCode, input, writeCh, readCh)
			if err != nil {
				color.Printf("@{r}%s\n", err.Error())
				color.Println("@{r}Please try again")
				continue
			}
		case "3":
			err := bankHandler(Protocol.BalanceStatusCode, input, writeCh, readCh)
			if err != nil {
				color.Printf("@{r}%s\n", err.Error())
				color.Println("@{r}Please try again")
				continue
			}
		default:
			color.Printf("@{r}%s\n", menuconfig.Menus[language].Invalid)
		}
	}
}

func readMessage(readCh chan *Protocol.Message, conn *net.Conn) {
	decoder := codec.NewDecoder(*conn, &mh)
	for {
		message := new(Protocol.Message)
		err := decoder.Decode(message)
		if err == io.EOF {
			color.Printf("@{r}Server closed connection")
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

func bankHandler(code uint16, input chan string, writeCh, readCh chan *Protocol.Message) error {
	switch code {
	case Protocol.BalanceStatusCode:
		color.Println("@{b}Contacting bank... Please wait.")
		message := &Protocol.Message{
			Code: Protocol.BalanceStatusCode,
		}
		writeCh <- message
		response := <-readCh
		color.Printf("@{g}Your balance: %d\n", response.Payload)
	case Protocol.WithdrawStatusCode:
		color.Println("@{b}Please input your scratch code")
		fmt.Print(prompt)
		numString := <-input
		if Protocol.ScratchTest.MatchString(numString) {
			num, _ := strconv.Atoi(numString)
			if num%2 == 1 {
				color.Println("@{b}Please input the amount of money that you want to widthdraw.")
				fmt.Print(prompt)
				moneyString := <-input
				if Protocol.MoneyTest.MatchString(moneyString) {
					money, _ := strconv.Atoi(moneyString)
					message := &Protocol.Message{
						Code:    Protocol.WithdrawStatusCode,
						Payload: uint16(money),
					}
					color.Println("@{b}Contacting bank... Please wait.")
					writeCh <- message
					response := <-readCh
					if response.Code == Protocol.ResponseOK {
						color.Println("@{b}Status : @{g}OK")
					} else if response.Code == Protocol.ResponseMoneyProblem {
						color.Println("@{b}Status : @{r}Error:The amount of money was bigger than your balance.")
					}
				} else {
					return errors.New("Not a number!")
				}
			} else {
				return errors.New("Invalid scratch code.")
			}
		} else {
			return errors.New("Not a number!")
		}
	case Protocol.DepositStatusCode:
		color.Println("@{b}Please input the amount of money that you want to deposit.")
		fmt.Print(prompt)
		inputString := <-input
		if Protocol.MoneyTest.MatchString(inputString) {
			money, _ := strconv.Atoi(inputString)
			message := &Protocol.Message{
				Code:    Protocol.DepositStatusCode,
				Payload: uint16(money),
			}
			color.Println("@{b}Contacting bank... Please wait.")
			writeCh <- message
			<-readCh //wait for the response.
			color.Println("@{b}Status : @{g}OK")
		} else {
			return errors.New("Not a number!")
		}
	default:
		return errors.New("Unkown code.")
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
		if Protocol.CardnumberTest.MatchString(cardNum) && Protocol.PassnumberTest.MatchString(passNum) {
			break
		} else {
			color.Println("@{r}\nInvalid credentials. Please try again.")
		}
	}

	card, _ := strconv.Atoi(cardNum)
	pass, _ := strconv.Atoi(passNum)

	message := &Protocol.Message{
		Code:    Protocol.LoginStatusCode,
		Number:  uint16(card),
		Payload: uint16(pass),
	}
	writeCh <- message   //send message from server.
	response := <-readCh //read answer from server.
	if response.Code == Protocol.ResponseNotAccepted {
		return errors.New("You were not accepted")
	} else if response.Code == Protocol.ResponseAlreadyLoggedIn {
		return errors.New("Already logged in")
	}
	color.Printf("@{gB}You were granted access.")
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
		color.Printf("@{r}Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}
