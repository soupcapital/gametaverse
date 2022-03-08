package db

import "time"

type Transaction struct {
	Timestamp time.Time
	Block     uint64
	From      string
	To        string
}
