package utils

const (
	DateFormat  = "2006-01-02"
	SecondofDay = 60 * 60 * 24
)

func StartSecForDay(ts uint64) uint64 {
	return ts / SecondofDay * SecondofDay
}
