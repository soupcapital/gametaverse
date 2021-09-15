package eth

import (
	"sync"

	"github.com/cz-theng/czkit-go/log"
)

type Watcher struct {
	opts  options
	games []*Game
	wg    sync.WaitGroup
}

func New() *Watcher {
	w := &Watcher{}
	return w
}

func (w *Watcher) Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&w.opts)
	}

	for _, info := range w.opts.Games {
		g := NewGame(&info)
		err = g.Init()
		if err != nil {
			log.Error("init game error:%s", err.Error())
			return
		}
		w.games = append(w.games, g)
	}
	return
}

func (w *Watcher) Run() (err error) {
	for _, g := range w.games {
		game := g
		w.wg.Add(1)
		go func() {
			game.Run()
			w.wg.Done()
		}()
	}
	w.wg.Wait()
	return
}
