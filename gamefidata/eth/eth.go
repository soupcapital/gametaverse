package eth

import "context"

type Game struct {
	ctx  context.Context
	opts options
}

func New() *Game {
	gm := &Game{}
	return gm
}

func (gm *Game) Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&gm.opts)
	}

	return
}

func (gm *Game) Run() (err error) {
	return
}
