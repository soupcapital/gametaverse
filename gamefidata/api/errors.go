package api

type APIError struct {
	ErrNo  int32  `json:"errno"`
	ErrMsg string `json:"errmsg"`
}

var (
	ErrParam         = &APIError{ErrNo: 53421, ErrMsg: "Input param is wrong"}
	ErrGame          = &APIError{ErrNo: 53422, ErrMsg: "GameID dose not exist"}
	ErrDB            = &APIError{ErrNo: 53423, ErrMsg: "Database operation error"}
	ErrTimestamp     = &APIError{ErrNo: 53424, ErrMsg: "Time should be timestamp of day on UTC"}
	ErrUnknownChain  = &APIError{ErrNo: 53425, ErrMsg: "Unknown chain name"}
	ErrGameExisted   = &APIError{ErrNo: 53426, ErrMsg: "The game is existed please update it "}
	ErrInsertGame    = &APIError{ErrNo: 53427, ErrMsg: "Insert game error"}
	ErrDeleteGame    = &APIError{ErrNo: 53428, ErrMsg: "Delete game error"}
	ErrNoSuchGame    = &APIError{ErrNo: 53429, ErrMsg: "No such game error"}
	ErrUpdateMonitor = &APIError{ErrNo: 53430, ErrMsg: "Update monitor error"}
)
