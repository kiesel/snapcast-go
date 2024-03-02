package main

import (
	"flag"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/kiesel/snapcast/client"
)

var serverUrl string
var channel chan bool

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
	go snapClient.PrintStatistics(2 * time.Second)
	go snapClient.Play()
	<-channel
}
