package cti

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"gitee.com/c_z/cti/db"
	"github.com/cz-theng/czkit-go/log"
	coingecko "github.com/superoo7/go-gecko/v3"
	"github.com/superoo7/go-gecko/v3/types"
	"go.mongodb.org/mongo-driver/mongo"
	mngopts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type CoinTwitterIndex struct {
	spider       *TwitterSpider
	tgbot        *TGBot
	tgbotMsgChan chan string
	opts         options
	msgChan      chan (TweetInfo)
	tokenRexp    *regexp.Regexp
	cgCli        *coingecko.Client
	coinIDs      *types.CoinList

	ctx        context.Context
	dbClient   *mongo.Client
	db         *mongo.Database
	coinWindow []map[string]int
}

var service CoinTwitterIndex

func (ctis *CoinTwitterIndex) start() (err error) {
	go service.tgbot.Run()
	go service.spider.Start()
	go service.refreshCoinID()
	go service.handleTGBotMsgLoop()
	for {
		select {
		case t := <-ctis.msgChan:
			ctis.dealTweet(&t)
		}
	}
}

func (ctis *CoinTwitterIndex) handleTGBotMsgLoop() {
	for {
		select {
		case msg := <-ctis.tgbotMsgChan:
			ctis.handleTGBotMsg(msg)
		}
	}
}

func (ctis *CoinTwitterIndex) handleTGBotMsg(msg string) {
	if strings.HasPrefix(msg, "$") {
		ctis.handleQueryCoinPrice(strings.TrimPrefix(msg, "$"))
		return
	}
	switch msg {
	case "top":
		ctis.handleTopMsg()
	}
}

func (ctis *CoinTwitterIndex) handleQueryCoinPrice(coin string) error {
	var ids []string
	for _, ct := range *ctis.coinIDs {
		if strings.ToUpper(ct.Symbol) == strings.ToUpper(coin) {
			ids = append(ids, ct.ID)
		}
	}
	msg := fmt.Sprintf("%s:", strings.ToUpper(coin))
	for _, id := range ids {
		p, err := ctis.cgCli.SimpleSinglePrice(id, "usd")
		if err != nil {
			log.Error("SimpleSinglePrice error:%s", err.Error())
			continue
		}
		msg = fmt.Sprintf("%s\n\t\t\t\t[%s]:%0.4f", msg, id, p.MarketPrice)
	}
	ctis.tgbot.SendMessage(msg)
	return nil
}

func (ctis *CoinTwitterIndex) handleTopMsg() {
	top := ctis.Top()
	msg := ""
	c := 0
	for _, i := range top {
		msg = fmt.Sprintf("%s\n%s:%d", msg, strings.ToUpper(i.Coin), i.Count)
		c++
		if c > 10 {
			break
		}
	}
	ctis.tgbot.SendMessage(msg)
}

func (ctis *CoinTwitterIndex) refreshCoinID() {
	coinIDs, err := ctis.cgCli.CoinsList()
	if err != nil {
		log.Error("refresh coin IDs error:%s", err.Error())
		return
	}
	ctis.coinIDs = coinIDs
}

func (ctis *CoinTwitterIndex) dealTweet(tweet *TweetInfo) {
	msg := fmt.Sprintf("%s@%s talk about:\n", tweet.Author, time.Time(tweet.CreateAt).Format("2006/01/02 15:04:05"))
	tokens := ctis.tokenRexp.FindAllString(tweet.FullText, -1)
	if nil == tokens || len(tokens) == 0 {
		return
	}
	coins := make(map[string]int)
	for _, t := range tokens {
		c := strings.TrimPrefix(strings.TrimSpace(t), "$")
		if ctis.filterCoins(c) {
			continue
		}
		coins[strings.ToLower(c)]++
	}
	ctis.coinWindow = append(ctis.coinWindow, coins)
	if len(ctis.coinWindow) > 24*6 {
		ctis.coinWindow = ctis.coinWindow[1:]
	}
	ctis.recordCoins(coins, tweet)
	prices, err := ctis.coinPrice(coins)
	if err != nil {
		log.Error("coin error:%s", err.Error())
		return
	}
	if len(prices) == 0 {
		return
	}
	for _, v := range prices {
		msg += fmt.Sprintf("\n%s[%s]:$%f \n24h change[%f%%] \n7d change[%f%%] \nhttps://www.coingecko.com/en/coins/%s \n", v.ID, v.Symbol, v.Price, v.PriceChangePercentage24h, v.PriceChangePercentage7d, v.ID)
		if tokens, err := ctis.uniTokens(v.Symbol); err == nil {
			for _, ID := range tokens {
				msg += fmt.Sprintf("[%s]:Uniswap URL: https://info.uniswap.org/token/%s\n", v.Symbol, ID)
			}
		}
	}
	log.Info(msg)

	ctis.tgbot.SendMessage(msg)
}

func (ctis *CoinTwitterIndex) filterCoins(coin string) bool {
	for _, cc := range ctis.opts.noCoins {
		if strings.ToLower(coin) == strings.ToLower(cc) {
			return true
		}
	}
	return false
}

