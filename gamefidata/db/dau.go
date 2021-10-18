package db

type DAU struct {
	GameID    string `json:"game" bson:"game"`
	Timestamp uint64 `json:"ts" bson:"ts"`
	User      string `json:"user" bson:"user"`
}
