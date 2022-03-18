package uviews

type ApiError struct {
	//Code  int    `json:"code,omitempty"`
	Desc  string `json:"description,omitempty"`
	Debug string `json:"debug,omitempty"`
}

var ApiErrWrongAncestors = &ApiError{
	Desc: "wrong number of ancestors",
}
