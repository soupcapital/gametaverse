package api

type APIError struct {
	ErrNo  int32  `json:"errno"`
	ErrMsg string `json:"errmsg"`
}

var (
	ErrParam = &APIError{ErrNo: 53421, ErrMsg: "Input param is wrong"}
	ErrGame  = &APIError{ErrNo: 53422, ErrMsg: "GameID dose not exist"}
	ErrDB    = &APIError{ErrNo: 53423, ErrMsg: "Database operation error"}
)
