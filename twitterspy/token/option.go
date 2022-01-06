package token

type options struct {
	ListenAddr string
}

func defaultOptions() options {
	return options{}
}

type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (fdo *funcOption) apply(do *options) {
	fdo.f(do)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func WithListenAddr(addr string) Option {
	return newFuncOption(func(o *options) {
		o.ListenAddr = addr
	})
}
