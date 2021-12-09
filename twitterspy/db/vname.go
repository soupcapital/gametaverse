package db

type VName struct {
	ID     string `json:"_id" bson:"_id"`
	Status int8   `json:"status" bson:"status"`
}
