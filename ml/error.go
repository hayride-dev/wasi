package ml

import "github.com/hayride-dev/bindgen/gen/go/wasi/nn/errors"

type mlErr struct {
	e *errors.Error
}

func (err *mlErr) Error() string {
	return err.e.Code().String()
}
