package client

type MessageHeader struct {
	msgType      uint16
	id           uint16
	refersTo     uint16
	sentSec      int32
	sentUsec     int32
	receivedSec  int32
	receivedUsec int32
	size         uint32
}

type HelloMessage struct {
	size    uint32
	payload []byte
}

// type ServerSettingsMessage struct {
// 	size    uint32
// 	payload []byte
// }

// type CodecHeaderMessage struct {
// 	codecSize uint32
// 	codec     []byte
// 	size      uint32
// 	payload   []byte
// }

type Hello struct {
	Arch                      string `json:"Arch"`
	ClientName                string `json:"ClientName"`
	HostName                  string `json:"HostName"`
	ID                        string `json:"ID"`
	Instance                  int    `json:"Instance"`
	MAC                       string `json:"MAC"`
	OS                        string `json:"OS"`
	SnapStreamProtocolVersion int    `json:"SnapStreamProtocolVersion"`
	Version                   string `json:"Version"`
}

type ServerSettings struct {
	BufferMS int  `json:"bufferMS"`
	Latency  int  `json:"latency"`
	Muted    bool `json:"muted"`
	Volume   int  `json:"volume"`
}

type CodecHeader struct {
	Codec   []byte
	Payload []byte
}
