package twitterspy

import (
	"context"
	"fmt"
	"regexp"
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
	userReg      *regexp.Regexp
	ctx          context.Context
}

var service Spy

func (s *Spy) start() (err error) {
	s.userReg = regexp.MustCompile(`(^|[^@\w])@(\w{1,15})\b`)
	if s.userReg == nil {
		log.Error("regexp err:%s", err.Error())
		return
	}
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

func (s *Spy) digUser(tweet *TweetInfo) {
	rst := s.userReg.FindAllStringSubmatch(tweet.FullText, -1)
	for _, m := range rst {
		if len(m) > 0 {
			user := m[0]
			user = strings.TrimSpace(user)
			if !strings.HasPrefix(user, "@") {
				continue
			}
			user = strings.TrimLeft(user, "@")
			log.Info("Add a dig user:%v", user)
			s.spider.UpdateDigUser(user)
		}
	}
}

func (s *Spy) dealTweet(tweet *TweetInfo) {
	{
		// TODO: for tset
		msg := fmt.Sprintf("%s@%s talk about:\n %s", tweet.Author, time.Time(tweet.CreateAt).Format("2006/01/02 15:04:05"), tweet.FullText)
		log.Info("deal tweet:%v", msg)
	}

	s.digUser(tweet)
	txt := strings.ToLower(tweet.FullText)
	for _, word := range s.opts.keyWords {
		word = strings.ToLower(word)
		words := []string{" " + word + " ",
			" " + word + "\n",
			"\n" + word + " ",
			"\n" + word + "\n",
			"#" + word + " ",
			"#" + word + "\n"}
		for _, w := range words {
			if strings.Contains(txt, w) {
				msg := fmt.Sprintf("%s@%s talk about:\n %s", tweet.Author, time.Time(tweet.CreateAt).Format("2006/01/02 15:04:05"), tweet.FullText)
				s.tgbot.SendMessage(msg)
				log.Info("[SEND]%s", msg)
				return
			}
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
	if err = service.spider.Init(service.msgChan, service.opts.vs, service.opts.twitterInterval, service.opts.twitterCount, service.opts.MongoURI); err != nil {
		log.Error("Twitter init error:%s", err.Error())
		return
	}

	return nil
}

func StartService() (err error) {
	err = service.start()
	return err
}
