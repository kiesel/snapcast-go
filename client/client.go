package client

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
)

type stats struct {
	Codec   string
	Packets int
	Bytes   int
}

var buffer Buffer = NewBuffer(1024 * 1024) // 1 MiB
var streamer beep.StreamSeekCloser
var format beep.Format
var codecHeader *CodecHeader
var statistics stats

func (c *Client) PrintStatistics(loop time.Duration) {
	for {
		fmt.Printf("Recv [%d] packets, [%d] bytes, %q, buf %0.2f%% filled\n",
			statistics.Packets, statistics.Bytes, statistics.Codec, float32(buffer.Length())/float32(buffer.Free()+buffer.Length()),
		)

		time.Sleep(loop)
	}
}

func (c *Client) Play() {
	err := c.Dial()
	if err != nil {
		panic(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	hello := Hello{
		Arch:       runtime.GOARCH,
		ClientName: "kiesel/snapcast-go",
		HostName:   hostname,
		OS:         runtime.GOOS,
		Version:    "0.0.1",
		ID:         strconv.Itoa(os.Getpid()),
		Instance:   1,
	}

	err = c.SendHello(&hello)
	if err != nil {
		panic(err)
	}

	for {
		message, err := c.ReadMessage()
		if err != nil {
			panic(err)
		}

		statistics.Packets++

		switch t := message.(type) {
		case *CodecHeader:
			fmt.Println("New codec:", message)
			statistics.Bytes += len(message.(*CodecHeader).Payload)

			handleCodecHeader(message.(*CodecHeader))

		case *ServerSettings:
			fmt.Println("New server settings:", message)

		case *WireChunk:
			wirechunk := message.(*WireChunk)
			buffer.Write(wirechunk.Payload)
			statistics.Bytes += len(wirechunk.Payload)

			if streamer == nil && codecHeader != nil && buffer.Length() > 10240 /* 10kb */ {
				initCodec()
			}

		default:
			fmt.Printf("Got unhandled type %T\n", t)
		}
	}
}

func handleCodecHeader(message *CodecHeader) {
	codecHeader = message
	buffer.Write(message.Payload)
}

func initCodec() {
	var err error

	codec := string(codecHeader.Codec)
	fmt.Printf("Attempting to init %q codec with [%d] bytes ...\n", codec, buffer.Length())
	statistics.Codec = codec

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
