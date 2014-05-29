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
	"flag"
	"fmt"
	"github.com/christopherL91/Protocol"
	"github.com/ugorji/go/codec"
	"github.com/wsxiaoys/terminal/color"
	"io/ioutil"
	"net"
	"os"
	"runtime"
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
	configPath string
	mh         codec.MsgpackHandle //MessagePack
)

/*---------------------------------------------------*/

//			Server area
/*---------------------------------------------------*/

//Handle every new connection here.
func connectionHandler(conn net.Conn) {
	//read menu and pass it to the client.
	color.Printf("@{c}New Client connected with IP %s\n", conn.RemoteAddr().String())
	encoder := gob.NewEncoder(conn)

	menuconfig := Protocol.MenuConfig{}
	data, err := ioutil.ReadFile("menus.json")
	checkError(err)
	err = json.Unmarshal(data, &menuconfig.Menus)
	checkError(err)

	err = encoder.Encode(menuconfig)
	checkError(err)
	encoder = nil

	listen(conn) //blocking.
	color.Printf("@{c}Client with IP %s disconnected\n", conn.RemoteAddr().String())
}

func listen(conn net.Conn) {
	message := new(Protocol.Login)
	decoder := codec.NewDecoder(conn, &mh)
	decoder.Decode(message)

	fmt.Println(message.Number)
	fmt.Println(message.Pass)

	// scanner := bufio.NewScanner(conn)
	// var buffer []byte
	// for scanner.Scan() {
	// 	buffer = scanner.Bytes()
	// 	fmt.Println(buffer)
	// }

	// if err := scanner.Err(); err != nil {
	// 	conn.Close()
	// 	checkError(err)
	// }
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
	address += ":" + port
	/*---------------------------------------------------*/

	listener, err := net.Listen("tcp", address)
	checkError(err)
	color.Printf("@{g}Listening on %s\n\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go connectionHandler(conn) //connection handler for every new connection.
	}
}

func checkError(err error) {
	if err != nil {
		color.Printf("@{r}Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
