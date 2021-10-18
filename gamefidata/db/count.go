package db

type Count struct {
	GameID    string `json:"game" bson:"game"`
	Timestamp uint64 `json:"ts" bson:"ts"`
	Count     uint64 `json:"count" bson:"count"`
}
