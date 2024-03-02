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
	Host           string
	Port           int
	Conn           *net.Conn
	ServerSettings *ServerSettings
	CodecHeader    *CodecHeader
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

func (client *Client) readServerSettings(header MessageHeader) (*ServerSettings, error) {
	reader := bufio.NewReaderSize(*client.Conn, int(header.size))
	buffer, err := readDynamicLengthBytes(reader)
	if err != nil {
		panic(err)
	}

	var payload *ServerSettings
	err = json.Unmarshal(*buffer, &payload)
	return payload, err
}

func (client *Client) readCodecHeader(header MessageHeader) (*CodecHeader, error) {
	reader := bufio.NewReaderSize(*client.Conn, int(header.size))
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

func (client *Client) readWireChunk(header MessageHeader) (*WireChunk, error) {
	chunk := WireChunk{}

	r := bufio.NewReaderSize(*client.Conn, int(header.size))

	err := readLE(r, []any{&chunk.Sec, &chunk.Usec})
	if err != nil {
		panic(err)
	}

	buffer, err := readDynamicLengthBytes(r)
	if err != nil {
		panic(err)
	}

	chunk.Payload = *buffer
	return &chunk, nil
}

func (client *Client) ReadMessage() (any, error) {
	header, err := client.readHeader()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Received [type %d], [id %d], [refers %d] length %d\n",
		header.msgType,
		header.id,
		header.refersTo,
		header.size,
	)

	switch header.msgType {
	case MSG_TYPE_CODEC_HEADER:
		codec, err := client.readCodecHeader(header)
		if err != nil {
			return nil, err
		}
		client.CodecHeader = codec
		return codec, nil

	case MSG_TYPE_SERVER_SETTINGS:
		settings, err := client.readServerSettings(header)
		if err != nil {
			return nil, err
		}
		client.ServerSettings = settings
		return settings, nil

	case MSG_TYPE_WIRE_CHUNK:
		return client.readWireChunk(header)

	default:
		fmt.Printf("Reading message type %d not implemented - discarding %d bytes\n", header.msgType, header.size)

		buffer := make([]byte, header.size)
		_, err = (*client.Conn).Read(buffer)

		if err != nil {
			panic(err)
		}
	}

	return nil, nil
}

func (client *Client) readHeader() (MessageHeader, error) {
	reader := bufio.NewReaderSize(*client.Conn, BASE_PACKET_LENGTH)

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
