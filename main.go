package main

import (
	"flag"
	"fmt"
	"net/url"
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

	client := client.Client{Host: parsed.Hostname(), Port: port}
	err = client.Dial()
	if err != nil {
		panic(err)
	}

	err = client.SendHello()
	if err != nil {
		panic(err)
	}

	server, err := client.ReceiveServerSettings()
	if err != nil {
		panic(err)
	}

	fmt.Println(server)
}
