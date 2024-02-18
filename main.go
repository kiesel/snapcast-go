package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strconv"

	"github.com/kiesel/snapcast/client"
)

var serverUrl string

func init() {
	flag.StringVar(&serverUrl, "url", "tcp://localhost:1704", "server address")
	flag.Parse()
}

func main() {
	fmt.Println("Connecting to", serverUrl)

	parsed, err := url.Parse(serverUrl)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		port = 1704

	}

	snapClient := client.Client{Host: parsed.Hostname(), Port: port}
	err = snapClient.Dial()
	if err != nil {
		panic(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	hello := client.Hello{
		Arch:       runtime.GOARCH,
		ClientName: "kiesel/snapcast-go",
		HostName:   hostname,
		OS:         runtime.GOOS,
		Version:    "0.0.1",
		ID:         strconv.Itoa(os.Getpid()),
		Instance:   1,
	}

	err = snapClient.SendHello(&hello)
	if err != nil {
		panic(err)
	}

	for {
		message, err := snapClient.ReadMessage()
		if err != nil {
			panic(err)
		}

		switch t := message.(type) {
		case *client.CodecHeader:
			fmt.Println("New codec:", message)
		case *client.ServerSettings:
			fmt.Println("New server settings:", message)
		default:
			fmt.Printf("Got type %T\n", t)
		}
	}
}
