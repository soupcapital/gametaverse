package db

const (
	//GameTableName     = "t_games"
	GameTableName     = "t_games_info"
	GameInfoTableName = "t_games_info"
)

type Game struct {
	ID        string   `json:"_id" bson:"_id"`
	GameID    string   `json:"id" bson:"id"`
	Name      string   `json:"name" bson:"name"`
	Chain     string   `json:"chain" bson:"chain"`
	Contracts []string `json:"contracts" bson:"contracts"`
}
