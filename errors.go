package uviews

type ApiError struct {
	//Code  int    `json:"code,omitempty"`
	Desc  string `json:"description,omitempty"`
	Debug string `json:"debug,omitempty"`
}

var ApiErrBadRequest = &ApiError{
	Desc: "bad request",
}

/*var ApiErrZeroModTime = &ApiError{
	Desc: "zero modification time",
}*/
