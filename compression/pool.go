package compression

import (
	"bufio"
	"io"
	"sync"
)

var (
	bufioReader32KPool = &sync.Pool{
		New: func() interface{} { return bufio.NewReaderSize(nil, 32*1024) },
	}
	bufioWriter32KPool = &sync.Pool{
		New: func() interface{} { return bufio.NewWriterSize(nil, 32*1024) },
	}
)

type poolReadCloserWrapper struct {
	p   *sync.Pool
	buf *bufio.Reader
	r   io.Reader
}

func (wrapper *poolReadCloserWrapper) Read(p []byte) (n int, err error) {
	return wrapper.r.Read(p)
}

func (wrapper *poolReadCloserWrapper) Close() error {
	if readCloser, ok := wrapper.r.(io.ReadCloser); ok {
		readCloser.Close()
	}
	wrapper.buf.Reset(nil)
	wrapper.p.Put(wrapper.buf)
	return nil
}

type poolWriteCloserWrapper struct {
	p   *sync.Pool
	buf *bufio.Writer
	w   io.Writer
}

func (wrapper *poolWriteCloserWrapper) Write(p []byte) (n int, err error) {
	return wrapper.w.Write(p)
}

func (wrapper *poolWriteCloserWrapper) Close() error {
	wrapper.buf.Flush()
	if writeCloser, ok := wrapper.w.(io.WriteCloser); ok {
		writeCloser.Close()
	}
	wrapper.buf.Reset(nil)
	wrapper.p.Put(wrapper.buf)
	return nil
}
