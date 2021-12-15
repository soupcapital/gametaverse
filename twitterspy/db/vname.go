package db

type VNameStatus int8

const (
	VNSTraced  = VNameStatus(1)
	VNSDigged  = VNameStatus(2)
	VNSBlocked = VNameStatus(3)
)

type VName struct {
	ID     string `json:"_id" bson:"_id"`
	Status int8   `json:"status" bson:"status"`
}
