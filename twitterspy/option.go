package twitterspy

import "time"

type options struct {
	vs              []string
	groups          []int64
	tgbotToken      string
	twitterInterval time.Duration
	twitterCount    uint32
	keyWords        []string
}

func defaultOptions() options {
	return options{
		vs:         []string{},
		groups:     []int64{},
		tgbotToken: "",
	}
}

type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (fdo *funcOption) apply(do *options) {
	fdo.f(do)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func WithGroups(groups []int64) Option {
	return newFuncOption(func(o *options) {
		_groups := make([]int64, len(groups))
		copy(_groups, groups)
		o.groups = _groups
	})
}

func WithVs(vs []string) Option {
	return newFuncOption(func(o *options) {
		_vs := make([]string, len(vs))
		copy(_vs, vs)
		o.vs = _vs
	})
}

func WithTGBotToken(token string) Option {
	return newFuncOption(func(o *options) {
		o.tgbotToken = token
	})
}

func WithTwitterInternal(internal uint32) Option {
	return newFuncOption(func(o *options) {
		o.twitterInterval = time.Duration(internal) * time.Second
	})
}

func WithTwitterCount(count uint32) Option {
	return newFuncOption(func(o *options) {
		o.twitterCount = count
	})
}

func WithKeyWords(words []string) Option {
	return newFuncOption(func(o *options) {
		o.keyWords = make([]string, len(words))
		copy(o.keyWords, words)
	})
}
