package db

type Digger struct {
	ID            string `json:"_id" bson:"_id"`
	Name          string `json:"name" bson:"name"`
	FollowerCount int    `json:"fc" bson:"fc"`
	TweetsCount   int    `json:"tc" bson:"tc"`
	Timestamp     int64  `json:"ts" bson:"ts"`
}
