package db

type Digger struct {
	ID            string  `json:"_id" bson:"_id"`
	Name          string  `json:"name" bson:"name"`
	FollowerCount int     `json:"fc" bson:"fc"`
	TweetsCount   int     `json:"tc" bson:"tc"`
	FavoriteCount int     `json:"ftc" bson:"ftc"`
	ReplyCount    int     `json:"rpc" bson:"rpc"`
	RetweetCount  int     `json:"rtc" bson:"rtc"`
	Timestamp     int64   `json:"ts" bson:"ts"`
	Score         float32 `json:"score" bson:"score"`
}
