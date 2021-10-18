package spider

import (
	"sync"

	"github.com/cz-theng/czkit-go/log"
)

type Service struct {
	opts options

	wg       sync.WaitGroup
	forward  Spider
	backward Spider
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

	switch s.opts.Chain {
	case "polygon", "eth", "bsc":
		err = s.initETHSpider(games)
	case "wax":
		err = s.initEOSSpider(games)
	}

	return err
}

func (s *Service) initEOSSpider(games []*Game) (err error) {
	s.forward = &EOSSpider{
		games:            games,
		bottomBlock:      uint32(s.opts.BottomBlock),
		rpcAddr:          s.opts.RPCAddr,
		mongoURI:         s.opts.MongoURI,
		backward:         false,
		forwardInterval:  s.opts.ForwardInterval,
		backwardInterval: s.opts.BackwardInterval,
		chainID:          s.opts.ChainID,
		chain:            s.opts.Chain,
		forwardWorks:     s.opts.ForwardWorks,
		backwardWorks:    s.opts.BackwardWorks,
	}
	err = s.forward.Init()
	if err != nil {
		log.Error("Init forward spider error:%s", err.Error())
		return err
	}

	s.backward = &EOSSpider{
		games:            games,
		bottomBlock:      uint32(s.opts.BottomBlock),
		rpcAddr:          s.opts.RPCAddr,
		mongoURI:         s.opts.MongoURI,
		backward:         true,
		forwardInterval:  s.opts.ForwardInterval,
		backwardInterval: s.opts.BackwardInterval,
		chainID:          s.opts.ChainID,
		chain:            s.opts.Chain,
	}
	err = s.backward.Init()
	if err != nil {
		log.Error("Init backward spider error:%s", err.Error())
		return err
	}

	return
}

func (s *Service) initETHSpider(games []*Game) (err error) {
	s.forward = &ETHSpider{
		games:            games,
		bottomBlock:      s.opts.BottomBlock,
		rpcAddr:          s.opts.RPCAddr,
		mongoURI:         s.opts.MongoURI,
		backward:         false,
		forwardInterval:  s.opts.ForwardInterval,
		backwardInterval: s.opts.BackwardInterval,
		chainID:          s.opts.ChainID,
		chain:            s.opts.Chain,
	}
	err = s.forward.Init()
	if err != nil {
		log.Error("Init forward spider error:%s", err.Error())
		return err
	}

	s.backward = &ETHSpider{
		games:            games,
		bottomBlock:      s.opts.BottomBlock,
		rpcAddr:          s.opts.RPCAddr,
		mongoURI:         s.opts.MongoURI,
		backward:         true,
		forwardInterval:  s.opts.ForwardInterval,
		backwardInterval: s.opts.BackwardInterval,
		chainID:          s.opts.ChainID,
		chain:            s.opts.Chain,
	}
	err = s.backward.Init()
	if err != nil {
		log.Error("Init backward spider error:%s", err.Error())
		return err
	}

	return
}

func (s *Service) routine(sp Spider) {
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
