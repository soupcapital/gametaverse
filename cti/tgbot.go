package cti

import (
	"time"

	"github.com/cz-theng/czkit-go/log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TGBot struct {
	bot     *tgbotapi.BotAPI
	groups  []int64
	msgChan chan string
}

func NewTGBot() *TGBot {
	bot := &TGBot{}
	return bot
}

func (tgb *TGBot) Init(token string, groups []int64, ch chan string) (err error) {
	tgb.bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Error("NewBotAPI error:%s", err.Error())
		return err
	}
	tgb.msgChan = ch
	tgb.UpdateGroups(groups)

	//tgb.bot.Debug = true
	return nil
}

func (tgb *TGBot) UpdateGroups(groups []int64) {
	_groups := make([]int64, len(groups))
	copy(_groups, groups)
	tgb.groups = _groups
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

		log.Info("Got Message %s", update.Message.Text)
		tgb.msgChan <- update.Message.Text
		// TODO: deal with messages
		// time.Sleep(time.Millisecond * 20)
	}
	return nil
}
