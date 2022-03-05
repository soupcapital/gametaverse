package db

const (
	GameTableName = "t_games"
)

type Game struct {
	ID        string   `json:"_id" bson:"_id"`
	Name      string   `json:"name" bson:"name"`
	Chain     string   `json:"chain" bson:"chain"`
	Contracts []string `json:"contracts" bson:"contracts"`
}
