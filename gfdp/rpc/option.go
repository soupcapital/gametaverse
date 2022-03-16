package rpc

type options struct {
	DbAddr     string
	DbUser     string
	DbPasswd   string
	DbName     string
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

func WithDbUrl(url string) Option {
	return newFuncOption(func(o *options) {
		o.DbAddr = url
	})
}

func WithDbUser(user string) Option {
	return newFuncOption(func(o *options) {
		o.DbUser = user
	})
}

func WithDbName(name string) Option {
	return newFuncOption(func(o *options) {
		o.DbName = name
	})
}

func WithDbPasswd(passwd string) Option {
	return newFuncOption(func(o *options) {
		o.DbPasswd = passwd
	})
}

func WithListenAddr(addr string) Option {
	return newFuncOption(func(o *options) {
		o.ListenAddr = addr
	})
}
