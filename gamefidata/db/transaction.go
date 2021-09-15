package db

type Transaction struct {
	GameID    string `json:"gameid" bson:"gameid"`
	Timestamp uint64 `json:"timestamp" bson:"timestamp"`
	Hash      string `json:"hash" bson:"hash"`
	From      string `json:"from" bson:"from"`
	To        string `json:"to" bson:"to"`
}
