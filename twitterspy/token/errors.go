package token

type APIError struct {
	ErrNo  int32  `json:"errno"`
	ErrMsg string `json:"errmsg"`
}

var (
	ErrParam        = &APIError{ErrNo: 53421, ErrMsg: "Input param is wrong"}
	ErrGame         = &APIError{ErrNo: 53422, ErrMsg: "GameID dose not exist"}
	ErrDB           = &APIError{ErrNo: 53423, ErrMsg: "Database operation error"}
	ErrTimestamp    = &APIError{ErrNo: 53424, ErrMsg: "Time should be timestamp of day on UTC"}
	ErrToken        = &APIError{ErrNo: 53425, ErrMsg: "Refresh twitter token error"}
	ErrUserInfo     = &APIError{ErrNo: 53426, ErrMsg: "Query  twitter user error"}
	ErrBSON         = &APIError{ErrNo: 53427, ErrMsg: "BSON code error"}
	ErrNoDataForDay = &APIError{ErrNo: 53427, ErrMsg: "No more data for that day"}
	ErrQueryTweets  = &APIError{ErrNo: 53428, ErrMsg: "Query tweets error"}
	ErrNoToken      = &APIError{ErrNo: 53429, ErrMsg: "Refresh with no token"}
)
