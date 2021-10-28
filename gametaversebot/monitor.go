package gametaversebot

import (
	"context"
	"fmt"

	"github.com/cz-theng/czkit-go/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Monitor struct {
	tgbot        *TGBot
	tgbotMsgChan chan *tgbotapi.Message
	opts         options
	newsChan     chan (*News)
	api          *API
	ctx          context.Context
}

var monitor Monitor

func (m *Monitor) start() (err error) {
	go monitor.tgbot.Run()
	go monitor.handleTGBotMsgLoop()
	go monitor.api.Run()
	for e := range m.newsChan {
		m.dealEvent(e)
	}
	return nil
}

func (m *Monitor) handleTGBotMsgLoop() {
	for msg := range m.tgbotMsgChan {
		m.handleTGBotMsg(msg)
	}
}

func (m *Monitor) handleTGBotMsg(msg *tgbotapi.Message) {
	if msg.NewChatMembers != nil {
		m.handleIncome(msg.NewChatMembers, msg.Chat)
	} else if msg.LeftChatMember != nil {
		m.handleLeft(msg.LeftChatMember, msg.Chat)
	} else if len(msg.Text) != 0 {
		m.handleMsg(msg.Text)
	}
}

func (m *Monitor) handleIncome(users *[]tgbotapi.User, chat *tgbotapi.Chat) {
	for _, user := range *users {
		if user.UserName == m.opts.robot {
			log.Info("Robot join group[%v]:%v", chat.ID, chat.Title)
			//m.tgbot.AddGroup(chat.ID)
		}
	}
}

func (m *Monitor) handleLeft(user *tgbotapi.User, chat *tgbotapi.Chat) {
	if user.UserName == m.opts.robot {
		log.Info("Robot leave group[%v]:%v", chat.ID, chat.Title)
		//m.tgbot.RemoveGroup(chat.ID)
	}
}

func (m *Monitor) handleMsg(text string) {
	m.tgbot.SendMessage(text)
}

func (m *Monitor) dealEvent(news *News) {
	msg := fmt.Sprintf("[%s]:%s \n%s", news.ProjectName, news.Title, news.Content)
	m.tgbot.SendMessage(msg)
}

func Init(opts ...Option) (err error) {
	for _, opt := range opts {
		opt.apply(&monitor.opts)
	}

	monitor.ctx = context.Background()

	monitor.newsChan = make(chan *News)

	monitor.tgbotMsgChan = make(chan *tgbotapi.Message)
	monitor.tgbot = NewTGBot()
	if err = monitor.tgbot.Init(monitor.opts.tgbotToken, monitor.tgbotMsgChan, monitor.opts.groups); err != nil {
		log.Error("TGBot init error:%s", err.Error())
		return
	}

	monitor.api = NewAPI()
	if err = monitor.api.Init(monitor.opts.RPCAddr, monitor.opts.MongoURI, monitor.newsChan); err != nil {
		log.Error("TGBot init error:%s", err.Error())
		return
	}

	return nil
}

func StartService() (err error) {
	err = monitor.start()
	return err
}
