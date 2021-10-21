package db

import "time"

type TweetCoin struct {
	KOL       string    `json:"kol" bson:"kol"`
	Coin      string    `json:"coin" bson:"coin"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	TweetTS   time.Time `json:"tweetts" bson:"tweetts"`
	//ID        string  `json:"_id" bson:"_id"`
}
