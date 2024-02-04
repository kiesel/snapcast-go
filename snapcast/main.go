package main 

import "fmt"
import "flag"
import "net/url"
import "github.com/kiesel/snapcast-go/client"

var serverUrl string

func init() {
	flag.StringVar(&serverUrl, "url", "tcp://localhost:1234", "server address")
	flag.Parse()
}

func main() {
	fmt.Println("Connecting to", serverUrl)

	parsed, err := url.Parse(serverUrl)
	if (err != nil) {
		panic(err)
	}

	client := Client{ parsed.Host, parsed.Port }
	err = client.Dial()
	if (err != nil) {
		panic(err)
	}
}