package twitterspy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cz-theng/czkit-go/log"
)

type Spy struct {
	spider       *TwitterSpider
	tgbot        *TGBot
	tgbotMsgChan chan string
	opts         options
	msgChan      chan (TweetInfo)

	ctx context.Context
}

var service Spy

func (s *Spy) start() (err error) {
	go service.tgbot.Run()
	go service.spider.Start()
	go service.handleTGBotMsgLoop()
	for {
		select {
		case t := <-s.msgChan:
			s.dealTweet(&t)
		}
	}
}

func (s *Spy) handleTGBotMsgLoop() {
	for {
		select {
		case msg := <-s.tgbotMsgChan:
			s.handleTGBotMsg(msg)
		}
	}
}

func (s *Spy) handleTGBotMsg(msg string) {

	switch msg {
	case "test":
		s.handleTestMsg()
	}
}

func (s *Spy) handleTestMsg() {
	msg := "test"
	s.tgbot.SendMessage(msg)
}

func (s *Spy) dealTweet(tweet *TweetInfo) {
	// {
	// 	// TODO: for tset
	// 	msg := fmt.Sprintf("%s@%s talk about:\n %s", tweet.Author, time.Time(tweet.CreateAt).Format("2006/01/02 15:04:05"), tweet.FullText)
	// 	s.tgbot.SendMessage(msg)
	// }
	txt := strings.ToLower(tweet.FullText)
	for _, word := range s.opts.keyWords {
		word = " " + word + " "
		if strings.Contains(txt, strings.ToLower(word)) {
			msg := fmt.Sprintf("%s@%s talk about:\n %s", tweet.Author, time.Time(tweet.CreateAt).Format("2006/01/02 15:04:05"), tweet.FullText)
			s.tgbot.SendMessage(msg)
		}
	}
}

func Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&service.opts)
	}

	service.ctx = context.Background()

	service.msgChan = make(chan TweetInfo)

	service.tgbotMsgChan = make(chan string)
	service.tgbot = NewTGBot()
	if err = service.tgbot.Init(service.opts.tgbotToken, service.opts.groups, service.tgbotMsgChan); err != nil {
		log.Error("TGBot init error:%s", err.Error())
		return
	}

	service.spider = NewTwitterSpider()
	if err = service.spider.Init(service.msgChan, service.opts.vs, service.opts.twitterInterval, service.opts.twitterCount); err != nil {
		log.Error("Twitter init error:%s", err.Error())
		return
	}

	return nil
}

func StartService() (err error) {
	err = service.start()
	return err
}
