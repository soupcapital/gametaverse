package twitterspy

import "time"

const JSONTimeFormat = time.RubyDate

type JSONTime time.Time

func (t *JSONTime) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 2 {
		*t = JSONTime(time.Time{})
		return
	}

	now, err := time.Parse(`"`+JSONTimeFormat+`"`, string(data))
	*t = JSONTime(now)
	return
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(JSONTimeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, JSONTimeFormat)
	b = append(b, '"')
	return b, nil
}

type TweetInfo struct {
	CreateAt      JSONTime `json:"created_at"`
	ID            uint64   `json:"id"`
	FullText      string   `json:"full_text"`
	Author        string   `json:"author"`
	RetweetCount  int      `json:"retweet_count"`
	FavoriteCount int      `json:"favorite_count"`
	ReplyCount    int      `json:"reply_count"`
	QuoteCount    int      `json:"quote_count"`
}
