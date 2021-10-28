package gametaversebot

import (
	"time"

	"github.com/cz-theng/czkit-go/log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TGBot struct {
	bot     *tgbotapi.BotAPI
	groups  []int64
	msgChan chan *tgbotapi.Message
}

func NewTGBot() *TGBot {
	bot := &TGBot{}
	return bot
}

func (tgb *TGBot) Init(token string, ch chan *tgbotapi.Message, groups []int64) (err error) {
	tgb.bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Error("NewBotAPI error:%s", err.Error())
		return err
	}
	tgb.msgChan = ch
	tgb.groups = make([]int64, len(groups))
	copy(tgb.groups, groups)
	//tgb.bot.Debug = true
	return nil
}

func (tgb *TGBot) AddGroup(group int64) {
	for _, g := range tgb.groups {
		if g == group {
			return
		}
	}
	tgb.groups = append(tgb.groups, group)
}

func (tgb *TGBot) RemoveGroup(group int64) {
	var groups []int64
	for _, g := range tgb.groups {
		if g != group {
			groups = append(groups, g)
		}
	}
	tgb.groups = groups
}

func (tgb *TGBot) SendMessage(txt string) (err error) {
	if tgb.groups == nil || len(tgb.groups) == 0 {
		return nil
	}

	for _, group := range tgb.groups {
		msg := tgbotapi.NewMessage(group, txt)
		tgb.bot.Send(msg)
	}

	return nil
}

func (tgb *TGBot) Run() (err error) {

	u := tgbotapi.NewUpdate(0)

	updates, err := tgb.bot.GetUpdatesChan(u)
	if err != nil {
		log.Error("NewBotAPI error:%s", err.Error())
		return err
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			time.Sleep(time.Millisecond * 20)
			continue
		}

		//log.Info("Got Message %v %v %v", update.Message, update.Message.Chat, update.Message.LeftChatMember)
		tgb.msgChan <- update.Message
	}
	return nil
}
