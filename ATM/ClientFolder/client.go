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
