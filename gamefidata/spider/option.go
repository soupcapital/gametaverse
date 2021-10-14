package spider

type options struct {
	PrivKey          string
	Games            []GameInfo
	MongoURI         string
	Chain            string
	ChainID          int
	RPCAddr          string
	BottomBlock      uint64
	ForwardInterval  float32
	BackwardInterval float32
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
		gs := make([]GameInfo, len(games))
		copy(gs, games)
		o.Games = gs
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

func WithBackwardInterval(interval float32) Option {
	return newFuncOption(func(o *options) {
		o.BackwardInterval = interval
	})
}

func WithBottomBlock(block uint64) Option {
	return newFuncOption(func(o *options) {
		o.BottomBlock = block
	})
}

func WithForwardInterval(interval float32) Option {
	return newFuncOption(func(o *options) {
		o.ForwardInterval = interval
	})
}
