package db

import "time"

type Transaction struct {
	Timestamp time.Time
	Block     uint64
	Index     uint16
	From      string
	To        string
}
