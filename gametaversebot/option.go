package gametaversebot

type options struct {
	tgbotToken string
	robot      string
	groups     []int64
	RPCAddr    string
	MongoURI   string
}

func defaultOptions() options {
	return options{
		tgbotToken: "",
	}
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

func WithTGBotToken(token string) Option {
	return newFuncOption(func(o *options) {
		o.tgbotToken = token
	})
}

func WithRobot(robot string) Option {
	return newFuncOption(func(o *options) {
		o.robot = robot
	})
}

func WithRPCAddr(addr string) Option {
	return newFuncOption(func(o *options) {
		o.RPCAddr = addr
	})
}

func WithGroups(groups []int64) Option {
	return newFuncOption(func(o *options) {
		o.groups = make([]int64, len(groups))
		copy(o.groups, groups)
	})
}

func WithMongoURI(URI string) Option {
	return newFuncOption(func(o *options) {
		o.MongoURI = URI
	})
}
