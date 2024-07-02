package transport

import (
	"errors"
	"io"

	"github.com/hayride-dev/wit/gen/go/http/client"
)

type wasiReadCloser struct {
	Stream           client.WasiHttp0_2_0_TypesInputStream
	Body             client.WasiHttp0_2_0_TypesIncomingBody
	OutgoingRequest  client.WasiHttp0_2_0_TypesOutgoingRequest
	IncomingResponse client.WasiHttp0_2_0_TypesIncomingResponse
	Future           client.WasiHttp0_2_0_TypesFutureIncomingResponse
}

func (reader wasiReadCloser) Read(p []byte) (int, error) {
	c := cap(p)
	result := reader.Stream.BlockingRead(uint64(c))
	isEof := result.IsErr() && result.UnwrapErr() == client.WasiIo0_2_0_StreamsStreamErrorClosed()
	if isEof {
		return 0, io.EOF
	} else if result.IsErr() {
		return 0, errors.New("failed to read response body")
	} else {
		chunk := result.Unwrap()
		copy(p, chunk)
		return len(chunk), nil
	}
}

func (reader wasiReadCloser) Close() error {
	reader.Stream.Drop()
	reader.Body.Drop()
	reader.IncomingResponse.Drop()
	reader.Future.Drop()
	reader.OutgoingRequest.Drop()
	return nil
}
