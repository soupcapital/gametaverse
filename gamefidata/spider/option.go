package spider

type options struct {
	PrivKey     string
	Games       []GameInfo
	MongoURI    string
	Chain       string
	ChainID     int
	RPCAddr     string
	BottomBlock uint64
	Interval    int
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

func WithMongoURI(URI string) Option {
	return newFuncOption(func(o *options) {
		o.MongoURI = URI
	})
}

func WithChain(name string) Option {
	return newFuncOption(func(o *options) {
		o.Chain = name
	})
}

func WithChainID(ID int) Option {
	return newFuncOption(func(o *options) {
		o.ChainID = ID
	})
}

func WithRPCAddr(Addr string) Option {
	return newFuncOption(func(o *options) {
		o.RPCAddr = Addr
	})
}

func WithBottomBlock(block uint64) Option {
	return newFuncOption(func(o *options) {
		o.BottomBlock = block
	})
}

func WithInterval(interval int) Option {
	return newFuncOption(func(o *options) {
		o.Interval = interval
	})
}
