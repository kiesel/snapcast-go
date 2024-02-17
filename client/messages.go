package client

type Base struct {
	msgType      uint16
	id           uint16
	refersTo     uint16
	sentSec      int32
	sentUsec     int32
	receivedSec  int32
	receivedUsec int32
	size         uint32
}

type Hello struct {
	Base
	size    uint32
	payload []byte
}

type HelloPayload struct {
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
	size    uint32
	payload []byte
}

type ServerSettingsPayload struct {
	BufferMS int  `json:"bufferMS"`
	Latency  int  `json:"latency"`
	Muted    bool `json:"muted"`
	Volume   int  `json:"volume"`
}
