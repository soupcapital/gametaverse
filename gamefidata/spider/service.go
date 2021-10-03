package spider

import (
	"sync"

	"github.com/cz-theng/czkit-go/log"
)

type Service struct {
	opts options

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
		g := NewGame(&info)
		games = append(games, g)
	}

	s.forward = &Spider{
		games:          games,
		bottomBlock:    s.opts.BottomBlock,
		rpcAddr:        s.opts.RPCAddr,
		mongoURI:       s.opts.MongoURI,
		backward:       false,
		interval:       s.opts.Interval,
		chainID:        s.opts.ChainID,
		backwardFactor: s.opts.BackwardFactor,
	}
	err = s.forward.Init()
	if err != nil {
		log.Error("Init forward spider error:%s", err.Error())
		return err
	}

	s.backward = &Spider{
		games:          games,
		bottomBlock:    s.opts.BottomBlock,
		rpcAddr:        s.opts.RPCAddr,
		mongoURI:       s.opts.MongoURI,
		backward:       true,
		interval:       s.opts.Interval,
		chainID:        s.opts.ChainID,
		backwardFactor: s.opts.BackwardFactor,
		chain:          s.opts.Chain,
	}
	err = s.backward.Init()
	if err != nil {
		log.Error("Init backward spider error:%s", err.Error())
		return err
	}

	return
}

func (s *Service) routine(sp *Spider) {
	s.wg.Add(1)
	go func() {
		sp.Run()
		s.wg.Done()
	}()
}

func (s *Service) Run() (err error) {

	s.routine(s.forward)
	s.routine(s.backward)
	s.wg.Wait()
	return
}
