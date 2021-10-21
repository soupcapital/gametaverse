package twitterspy

import (
	"fmt"
	"time"

	"github.com/cz-theng/czkit-go/log"
)

type TwitterSpider struct {
	tgbot    *TGBot
	token    *Token
	vs       []string
	conn     *TwitterSearchConn
	internal time.Duration
	perCount uint32
	msgChan  chan (TweetInfo)
}

func NewTwitterSpider() *TwitterSpider {
	ts := &TwitterSpider{}
	ts.token = NewToken()
	ts.conn = NewTwitterSearchConn()
	return ts
}

func (ts *TwitterSpider) Init(msgChan chan (TweetInfo), vs []string, internal time.Duration, count uint32) (err error) {
	_vs := make([]string, len(vs))
	copy(_vs, vs)
	ts.vs = _vs

	if err = ts.token.Refresh(); err != nil {
		return err
	}
	if err = ts.conn.Init(ts.token.token); err != nil {
		return err
	}
	ts.internal = internal
	ts.perCount = count
	ts.msgChan = msgChan
	return nil
}

func (ts *TwitterSpider) Start() (err error) {
	ticker := time.NewTicker(ts.internal)
	ts.updateTwitter()
	for {
		select {
		case <-ticker.C:
			ts.updateTwitter()
		}
	}
}

func (ts *TwitterSpider) updateTwitter() {
	for _, v := range ts.vs {
	AGAIN:
		tweets, err := ts.conn.QueryV(v, ts.internal, ts.perCount)
		if err != nil {
			if err == ErrTokenForbid {
				if err = ts.token.Refresh(); err == nil {
					ts.conn.token = ts.token.token
					log.Info("Refresh token success and goto Again")
					goto AGAIN
				}
			}
			log.Error("QueryV error:%s", err.Error())
			continue
		}
		log.Info("Query %v Got :%v", v, tweets)
		for _, t := range tweets {
			t.Author = v
			msg := fmt.Sprintf("[%s@%s]:%s", v, time.Time(t.CreateAt).String(), t.FullText)
			log.Info("TWEET:%s", msg)
			ts.msgChan <- t
		}
	}
}
