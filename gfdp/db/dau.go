package db

type DAU struct {
	ID        uint64 `json:"_id" bson:"_id"`
	GameID    string `json:"game" bson:"game"`
	Timestamp uint64 `json:"ts" bson:"ts"`
	User      string `json:"user" bson:"user"`
	Chain     string `json:"chain" bson:"chain"`
}
