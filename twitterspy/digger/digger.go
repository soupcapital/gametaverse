package digger

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy"
)

type Digger struct {
	spider  *twitterspy.TwitterSpider
	msgChan chan (twitterspy.TweetInfo)
	ctx     context.Context
	vs      []string
	count   uint32
}

var _digger Digger

func Init(addr string, count uint32) (err error) {
	_digger.ctx = context.Background()
	_digger.count = count
	_digger.msgChan = make(chan twitterspy.TweetInfo)
	_digger.spider = twitterspy.NewTwitterSpider()
	if err = _digger.updateVs(addr); err != nil {
		log.Error("update vs error:%s", err.Error())
		return
	}
	if err = _digger.spider.Init(_digger.msgChan, _digger.vs, 0, 0, ""); err != nil {
		log.Error("Twitter init error:%s", err.Error())
		return
	}
	return
}

func Start() {
	done := make(chan (struct{}))
	go _digger.spider.Digger(done, _digger.vs, _digger.count)
	for {
		select {
		case t := <-_digger.msgChan:
			_digger.dealTweet(&t)
		case <-done:
			return
		}
	}
}

func (dg *Digger) updateVs(addr string) (err error) {
	resp, err := http.Get(addr)
	if err != nil {
		log.Error("get vsname err:%s", err.Error())
		return
	}
	respJOSN := struct {
		Vanmes []string `json:"vanmes"`
		Errno  int      `json:"errno"`
		Errmsg string   `json:"errmsg"`
	}{}
	bodyDecoder := json.NewDecoder(resp.Body)
	if err = bodyDecoder.Decode(&respJOSN); err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	if respJOSN.Errno != 0 {
		log.Error("resp is :%v", respJOSN)
		return errors.New(respJOSN.Errmsg)
	}
	log.Info("resp is :%v", respJOSN.Vanmes)
	dg.vs = make([]string, len(respJOSN.Vanmes))
	copy(dg.vs, respJOSN.Vanmes[:])
	log.Info("db vs %v", dg.vs)
	return
}

func (dg *Digger) dealTweet(tweet *twitterspy.TweetInfo) {
	txt := strings.ToLower(tweet.FullText)
	log.Info("txt:%s", txt)
}
