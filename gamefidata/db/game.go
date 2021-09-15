package db

type Game struct {
	ID       string   `json:"_id" bson:"_id"`
	Name     string   `json:"name" bson:"name"`
	Contract []string `json:"contract" bson:"contract"`
}
