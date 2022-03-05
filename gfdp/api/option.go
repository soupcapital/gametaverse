package api

type options struct {
	PrivKey    string
	MongoURI   string
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

func WithMongoURI(URI string) Option {
	return newFuncOption(func(o *options) {
		o.MongoURI = URI
	})
}

func WithListenAddr(addr string) Option {
	return newFuncOption(func(o *options) {
		o.ListenAddr = addr
	})
}