func (ctis *CoinTwitterIndex) recordCoins(coins map[string]int, tweet *TweetInfo) {
	tweetsTbl := ctis.db.Collection("tweets")
	if tweetsTbl == nil {
		log.Error("collection is null, please init db first")
		return
	}
	var tcCache []interface{}
	now := time.Now()
	for c := range coins {
		tc := db.TweetCoin{
			TweetTS:   time.Time(tweet.CreateAt),
			KOL:       tweet.Author,
			Coin:      strings.ToUpper(c),
			Timestamp: now,
		}
		tcCache = append(tcCache, tc)
	}

	ctx, cancel := context.WithTimeout(ctis.ctx, 3*time.Second)
	defer cancel()
	rst, err := tweetsTbl.InsertMany(ctx, tcCache)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate key error collection") {
			log.Error("insert tweet[%v]  error: %s", tweet, err.Error())
		} else {
			log.Error("duplicate key error collection")
		}
	} else {
		log.Info("insert  tweet :%d", len(rst.InsertedIDs))
	}
}

func (ctis *CoinTwitterIndex) uniTokens(symbol string) (tokenIDs []string, err error) {
	tokens, err := GetTokens(symbol)
	if err != nil {
		return nil, err
	}
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].Liquidity > tokens[j].Liquidity
	})

	max := 0
	for _, t := range tokens {
		log.Info("token[%s]:%f", t.Symbol, t.Liquidity)
		if t.Liquidity < 1 {
			return
		}
		tokenIDs = append(tokenIDs, t.ID)
		max++
		if max >= 3 {
			break
		}
	}
	return
}

func (ctis *CoinTwitterIndex) uniPairs(symbol string) (pairs []string, err error) {
	tokens, err := GetTokens(symbol)
	if err != nil {
		return nil, err
	}
	var ts []string
	for _, t := range tokens {
		ts = append(ts, t.ID)
	}
	ps, err := GetParis(ts)
	if err != nil {
		return nil, err
	}
	for _, p := range ps {
		pairs = append(pairs, p.ID)
	}
	return
}

func (ctis *CoinTwitterIndex) coinPrice(coins map[string]int) (prices []CoinInfo, err error) {
	vc := []string{"usd"}
	log.Info("coins: %v", coins)
	if ctis.coinIDs == nil {
		return nil, fmt.Errorf("coinIDs is null")
	}
	var ids []string
	for c := range coins {
		for _, ct := range *ctis.coinIDs {
			if ct.Symbol == c {
				ids = append(ids, ct.ID)
			}
		}
	}
	sp, err := ctis.cgCli.SimplePrice(ids, vc)
	if err != nil {
		log.Error("SimplePrice error:%s", err.Error())
		return nil, err
	}

	log.Info("prices:%v", sp)
	for k, v := range *sp {
		p := v["usd"]
		price := CoinInfo{
			ID:    k,
			Price: p,
		}
		for _, ct := range *ctis.coinIDs {
			if ct.ID == k {
				price.Symbol = strings.ToUpper(ct.Symbol)
				price.Name = ct.Name
			}
		}
		if coinMarket, err := ctis.cgCli.CoinsID(price.ID, false, false, true, false, false, false); err == nil {
			price.PriceChangePercentage24h = coinMarket.MarketData.PriceChangePercentage24h
			price.PriceChangePercentage7d = coinMarket.MarketData.PriceChangePercentage7d
		}

		prices = append(prices, price)
	}

	return
}

func (ctis *CoinTwitterIndex) initDB() (err error) {
	URI := fmt.Sprintf("mongodb+srv://%s:%s@cluster0.g9w77.mongodb.net/?retryWrites=true&w=majority",
		ctis.opts.dbUser, ctis.opts.dbPasswd)
	ctis.dbClient, err = mongo.NewClient(mngopts.Client().ApplyURI(URI))
	if err != nil {
		log.Error("new client error: %s", err.Error())
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = ctis.dbClient.Connect(ctx)
	if err != nil {
		log.Error("connect mongo error:%s", err.Error())
		return
	}

	err = ctis.dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("ping mongo error:%s", err.Error())
	} else {
		log.Info("connect mongo success")
	}

	ctis.db = ctis.dbClient.Database(ctis.opts.dbName)
	if ctis.db == nil {
		log.Error("db db_cti is null, please init db first")
		return
	}
	return
}

func (ctis *CoinTwitterIndex) Top() []CoinWindowItem {
	coins := make(map[string]int)
	for _, i := range ctis.coinWindow {
		for k, v := range i {
			coins[k] += v
		}
	}

	var top []CoinWindowItem
	for k, v := range coins {
		cwi := CoinWindowItem{
			Coin:  k,
			Count: v,
		}
		top = append(top, cwi)
	}
	sort.Slice(top, func(i, j int) bool {
		return top[i].Count > top[j].Count
	})
	return top
}

func Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&service.opts)
	}

	service.ctx = context.Background()

	service.msgChan = make(chan TweetInfo)
	service.tokenRexp = regexp.MustCompile(`\$[A-Z]+\s+`)

	if err = service.initDB(); err != nil {
		log.Error("Mongon init error:%s", err.Error())
		return
	}

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	service.cgCli = coingecko.NewClient(httpClient)

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
