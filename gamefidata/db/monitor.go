package db

const (
	MonitorFieldName = "f_monitor"
	MonitorTableName = "t_monitor"
)

type Monitor struct {
	ID       string `json:"_id" bson:"_id"`
	TopBlock uint64 `json:"topblock" bson:"topblock"`
}
