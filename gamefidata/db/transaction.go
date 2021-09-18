package db

type Transaction struct {
	GameID    string `json:"gameid" bson:"gameid"`
	Timestamp uint64 `json:"timestamp" bson:"timestamp"`
	ID        string `json:"_id" bson:"_id"`
	From      string `json:"from" bson:"from"`
	To        string `json:"to" bson:"to"`
	BlockNum  uint64 `json:"blocknum" bson:"blocknum"`
}
