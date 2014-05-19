//Client to the ATM server.

/*
	TODO:
	Read configurations from file.
	Make contact with server.
*/

package main

import (
	"code.google.com/p/gcfg"
	"flag"
	"fmt"
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
	Config
	conn net.Conn
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
		fmt.Fprintln(os.Stderr, "Bye")
		os.Exit(1) //will just quit client if user pressed CTRL - C
	}() //Execute goroutine
}

func main() {
	//create a new client.
	client := new(client)
}
