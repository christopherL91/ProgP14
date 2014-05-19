//Client to the ATM server.

/*
	TODO:

	Implement a menu for the client.
*/

package main

import (
	"fmt"
	"os"
	"os/signal"
)

const (
	prompt  = "Unicorn@ATM>"
	version = 1.0
	author  = "Christopher Lillthors. Unicorn INC"
)

func init() {
	c := make(chan os.Signal, 1)   //A channel to listen on keyboard events.
	signal.Notify(c, os.Interrupt) //If user pressed CTRL - C.

	//A goroutine to check for keyboard events.
	go func() {
		<-c //blocking.
		//inform server that I will quit.
		fmt.Fprintln(os.Stderr, "Bye")
		os.Exit(1)
	}()
}

func main() {
}
