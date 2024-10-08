package ml

import (
	"context"

	"github.com/bytecodealliance/wasm-tools-go/cm"
	"github.com/hayride-dev/bindgen/gen/go/wasi/nn/graph"
	"github.com/hayride-dev/bindgen/gen/go/wasi/nn/inference"
	"github.com/hayride-dev/bindgen/gen/go/wasi/nn/tensor"
)

type Model struct {
	graphExecCtx *inference.GraphExecutionContext
	inputTensor  *tensor.Tensor
}

func New(ctx context.Context, options ...Option[*ModelOptions]) (*Model, error) {
	opts := defaultModelOptions()
	for _, opt := range options {
		if err := opt.Apply(opts); err != nil {
			return nil, err
		}
	}

	graphResult := graph.LoadByName(opts.name)
	if graphResult.IsErr() {
		e := &mlErr{graphResult.Err()}
		return nil, e
	}

	graph := graphResult.OK()
	execCtxResult := graph.InitExecutionContext()
	if execCtxResult.IsErr() {
		return nil, &mlErr{execCtxResult.Err()}
	}
	execCtx := execCtxResult.OK()

	return &Model{graphExecCtx: execCtx}, nil
}

func (m *Model) Input(text string, data any) error {
	d := tensor.TensorDimensions(cm.ToList([]uint32{1}))
	td := tensor.TensorData(cm.ToList([]uint8(text)))
	t := tensor.NewTensor(d, tensor.TensorTypeU8, td)
	inputResult := m.graphExecCtx.SetInput("prompt", t)
	if inputResult.IsErr() {
		return &mlErr{inputResult.Err()}
	}
	m.inputTensor = &t
	return nil
}

func (m *Model) Output() (string, error) {
	outputResult := m.graphExecCtx.GetOutput("prompt")
	if outputResult.IsErr() {
		return "", &mlErr{outputResult.Err()}
	}
	tensor := outputResult.OK()
	return string(tensor.Data().Slice()), nil
}

func (m *Model) Compute(ctx context.Context) error {
	computeResult := m.graphExecCtx.Compute()
	if computeResult.IsErr() {
		return &mlErr{computeResult.Err()}
	}
	return nil
}

func (m *Model) ResetCtx() error {
	return nil
}

func (m *Model) ExtendCtx(text string) error {
	return nil
}
