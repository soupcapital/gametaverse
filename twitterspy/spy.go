package twitterspy

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Spy struct {
	spider       *TwitterSpider
	tgbot        *TGBot
	tgbotMsgChan chan string
	opts         options
	msgChan      chan (TweetInfo)
	userReg      *regexp.Regexp
	ctx          context.Context
	db           *mongo.Database
	dbClient     *mongo.Client
	tweetTbl     *mongo.Collection
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

func (s *Spy) storeTweet(tweet *TweetInfo) (err error) {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()
	opt := mngopts.Update()
	opt.SetUpsert(true)

	ts := time.Time(tweet.CreateAt).Unix()
	update := bson.M{
		"$set": bson.M{
			"status": db.TSFound,
			"txt":    tweet.FullText,
			"ts":     ts,
			"vname":  tweet.Author,
			"tid":    strconv.FormatInt(int64(tweet.ID), 10),
		},
	}
	_, err = s.tweetTbl.UpdateByID(ctx, tweet.ID, update, opt)
	if err != nil {
		log.Error("Update vname error: ", err.Error())
		return
	}
	return
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
				s.storeTweet(tweet)
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

	if err = service.initDB(service.opts.MongoURI); err != nil {
		log.Error("DB init error:%s", err.Error())
		return

	}

	service.spider = NewTwitterSpider()
	if err = service.spider.Init(service.msgChan, service.opts.vs, service.opts.twitterInterval, service.opts.twitterCount, service.opts.MongoURI, service.opts.TokenRPC); err != nil {
		log.Error("Twitter init error:%s", err.Error())
		return
	}

	return nil
}

func (s *Spy) initDB(uri string) (err error) {
	s.dbClient, err = mongo.NewClient(mngopts.Client().ApplyURI(uri))
	if err != nil {
		log.Error("new client error: %s", err.Error())
		return
	}
	ctx, _ := context.WithTimeout(s.ctx, 10*time.Second)
	err = s.dbClient.Connect(ctx)
	if err != nil {
		log.Error("connect mongo error:%s", err.Error())
		return
	}

	err = s.dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("ping mongo error:%s", err.Error())
	} else {
		log.Info("connect mongo success")
	}

	s.db = s.dbClient.Database(db.DBName)
	if s.db == nil {
		log.Error("db is null, please init db first")
		return
	}

	s.tweetTbl = s.db.Collection(db.TweetTable)
	return
}

func StartService() (err error) {
	err = service.start()
	return err
}
