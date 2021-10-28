package db

type Guard struct {
	ID   string `json:"_id" bson:"_id"`
	News []int  `json:"news" bson:"news"`
}
