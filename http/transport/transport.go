package transport

import (
	"fmt"
	"io"
	"net/http"

	outgoinghandler "github.com/hayride-dev/wit/gen/platform/wasi/http/outgoing-handler"
	"github.com/hayride-dev/wit/gen/platform/wasi/http/types"
	"github.com/ydnar/wasm-tools-go/cm"
)

var _ http.RoundTripper = &RoundTrip{}

type RoundTrip struct {
}

func NewWasiRoundTripper() *RoundTrip {
	return &RoundTrip{}
}

func (r *RoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	// headers
	wasiHeaders := wasiHeader(req.Header)
	// method
	wasiMethod := wasiMethod(req.Method)
	// path
	wasiPath := cm.Some(req.URL.RequestURI())
	// scheme
	wasiScheme := cm.Some(wasiScheme(req.URL.Scheme))
	// authority
	wasiAuthority := cm.Some(req.URL.Host)

	wasiRequest := types.NewOutgoingRequest(wasiHeaders)
	wasiRequest.SetMethod(wasiMethod)
	wasiRequest.SetPathWithQuery(wasiPath)
	wasiRequest.SetScheme(wasiScheme)
	wasiRequest.SetAuthority(wasiAuthority)

	result := outgoinghandler.Handle(wasiRequest, cm.None[types.RequestOptions]())
	if result.IsErr() {
		return nil, fmt.Errorf("error %v", result.Err())
	}

	if result.IsOK() {
		result.OK().Subscribe().Block()
		incomingResponse := result.OK().Get()
		if incomingResponse.Some().IsErr() {
			return nil, fmt.Errorf("error %v", incomingResponse.Some().Err())
		}
		if incomingResponse.Some().OK().IsErr() {
			return nil, fmt.Errorf("error %v", incomingResponse.Some().OK().Err())
		}
		ok := incomingResponse.Some().OK().OK()
		var body io.ReadCloser
		if consume := ok.Consume(); consume.IsErr() {
			return nil, fmt.Errorf("error %v", consume.Err())
		} else if stream := consume.OK().Stream(); stream.IsErr() {
			return nil, fmt.Errorf("error %v", stream.Err())
		} else {
			body = NewReadCloser(*stream.OK())
		}

		response := &http.Response{
			StatusCode:    int(ok.Status()),
			Status:        http.StatusText(int(ok.Status())),
			ContentLength: 0,
			Body:          body,
			Request:       req,
		}

		return response, nil
	}
	return nil, fmt.Errorf("failed to get response")
}

func wasiHeader(headers http.Header) types.Fields {
	fields := types.NewFields()
	for key, values := range headers {
		fieldValues := []types.FieldValue{}
		for _, v := range values {
			fieldValues = append(fieldValues, types.FieldValue(cm.ToList([]uint8(v))))
		}
		fields.Set(types.FieldKey(key), cm.ToList(fieldValues))
	}
	return fields
}

func wasiMethod(method string) types.Method {
	switch method {
	case "GET":
		return types.MethodGet()
	case "POST":
		return types.MethodPost()
	case "PUT":
		return types.MethodPut()
	case "DELETE":
		return types.MethodDelete()
	case "PATCH":
		return types.MethodPatch()
	case "HEAD":
		return types.MethodHead()
	case "OPTIONS":
		return types.MethodOptions()
	case "TRACE":
		return types.MethodTrace()
	case "CONNECT":
		return types.MethodConnect()
	default:
		return types.MethodOther(method)
	}
}

func wasiScheme(scheme string) types.Scheme {
	switch scheme {
	case "http":
		return types.SchemeHTTP()
	case "https":
		return types.SchemeHTTPS()
	default:
		return types.SchemeOther(scheme)
	}
}
