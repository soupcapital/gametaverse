package eth

type options struct {
	PrivKey string
	Games   []GameInfo
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

func WithPrivKey(key string) Option {
	return newFuncOption(func(o *options) {
		o.PrivKey = key
	})
}

func WithGames(games []GameInfo) Option {
	return newFuncOption(func(o *options) {
		o.Games = games
	})
}
