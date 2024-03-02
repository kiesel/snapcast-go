package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
	"github.com/kiesel/snapcast/client"
)

var serverUrl string

var buffer client.Buffer = client.NewBuffer(1024 * 1024) // 1 MiB
var streamer beep.StreamSeekCloser
var format beep.Format
var codecHeader *client.CodecHeader

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
			buffer.Write(wirechunk.Payload)

			if streamer == nil && codecHeader != nil && buffer.Length() > 10240 /* 10kb */ {
				initCodec()
			}

		default:
			fmt.Printf("Got type %T\n", t)
		}
	}
}

func handleCodecHeader(message *client.CodecHeader) {
	codecHeader = message
	buffer.Write(message.Payload)
}

func initCodec() {
	var err error

	codec := string(codecHeader.Codec)
	fmt.Printf("Attempting to init %q codec with [%d] bytes ...\n", codec, buffer.Length())
	switch codec {
	case "pcm":
		streamer, format, err = wav.Decode(buffer)

	case "flac":
		streamer, format, err = flac.Decode(buffer)

	case "ogg":
		streamer, format, err = vorbis.Decode(buffer)

	default:
		panic(fmt.Errorf("unsupported codec %s", codec))
	}

	if err != nil {
		panic(err)
	}

	bufferSize := format.SampleRate.N(time.Second / 10)
	fmt.Printf("Init player, sample rate [%d], buffer size [%d]\n", format.SampleRate, bufferSize)
	err = speaker.Init(format.SampleRate, bufferSize)
	if err != nil {
		panic(err)
	}
	speaker.Play(streamer)
	codecHeader = nil

	fmt.Println("Sucessfully initialized player")
}
