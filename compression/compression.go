package compression

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

type (
	// Compression is the state represents if compressed or not.
	Compression int
)

const (
	// Uncompressed represents the uncompressed.
	Uncompressed Compression = iota
	// Gzip is gzip compression algorithm.
	Gzip
)

// DetectCompression detects the compression algorithm of the source.
func DetectCompression(source []byte) Compression {
	for compression, m := range map[Compression][]byte{
		Gzip: {0x1F, 0x8B, 0x08},
	} {
		if len(source) < len(m) {
			// Len too short
			continue
		}
		if bytes.Compare(m, source[:len(m)]) == 0 {
			return compression
		}
	}
	return Uncompressed
}

// DecompressStream decompresses the archive and returns a ReaderCloser with the decompressed archive.
func DecompressStream(archive io.Reader) (io.ReadCloser, error) {
	p := bufioReader32KPool
	buf := p.Get().(*bufio.Reader)
	buf.Reset(archive)
	bs, err := buf.Peek(10)
	if err != nil && err != io.EOF {
		// Note: we'll ignore any io.EOF error because there are some odd
		// cases where the layer.tar file will be empty (zero bytes) and
		// that results in an io.EOF from the Peek() call. So, in those
		// cases we'll just treat it as a non-compressed stream and
		// that means just create an empty layer.
		// See Issue 18170
		return nil, err
	}

	compression := DetectCompression(bs)
	switch compression {
	case Uncompressed:
		readBufWrapper := &poolReadCloserWrapper{p, buf, buf}
		return readBufWrapper, nil
	case Gzip:
		gzReader, err := gzip.NewReader(buf)
		if err != nil {
			return nil, err
		}
		readBufWrapper := &poolReadCloserWrapper{p, buf, gzReader}
		return readBufWrapper, nil
	default:
		return nil, fmt.Errorf("unsupported compression format %s", (&compression).Extension())
	}
}

// CompressStream compresseses the dest with specified compression algorithm.
func CompressStream(dest io.Writer, compression Compression) (io.WriteCloser, error) {
	p := bufioWriter32KPool
	buf := p.Get().(*bufio.Writer)
	buf.Reset(dest)
	switch compression {
	case Uncompressed:
		writeBufWrapper := &poolWriteCloserWrapper{p, buf, buf}
		return writeBufWrapper, nil
	case Gzip:
		gzWriter := gzip.NewWriter(dest)
		writeBufWrapper := &poolWriteCloserWrapper{p, buf, gzWriter}
		return writeBufWrapper, nil
	default:
		return nil, fmt.Errorf("unsupported compression format %s", (&compression).Extension())
	}
}

// Extension returns the extension of a file that uses the specified compression algorithm.
func (compression *Compression) Extension() string {
	switch *compression {
	case Gzip:
		return "gz"
	}
	return ""
}
