package transport

import (
	"fmt"
	"io"

	"github.com/hayride-dev/bindgen/gen/go/wasi/io/streams"
)

type wasiReadCloser struct {
	stream streams.InputStream
}

// create an io.Reader from the input stream
func NewReader(s streams.InputStream) io.Reader {
	return &wasiReadCloser{
		stream: s,
	}
}

func NewReadCloser(s streams.InputStream) io.ReadCloser {
	return &wasiReadCloser{
		stream: s,
	}
}

func (w *wasiReadCloser) Read(p []byte) (int, error) {
	readResult := w.stream.Read(uint64(len(p)))
	if readResult.IsErr() {
		readErr := readResult.Err()
		if readErr.Closed() {
			return 0, io.EOF
		}
		return 0, fmt.Errorf("failed to read from InputStream %s", readErr.LastOperationFailed().ToDebugString())
	}

	readList := readResult.OK()
	copy(p, readList.Slice())
	return int(readList.Len()), nil
}

func (w *wasiReadCloser) Close() error {
	w.stream.ResourceDrop()
	return nil
}
