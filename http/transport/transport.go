package transport

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/hayride-dev/wit/gen/go/http/client"
)

// Transport implements http.RoundTripper
type Transport struct {
}

// NewTransport returns http.RoundTripper based on wasi-http
func NewTransport() http.RoundTripper {
	return &Transport{}
}

// RouteTrip implements http.RoundTripper for wasi-http runtimes
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	headers := wasiHeaders(req.Header)
	method := wasiMethod(req.Method)
	scheme := wasiScheme(req.URL.Scheme)
	path_with_query := client.Some(req.URL.RequestURI())
	authority := client.Some(req.URL.Host)

	wasiRequest := client.NewOutgoingRequest(headers)
	wasiRequest.SetMethod(method)
	wasiRequest.SetPathWithQuery(path_with_query)
	wasiRequest.SetScheme(client.Some(scheme))
	wasiRequest.SetAuthority(authority)

	body := wasiRequest.Body().Unwrap()
	defer body.Drop()

	if req.Body != nil {
		stream := body.Write().Unwrap()
		defer stream.Drop()
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		if result := stream.BlockingWriteAndFlush(b); result.IsErr() {
			return nil, errors.New(result.UnwrapErr().GetLastOperationFailed().ToDebugString())
		}
	}
	result := client.StaticOutgoingBodyFinish(body, client.None[client.WasiHttp0_2_0_TypesTrailers]())
	if result.IsErr() {
		return nil, errors.New("failed to finish request body")
	}

	options := client.None[client.WasiHttp0_2_0_TypesRequestOptions]()
	futureResponse := client.WasiHttp0_2_0_OutgoingHandlerHandle(wasiRequest, options).Unwrap()

	poll := futureResponse.Subscribe()
	poll.Block()
	incomingResponse := futureResponse.Get()
	var wasiResponse client.WasiHttp0_2_0_TypesIncomingResponse
	if incomingResponse.IsSome() {
		result2 := incomingResponse.Unwrap()
		if result2.IsErr() {
			return nil, errors.New("failed to get response")
		}
		result3 := result2.Unwrap()
		if result3.IsErr() {
			return nil, errors.New("failed to get response")
		}
		wasiResponse = result3.Unwrap()
	}
	poll.Drop()

	status := int(wasiResponse.Status())
	responseHeaders := wasiResponse.Headers()
	defer responseHeaders.Drop()

	responseHeaderEntries := responseHeaders.Entries()
	header := http.Header{}

	for _, tuple := range responseHeaderEntries {
		ck := http.CanonicalHeaderKey(tuple.F0)
		header[ck] = append(header[ck], string(tuple.F1))
	}

	var contentLength int64
	contentLengthStr := header.Get("Content-Length")
	switch {
	case contentLengthStr != "":
		value, err := strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("malformed content-length header value: %v", err)
		}
		if value < 0 {
			return nil, fmt.Errorf("invalid content-length header value: %q", contentLengthStr)
		}
		contentLength = value
	default:
		contentLength = -1
	}

	responseBodyResult := wasiResponse.Consume()
	if responseBodyResult.IsErr() {
		return nil, errors.New("failed to consume response body")
	}
	responseBody := responseBodyResult.Unwrap()

	responseBodyStreamResult := responseBody.Stream()
	if responseBodyStreamResult.IsErr() {
		return nil, errors.New("failed to get response body stream")
	}
	responseBodyStream := responseBodyStreamResult.Unwrap()

	responseReader := wasiReadCloser{
		Stream:           responseBodyStream,
		Body:             responseBody,
		OutgoingRequest:  wasiRequest,
		IncomingResponse: wasiResponse,
		Future:           futureResponse,
	}

	response := &http.Response{
		Status:        fmt.Sprintf("%d %s", status, http.StatusText(status)),
		StatusCode:    status,
		Header:        header,
		ContentLength: contentLength,
		Body:          responseReader,
		Request:       req,
	}

	return response, nil
}

func wasiMethod(method string) client.WasiHttp0_2_0_TypesMethod {
	switch method {
	case "GET":
		return client.WasiHttp0_2_0_TypesMethodGet()
	case "POST":
		return client.WasiHttp0_2_0_TypesMethodPost()
	case "PUT":
		return client.WasiHttp0_2_0_TypesMethodPut()
	case "DELETE":
		return client.WasiHttp0_2_0_TypesMethodDelete()
	case "PATCH":
		return client.WasiHttp0_2_0_TypesMethodPatch()
	case "HEAD":
		return client.WasiHttp0_2_0_TypesMethodHead()
	case "OPTIONS":
		return client.WasiHttp0_2_0_TypesMethodOptions()
	case "TRACE":
		return client.WasiHttp0_2_0_TypesMethodTrace()
	case "CONNECT":
		return client.WasiHttp0_2_0_TypesMethodConnect()
	default:
		return client.WasiHttp0_2_0_TypesMethodOther(method)
	}
}

func wasiScheme(scheme string) client.WasiHttp0_2_0_TypesScheme {
	switch scheme {
	case "http":
		return client.WasiHttp0_2_0_TypesSchemeHttp()
	case "https":
		return client.WasiHttp0_2_0_TypesSchemeHttps()
	default:
		return client.WasiHttp0_2_0_TypesSchemeOther(scheme)
	}
}

func wasiHeaders(headers map[string][]string) client.WasiHttp0_2_0_TypesFields {
	if headers == nil {
		return client.StaticFieldsFromList([]client.WasiHttp0_2_0_TypesTuple2FieldKeyFieldValueT{}).Unwrap()
	}
	wasiHeaders := []client.WasiHttp0_2_0_TypesTuple2FieldKeyFieldValueT{}
	for k, v := range headers {
		header := client.WasiHttp0_2_0_TypesTuple2FieldKeyFieldValueT{
			F0: k,
			F1: []uint8(v[0]),
		}
		if len(v) > 1 {
			for _, value := range v[1:] {
				header.F1 = append(header.F1, value...)
			}
		}
		wasiHeaders = append(wasiHeaders, header)
	}
	requestHeaders := client.StaticFieldsFromList(wasiHeaders).Unwrap()
	return requestHeaders
}
