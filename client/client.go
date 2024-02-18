package client

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
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

func (client *Client) SendHello(helloPayload *Hello) error {
	helloPayloadString, err := json.Marshal(*helloPayload)
	if err != nil {
		return err
	}

	hello := HelloMessage{
		payload: helloPayloadString,
		size:    uint32(len(helloPayloadString)),
	}

	err = client.send(MessageHeader{
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

func (client *Client) ReceiveServerSettings() (*ServerSettings, error) {
	var payload *ServerSettings
	conn := *client.Conn

	reader := bufio.NewReaderSize(conn, BASE_PACKET_LENGTH)
	base, err := client.readBase(reader)

	if err != nil {
		panic(err)
	}

	if base.msgType != MSG_TYPE_SERVER_SETTINGS {
		return payload, fmt.Errorf("expected message type SERVER_SETTINGS (%d), received %d", MSG_TYPE_SERVER_SETTINGS, base.msgType)

	}

	reader = bufio.NewReaderSize(conn, int(base.size))
	buffer, err := readDynamicLengthBytes(reader)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(*buffer, &payload)
	return payload, err
}

func (client *Client) ReceiveCodecHeader() (*CodecHeader, error) {
	var payload *CodecHeader
	conn := *client.Conn

	reader := bufio.NewReaderSize(conn, BASE_PACKET_LENGTH)
	base, err := client.readBase(reader)

	if err != nil {
		panic(err)
	}

	if base.msgType != MSG_TYPE_CODEC_HEADER {
		return payload, fmt.Errorf("expected message type CODEC_HEADER (%d), received %d", MSG_TYPE_CODEC_HEADER, base.msgType)
	}

	reader = bufio.NewReaderSize(conn, int(base.size))

	message := CodecHeader{}

	buffer, err := readDynamicLengthBytes(reader)
	if err != nil {
		panic(err)
	}
	message.Codec = *buffer

	buffer, err = readDynamicLengthBytes(reader)
	if err != nil {
		panic(err)
	}
	message.Payload = *buffer

	return &message, nil
}

func (client *Client) readBase(reader io.Reader) (MessageHeader, error) {
	base := MessageHeader{}
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

func (client *Client) send(base MessageHeader) error {
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

func readDynamicLengthBytes(r io.Reader) (*[]byte, error) {
	var size uint32
	err := readLE(r, []any{&size})
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, size)
	_, err = io.ReadFull(r, buffer)

	return &buffer, err
}
