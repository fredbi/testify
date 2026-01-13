package difflib

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

const bufferSize = 4096

var (
	writersPool = &poolOfWriters{
		Pool: sync.Pool{
			New: func() any {
				return bufio.NewWriterSize(&defaultBuf, bufferSize)
			},
		},
	}

	buffersPool = &poolOfBuffers{
		Pool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}

	defaultBuf bytes.Buffer
)

type poolOfWriters struct {
	sync.Pool
}

func (p *poolOfWriters) BorrowWithWriter(w io.Writer) *bufio.Writer {
	raw := p.Get()
	buf := raw.(*bufio.Writer)
	buf.Reset(w)

	return buf
}

func (p *poolOfWriters) Redeem(buf *bufio.Writer) {
	p.Put(buf)
}

type poolOfBuffers struct {
	sync.Pool
}

func (p *poolOfBuffers) Borrow() *bytes.Buffer {
	raw := p.Get()
	buf := raw.(*bytes.Buffer)
	buf.Reset()

	return buf
}

func (p *poolOfBuffers) Redeem(buf *bytes.Buffer) {
	p.Put(buf)
}
