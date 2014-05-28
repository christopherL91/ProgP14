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
	"encoding/json"
	"flag"
	"fmt"
	"github.com/christopherL91/Protocol"
	"github.com/wsxiaoys/terminal/color"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	// "strings"
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

var configPath string

/*---------------------------------------------------*/

//			Server area
/*---------------------------------------------------*/

//Handle every new connection here.
func connectionHandler(conn net.Conn) {
	//read menu and pass it to the client.
	defer conn.Close()
	color.Printf("@{c}New Client connected with IP %s\n", conn.RemoteAddr().String())

	menuconfig := Protocol.MenuConfig{}
	data, err := ioutil.ReadFile("menus.json")
	checkError(err)
	err = json.Unmarshal(data, &menuconfig.Menus)
	checkError(err)

	// fmt.Println(strings.Join(menus["swedish"].Menu, "\n"))
	err = gob.NewEncoder(conn).Encode(menuconfig)
	checkError(err)

	go listen(conn)

	for {
		select {
		// case data := <-listenCh:
		// 	//data came in. Do something about it.
		}
	} //blocking.
}

func listen(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	var buffer string
	for scanner.Scan() {
		buffer = scanner.Text()
		fmt.Println(buffer)
	}

	if err := scanner.Err(); err != nil {
		conn.Close()
		checkError(err)
	}
}

func init() {
	//For configurations.
	flag.StringVar(&configPath, "config", "server.gcfg", "Path to config file")
	flag.Parse()                         //Parse the actual string.
	runtime.GOMAXPROCS(runtime.NumCPU()) //Use maximal number of cores.
}

func main() {
	/*---------------------------------------------------*/
	config := new(Config)                        //new config struct.
	var address string                           //holds the address to the server.
	var port string                              //holds the port to the server.
	color.Println("\t\t\t\t@{b}ATM started")     //Print out with colors.
	color.Println("@{g}Reading config file...")  //Print out with colors.
	err := gcfg.ReadFileInto(config, configPath) //Read config file.
	checkError(err)
	color.Println("@{g}Config read OK")
	address = config.Server.Address
	port = config.Server.Port
	/*---------------------------------------------------*/

	address += ":" + port
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	color.Printf("@{g}Listening on %s\n\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			//Don't let one bad connection bring you down.
			continue
		}
		go connectionHandler(conn) //connection handler for every new connection.
	}
}

//Convinience function.
func checkError(err error) {
	if err != nil {
		color.Printf("@{r}Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
