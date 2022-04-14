package db

const (
	HauTableName = "t_hau"
)

type Hau struct {
	ID        string `json:"_id" bson:"_id"`
	Timestamp uint64 `json:"ts" bson:"ts"`
	Chain     string `json:"chain" bson:"chain"`
	Hau       int64  `json:"hau" bson:"hau"`
}
