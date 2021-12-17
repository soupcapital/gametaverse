package api

type APIError struct {
	ErrNo  int32  `json:"errno"`
	ErrMsg string `json:"errmsg"`
}

var (
	ErrParam     = &APIError{ErrNo: 53421, ErrMsg: "Input param is wrong"}
	ErrGame      = &APIError{ErrNo: 53422, ErrMsg: "GameID dose not exist"}
	ErrDB        = &APIError{ErrNo: 53423, ErrMsg: "Database operation error"}
	ErrTimestamp = &APIError{ErrNo: 53424, ErrMsg: "Time should be timestamp of day on UTC"}
	ErrToken     = &APIError{ErrNo: 53425, ErrMsg: "Refresh twitter token error"}
	ErrUserInfo  = &APIError{ErrNo: 53426, ErrMsg: "Query  twitter user error"}
)
