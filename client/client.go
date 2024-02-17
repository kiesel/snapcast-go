package client

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
)

const (
	MSG_TYPE_BASE            = uint16(0)
	MSG_TYPE_CODEC_HEADER    = uint16(1)
	MSG_TYPE_WIRE_CHUNK      = uint16(2)
	MSG_TYPE_SERVER_SETTINGS = uint16(3)
	MSG_TYPE_TIME            = uint16(4)
	MSG_TYPE_HELLO           = uint16(5)
	MSG_TYPE_STREAM_TAGS     = uint16(6)
	BASE_PACKET_LENGTH       = 26 /* size of all base struct members in bytes */
)

type Client struct {
	Host string
	Port int
	Conn *net.Conn
}

func (client *Client) Dial() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.Host, client.Port))
	if err != nil {
		return err
	}

	client.Conn = &conn
	return nil
}

func (client *Client) SendHello() error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	helloPayload := HelloPayload{
		Arch:       runtime.GOARCH,
		ClientName: "kiesel/snapcast-go",
		HostName:   hostname,
		OS:         runtime.GOOS,
		Version:    "0.0.1",
		ID:         strconv.Itoa(os.Getpid()),
		Instance:   1,
	}
	helloPayloadString, err := json.Marshal(helloPayload)
	if err != nil {
		return err
	}

	hello := Hello{
		payload: helloPayloadString,
		size:    uint32(len(helloPayloadString)),
	}

	err = client.send(Base{
		msgType: MSG_TYPE_HELLO,
		size:    hello.size + 4,
	})

	if err != nil {
		return err
	}

	writer := bufio.NewWriterSize(*client.Conn, int(hello.size))

	err = writeLE(writer, []any{hello.size, hello.payload})
	if err != nil {
		return err
	}

	return writer.Flush()
}

func (client *Client) ReceiveServerSettings() (*ServerSettingsPayload, error) {
	var payload *ServerSettingsPayload
	conn := *client.Conn

	reader := bufio.NewReaderSize(conn, BASE_PACKET_LENGTH)
	base, err := client.readBase(reader)

	if err != nil {
		panic(err)
	}

	if base.msgType != MSG_TYPE_SERVER_SETTINGS {
		return payload, fmt.Errorf("expected message type SERVER_SETTINGS, received %d", base.msgType)
	}

	reader = bufio.NewReaderSize(conn, int(base.size))

	serverSettings := ServerSettings{}
	err = readLE(reader, []any{&serverSettings.size})
	if err != nil {
		panic(err)
	}

	serverSettings.payload = make([]byte, serverSettings.size)
	_, err = io.ReadFull(reader, serverSettings.payload)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(serverSettings.payload, &payload)
	return payload, err
}

func (client *Client) readBase(reader io.Reader) (Base, error) {
	base := Base{}
	err := readLE(reader, []any{
		&base.msgType,
		&base.id,
		&base.refersTo,
		&base.sentSec,
		&base.sentUsec,
		&base.receivedSec,
		&base.receivedUsec,
		&base.size,
	})
	return base, err
}

func (client *Client) send(base Base) error {
	writer := bufio.NewWriterSize(*client.Conn, BASE_PACKET_LENGTH)

	err := writeLE(writer, []any{
		base.msgType,
		base.id,
		base.refersTo,
		base.sentSec,
		base.sentUsec,
		base.receivedSec,
		base.receivedUsec,
		base.size,
	})

	if err != nil {
		return err
	}

	return writer.Flush()
}

func writeLE(w io.Writer, data []any) error {
	for _, v := range data {
		err := binary.Write(w, binary.LittleEndian, v)

		if err != nil {
			return err
		}
	}

	return nil
}

func readLE(r io.Reader, data []any) error {
	for _, v := range data {
		err := binary.Read(r, binary.LittleEndian, v)

		if err != nil {
			return err
		}
	}

	return nil
}
