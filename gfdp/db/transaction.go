package db

import "time"

type Transaction struct {
	Hash      string
	Timestamp time.Time
	Block     uint64
	From      string
	To        string
	Data      string
}
