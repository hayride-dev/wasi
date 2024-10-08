package ml

type ModelOptions struct {
	name         string
	maxContext   uint32
	systemPrompt string
}

type OptionType interface {
	*ModelOptions
}

type Option[T OptionType] interface {
	Apply(T) error
}

type funcOption[T OptionType] struct {
	f func(T) error
}

func (fo *funcOption[T]) Apply(opt T) error {
	return fo.f(opt)
}

func newFuncOption[T OptionType](f func(T) error) *funcOption[T] {
	return &funcOption[T]{
		f: f,
	}
}

func WithName(name string) Option[*ModelOptions] {
	return newFuncOption(func(m *ModelOptions) error {
		m.name = name
		return nil
	})
}

func WithMaxContext(maxContext uint32) Option[*ModelOptions] {
	return newFuncOption(func(m *ModelOptions) error {
		m.maxContext = maxContext
		return nil
	})
}

func WithSystemPrompt(text string, data any) Option[*ModelOptions] {
	return newFuncOption(func(m *ModelOptions) error {
		m.systemPrompt = text
		return nil
	})
}

func defaultModelOptions() *ModelOptions {
	return &ModelOptions{
		name:       "todo.ggml",
		maxContext: 1000,
	}
}
