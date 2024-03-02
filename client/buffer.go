package client

import "github.com/smallnest/ringbuffer"

type Buffer struct {
	buffer *ringbuffer.RingBuffer
}

func NewBuffer(size int) Buffer {
	return Buffer{buffer: ringbuffer.New(size)}
}

func (b Buffer) Write(p []byte) (n int, err error) {
	return b.buffer.Write(p)
}

func (b Buffer) Read(p []byte) (n int, err error) {
	return b.buffer.Read(p)
}

func (b Buffer) Close() error {
	b.buffer.Reset()
	return nil
}

func (b Buffer) Free() int {
	return b.buffer.Free()
}

func (b Buffer) Length() int {
	return b.buffer.Length()
}
