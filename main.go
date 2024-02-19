package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
	"github.com/kiesel/snapcast/client"
)

var serverUrl string

var streamer beep.Streamer
var buffer bytes.Buffer
var writer io.Writer = bufio.NewWriterSize(&buffer, 102400)
var reader = bufio.NewReaderSize(&buffer, 1024)
var lastCodec *client.CodecHeader

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
			handleCodecHeader(message.(*client.CodecHeader))

		case *client.ServerSettings:
			fmt.Println("New server settings:", message)

		case *client.WireChunk:
			wirechunk := message.(*client.WireChunk)
			writer.Write(wirechunk.Payload)
			lateInitCodec()

		default:
			fmt.Printf("Got type %T\n", t)
		}
	}
}

func handleCodecHeader(message *client.CodecHeader) {
	lastCodec = message
	writer.Write(message.Payload)
}

func lateInitCodec() {
	if lastCodec == nil {
		return
	}

	codec := string(lastCodec.Codec)
	switch codec {
	case "pcm":
		fmt.Printf("Attempting to init PCM codec with %d buffered bytes ...\n", buffer.Len())
		streamer, format, err := wav.Decode(reader)
		if err != nil {
			fmt.Println(err)
			return
		}
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second))
		if err != nil {
			panic(err)
		}
		speaker.Play(streamer)
		lastCodec = nil
		return

	case "flac":
		fmt.Printf("%d | %d", buffer.Available(), buffer.Cap())
		if buffer.Len() < (102400 / 2) {
			return
		}
		fmt.Printf("Attempting to init FLAC codec with %d buffered bytes ...\n", buffer.Len())
		streamer, format, err := flac.Decode(reader)
		if err != nil {
			fmt.Println(err)
			return
		}
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		if err != nil {
			panic(err)
		}
		speaker.Play(streamer)
		lastCodec = nil
		return

	default:
		panic(fmt.Errorf("unsupported codec %s", codec))
	}
}
