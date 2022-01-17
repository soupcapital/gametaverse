package db

type TweetStatus int8

const (
	TSFound = TweetStatus(1)
	TSDone  = TweetStatus(2)
)

type Tweet struct {
	TweetID   string `json:"tid" bson:"tid"`
	Status    int8   `json:"status" bson:"status"`
	Txt       string `json:"txt" bson:"txt"`
	Timestamp int64  `json:"ts" bson:"ts"`
	VName     string `json:"vname" bson:"vname"`
}
