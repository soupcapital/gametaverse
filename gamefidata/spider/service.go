package spider

import (
	"sync"

	"github.com/cz-theng/czkit-go/log"
)

type Service struct {
	opts     options
	starting chan struct{}
	wg       sync.WaitGroup
	forward  *Spider
	backward *Spider
}

func New() *Service {
	s := &Service{}
	return s
}

func (s *Service) Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&s.opts)
	}

	var games []*Game
	for _, info := range s.opts.Games {
		g := NewGame(info)
		games = append(games, g)
	}

	s.forward = NewSpider(games, s.opts, false)
	err = s.forward.Init()
	if err != nil {
		log.Error("Init forward spider error:%s", err.Error())
		return err
	}

	s.backward = NewSpider(games, s.opts, true)
	err = s.backward.Init()
	if err != nil {
		log.Error("Init backward spider error:%s", err.Error())
		return err
	}

	return err
}

func (s *Service) routine(sp *Spider) {
	s.wg.Add(1)
	go func() {
		sp.Run(s.starting)
		s.wg.Done()
	}()
}

func (s *Service) Run() (err error) {
	s.starting = make(chan struct{})
	s.routine(s.forward)
	s.routine(s.backward)
	s.wg.Wait()
	return
}
