package spider

import (
	"context"
	"sync"

	"github.com/cz-theng/czkit-go/log"
)

type Service struct {
	opts   options
	wg     sync.WaitGroup
	spider *Spider
	ctx    context.Context
	cancel context.CancelFunc
}

func New() *Service {
	s := &Service{}
	return s
}

func (s *Service) Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&s.opts)
	}
	s.ctx = context.Background()

	s.spider = NewSpider(s.opts)
	err = s.spider.Init()
	if err != nil {
		log.Error("Init  spider error:%s", err.Error())
		return err
	}

	return err
}

func (s *Service) routine(ctx context.Context, sp *Spider) {
	s.wg.Add(1)
	sp.Run(ctx, &s.wg)
}

func (s *Service) Run() (err error) {

	ctx, cancel := context.WithCancel(s.ctx)
	s.cancel = cancel
	s.routine(ctx, s.spider)
	s.wg.Wait()
	return
}
